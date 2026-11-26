# Etapa 04 — API do TerraformStack via GitOps

## Objetivo

Instalar o contrato `TerraformStack` no Hub e conferir o que a Kubernetes API aceita.

## O que será feito

Sincronizar a Application do CRD e usar as solicitações do Simulator em validação no API server.

## Por que fizemos assim

O CRD `v1alpha1` é o contrato instalado para as próximas etapas. `sourceRef` declara a fonte, `target` identifica o destino lógico e `status` recebe o estado da execução. O controller da próxima etapa usa o mesmo contrato para orquestrar a pipeline.

## 1. Sincronizar e validar

Com o Argo CD central da [Etapa 02](02-argocd-central.md) e o Argo Workflows da [Etapa 03](03-argo-workflows.md), abra **terraformstack-crd**. Confira a revisão, o destino Hub e o diff com um CRD de escopo Namespaced. Faça o sync manual. Para uma revisão em branch, use um preview fixado no SHA publicado com o mesmo projeto e caminho. [Entrada do catálogo](../platform/hub/addons/catalog/terraformstack-crd.yaml)

```bash
cluster="${HUB_KIND_NAME:-platform-management}"
hub() { kubectl --context "kind-$cluster" "$@"; }

hub wait --for=condition=Established crd/terraformstacks.platform.example.io --timeout=90s
hub explain terraformstack.spec.sourceRef
hub explain terraformstack.spec.configuration.version
kubectl kustomize control-plane/terraformstack/examples/simulator |
  hub apply --dry-run=server --validate=strict -f -
```

Repita a validação em uma cópia do manifesto com `sourceRef.path: foundation/../simulator` e confira a rejeição pelo schema. Mantenha o CRD instalado para a Etapa 05.

## O que observar

A API publica os campos de fonte, destino, ciclo de vida e reconciliação. O dry-run valida as duas solicitações sem criar CRs. Se o tipo estiver ausente, confira a condição `Established` e o resultado do sync da Application.

[← Etapa 03](03-argo-workflows.md) · [Etapa 05 →](05-terraformstack-runtime.md)
