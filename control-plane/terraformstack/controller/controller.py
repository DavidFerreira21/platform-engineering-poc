"""Reconcilia a solicitação de demonstração da Etapa 06."""

import logging
import os
from datetime import datetime, timezone
from threading import Event, Thread
from time import sleep

from kubernetes import client, config, watch
from kubernetes.client.exceptions import ApiException
from urllib3.exceptions import HTTPError

from leadership import Leadership

STACK = ("platform.example.io", "v1alpha1", "platform-system", "terraformstacks")
WORKFLOW = ("argoproj.io", "v1alpha1", "argo-workflows", "workflows")
STACK_NAME = "lab06-network"
TEMPLATE_NAME = "terraformstack-plan-simulado"
CONDITION = "SimulatedPlanReady"
UID_LABEL = "lab.platform.example/stack-uid"
GENERATION_LABEL = "lab.platform.example/stack-generation"
EXERCISE_LABEL = "lab.platform.example/exercise"
TIMEOUT = 10


def update_condition(api, stack, state, reason, message, check_leadership):
    metadata = stack["metadata"]
    conditions = stack.get("status", {}).get("conditions", [])
    previous = next((condition for condition in conditions if condition["type"] == CONDITION), {})
    transition = previous.get("lastTransitionTime")
    if state != previous.get("status") or not transition:
        transition = datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")
    condition = {
        "type": CONDITION,
        "status": state,
        "reason": reason,
        "message": message,
        "observedGeneration": metadata["generation"],
        "lastTransitionTime": transition,
    }
    if condition == previous:
        return

    # O resourceVersion impede que uma resposta atrasada atualize outra geração.
    check_leadership()
    api.patch_namespaced_custom_object_status(
        *STACK,
        STACK_NAME,
        {
            "metadata": {"resourceVersion": metadata["resourceVersion"]},
            "status": {"conditions": [item for item in conditions if item["type"] != CONDITION] + [condition]},
        },
        _request_timeout=TIMEOUT,
    )


def workflow_for(stack):
    """Cria somente o Workflow simulado; source/version não são resolvidos neste lab."""
    metadata = stack["metadata"]
    spec = stack["spec"]
    uid, generation = metadata["uid"], metadata["generation"]
    source = spec.get("sourceRef")
    legacy = spec.get("module")
    if bool(source) == bool(legacy):
        raise ValueError("Informar exatamente um de spec.sourceRef ou spec.module")
    if source:
        source_name = source["name"]
        path = source["path"]
        parts = path.split("/")
        if any(part in {"", ".", ".."} for part in parts) or path.startswith("/"):
            raise ValueError("sourceRef.path deve ser um caminho relativo permitido")
        module = parts[-1]
        version = source.get("version") or "latest"
    else:
        source_name = "legacy-module"
        path = legacy["name"]
        module = legacy["name"]
        version = legacy["version"]
    config_revision = spec.get("configuration", {}).get("revision") or "latest"
    return {
        "apiVersion": "argoproj.io/v1alpha1",
        "kind": "Workflow",
        "metadata": {
            "name": f"lab06-plan-{uid}-{generation}",
            "namespace": WORKFLOW[2],
            "labels": {
                EXERCISE_LABEL: "lab06",
                UID_LABEL: uid,
                GENERATION_LABEL: str(generation),
            },
        },
        "spec": {
            "workflowTemplateRef": {"name": TEMPLATE_NAME},
            "arguments": {"parameters": [
                {"name": "source", "value": source_name},
                {"name": "path", "value": path},
                {"name": "module", "value": module},
                {"name": "version", "value": version},
                {"name": "config-revision", "value": config_revision},
                {"name": "entity", "value": spec["target"]["entity"]},
                {"name": "environment", "value": spec["target"]["environment"]},
            ]},
        },
    }


