"""Reconcilia a solicitação de demonstração da Etapa 06."""

import logging
from datetime import datetime, timezone
from time import sleep

from kubernetes import client, config
from kubernetes.client.exceptions import ApiException
from urllib3.exceptions import HTTPError

STACK = ("platform.example.io", "v1alpha1", "platform-system", "terraformstacks")
WORKFLOW = ("argoproj.io", "v1alpha1", "argo-workflows", "workflows")
STACK_NAME = "lab06-network"
TEMPLATE_NAME = "terraformstack-plan-simulado"
CONDITION = "SimulatedPlanReady"
UID_LABEL = "lab.platform.example/stack-uid"
GENERATION_LABEL = "lab.platform.example/stack-generation"
TIMEOUT = 10


def update_condition(api, stack, state, reason, message):
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
    metadata = stack["metadata"]
    spec = stack["spec"]
    uid, generation = metadata["uid"], metadata["generation"]
    return {
        "apiVersion": "argoproj.io/v1alpha1",
        "kind": "Workflow",
        "metadata": {
            "name": f"lab06-plan-{uid}-{generation}",
            "namespace": WORKFLOW[2],
            "labels": {
                "lab.platform.example/exercise": "lab06",
                UID_LABEL: uid,
                GENERATION_LABEL: str(generation),
            },
        },
        "spec": {
            "workflowTemplateRef": {"name": TEMPLATE_NAME},
            "arguments": {"parameters": [
                {"name": "module", "value": spec["module"]["name"]},
                {"name": "version", "value": spec["module"]["version"]},
                {"name": "entity", "value": spec["target"]["entity"]},
                {"name": "environment", "value": spec["target"]["environment"]},
            ]},
        },
    }


def reconcile(api):
    try:
        stack = api.get_namespaced_custom_object(*STACK, STACK_NAME, _request_timeout=TIMEOUT)
    except ApiException as error:
        if error.status == 404:
            return
        raise
    if stack["metadata"].get("deletionTimestamp"):
        return

    desired = workflow_for(stack)
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
            update_condition(api, stack, "False", "WorkflowMissing", f"Workflow {name} removido.")
            return
        # Se outra instância criou o Workflow entre GET e POST, o próximo ciclo o encontra.
        workflow = api.create_namespaced_custom_object(*WORKFLOW, desired, _request_timeout=TIMEOUT)
        logging.info("Workflow %s criado em %s", name, WORKFLOW[2])

    labels = workflow["metadata"].get("labels", {})
    expected_labels = desired["metadata"]["labels"]
    if any(labels.get(key) != value for key, value in expected_labels.items()) or (
        workflow["spec"].get("workflowTemplateRef") != desired["spec"]["workflowTemplateRef"]
    ):
        update_condition(api, stack, "False", "WorkflowMismatch", f"Workflow {name} sem a correlação esperada.")
        return

    phase = workflow.get("status", {}).get("phase") or "Pending"
    state, reason = {
        "Succeeded": ("True", "WorkflowSucceeded"),
        "Failed": ("False", "WorkflowFailed"),
        "Error": ("False", "WorkflowError"),
    }.get(phase, ("Unknown", "WorkflowInProgress"))
    update_condition(api, stack, state, reason, f"Workflow {name} em fase {phase}.")


def main():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    config.load_incluster_config()
    api = client.CustomObjectsApi()
    logging.info("Observando TerraformStack %s/%s", STACK[2], STACK_NAME)
    while True:
        try:
            reconcile(api)
        except (ApiException, HTTPError) as error:
            logging.warning("Reconciliação será repetida: %s", getattr(error, "status", type(error).__name__))
        sleep(5)


if __name__ == "__main__":
    main()