def reconcile(api, check_leadership):
    try:
        stack = api.get_namespaced_custom_object(*STACK, STACK_NAME, _request_timeout=TIMEOUT)
    except ApiException as error:
        if error.status == 404:
            return
        raise
    if stack["metadata"].get("deletionTimestamp"):
        # Lab06 não possui finalizer nem executor de destroy; nunca executar teardown.
        return

    spec = stack["spec"]
    if spec.get("lifecycle", {}).get("destroy"):
        update_condition(api, stack, "False", "DestroyNotImplemented",
                         "Destroy exige plan, policy e aprovação; indisponível no simulador.", check_leadership)
        return
    if spec.get("lifecycle", {}).get("destroyResourcesOnDeletion"):
        update_condition(api, stack, "False", "DeletionGuardNotImplemented",
                         "Não excluir a stack com essa flag: finalizer e destroy não foram implementados.", check_leadership)
        return
    if spec.get("suspend", False):
        update_condition(api, stack, "Unknown", "Suspended",
                         "Novas execuções simuladas estão suspensas.", check_leadership)
        return
    if spec.get("dependsOn"):
        update_condition(api, stack, "Unknown", "DependenciesNotImplemented",
                         "dependsOn será resolvido na implementação do controller; sem Workflow simulado.", check_leadership)
        return

    try:
        desired = workflow_for(stack)
    except (ValueError, KeyError) as error:
        update_condition(api, stack, "False", "InvalidStackSource", str(error), check_leadership)
        return
    name = desired["metadata"]["name"]
    try:
        workflow = api.get_namespaced_custom_object(*WORKFLOW, name, _request_timeout=TIMEOUT)
    except ApiException as error:
        if error.status != 404:
            raise
        generation = stack["metadata"]["generation"]
        previously_observed = any(
            item["type"] == CONDITION and item.get("observedGeneration") == generation
            for item in stack.get("status", {}).get("conditions", [])
        )
        if previously_observed:
            update_condition(api, stack, "False", "WorkflowMissing", f"Workflow {name} removido.", check_leadership)
            return
        # Se outra instância criou o Workflow entre GET e POST, o próximo ciclo o encontra.
        check_leadership()
        workflow = api.create_namespaced_custom_object(*WORKFLOW, desired, _request_timeout=TIMEOUT)
        logging.info("Workflow %s criado em %s", name, WORKFLOW[2])

    labels = workflow["metadata"].get("labels", {})
    expected_labels = desired["metadata"]["labels"]
    if any(labels.get(key) != value for key, value in expected_labels.items()) or (
        workflow["spec"].get("workflowTemplateRef") != desired["spec"]["workflowTemplateRef"]
    ):
        update_condition(
            api, stack, "False", "WorkflowMismatch", f"Workflow {name} sem a correlação esperada.", check_leadership
        )
        return

    phase = workflow.get("status", {}).get("phase") or "Pending"
    state, reason = {
        "Succeeded": ("True", "WorkflowSucceeded"),
        "Failed": ("False", "WorkflowFailed"),
        "Error": ("False", "WorkflowError"),
    }.get(phase, ("Unknown", "WorkflowInProgress"))
    update_condition(api, stack, state, reason, f"Workflow {name} em fase {phase}.", check_leadership)


def interval(name, default):
    value = int(os.getenv(name, default))
    if value <= 0:
        raise ValueError(f"{name} precisa ser um inteiro positivo em segundos")
    return value


def stack_revision(stack):
    metadata = stack["metadata"]
    return metadata["uid"], metadata["generation"], metadata.get("deletionTimestamp")


def observe(api, resource, selector, wake, timeout, retry, revision=None):
    list_objects = api.list_namespaced_custom_object
    while True:
        stream = watch.Watch()
        try:
            snapshot = list_objects(*resource, **selector, _request_timeout=TIMEOUT)
            items = snapshot["items"]
            previous = revision(items[0]) if revision and items else None
            # Reconcilia o estado da lista e retoma a partir da mesma revisão.
            wake.set()
            for event in stream.stream(
                list_objects,
                *resource,
                **selector,
                resource_version=snapshot["metadata"]["resourceVersion"],
                timeout_seconds=timeout,
                _request_timeout=(TIMEOUT, timeout + TIMEOUT),
            ):
                if event["type"] == "BOOKMARK":
                    continue
                if revision:
                    current = None if event["type"] == "DELETED" else revision(event["object"])
                    if current == previous:
                        continue
                    previous = current
                wake.set()
        except (ApiException, HTTPError) as error:
            logging.warning("Watch de %s será restabelecido: %s", resource[-1],
                            getattr(error, "status", type(error).__name__))
            sleep(retry)
        finally:
            stream.stop()


def lease_guard(wake):
    enabled = os.getenv("LEADER_ELECTION_ENABLED", "false").lower()
    if enabled not in {"true", "false"}:
        raise ValueError("LEADER_ELECTION_ENABLED precisa ser true ou false")
    if enabled == "false":
        return lambda: None

    lease_config = client.Configuration.get_default_copy()
    lease_config.retries = 0
    lease_api = client.CoordinationV1Api(client.ApiClient(lease_config))
    leadership = Leadership(lease_api, STACK[2], os.environ["POD_UID"])
    leadership.acquire()
    Thread(target=leadership.renew, args=(wake,), daemon=True).start()
    return leadership.check


def main():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    poll_interval = interval("POLL_INTERVAL_SECONDS", "300")
    watch_timeout = interval("WATCH_TIMEOUT_SECONDS", "60")
    watch_retry = interval("WATCH_RETRY_SECONDS", "2")
    config.load_incluster_config()
    api = client.CustomObjectsApi()
    wake = Event()
    check_leadership = lease_guard(wake)
    Thread(target=observe, args=(
        api, STACK, {"field_selector": f"metadata.name={STACK_NAME}"},
        wake, watch_timeout, watch_retry, stack_revision,
    ), daemon=True).start()
    Thread(target=observe, args=(
        api, WORKFLOW, {"label_selector": f"{EXERCISE_LABEL}=lab06"},
        wake, watch_timeout, watch_retry,
    ), daemon=True).start()
    logging.info("Observando TerraformStack %s/%s com polling de %ss", STACK[2], STACK_NAME, poll_interval)
    while True:
        wake.clear()
        try:
            check_leadership()
            reconcile(api, check_leadership)
        except (ApiException, HTTPError) as error:
            logging.warning("Reconciliação será repetida: %s", getattr(error, "status", type(error).__name__))
        wake.wait(poll_interval)


if __name__ == "__main__":
    main()
