# Etapa 04 — API do TerraformStack via GitOps

## Objetivo

Instalar o contrato `TerraformStack` no Hub e conferir o que a Kubernetes API aceita.

## O que será feito

Sincronizar a Application do CRD e usar as solicitações do Simulator em validação no API server.

## Por que fizemos assim

O CRD `v1alpha1` é o contrato instalado para as próximas etapas. `sourceRef` declara o repositório, o caminho e a versão. O mapa opcional `context` acompanha a intenção. O subrecurso `status` recebe o estado da execução. O controller da próxima etapa usa o mesmo contrato para orquestrar a pipeline.

## 1. Sincronizar e validar

Com o Argo CD e o Argo Workflows instalados, abra **terraformstack-crd**. Confira a revisão, o destino Hub e o diff com um CRD de escopo Namespaced. Faça o sync manual e aguarde a operação terminar. Para testar uma revisão publicada em branch, informe o SHA no campo **Revision** do painel **Sync** da mesma Application. [Argo CD](02-argocd-central.md) · [Argo Workflows](03-argo-workflows.md) · [Entrada do catálogo](../platform/hub/addons/catalog/terraformstack-crd.yaml)

```bash
hub() { kubectl --context kind-platform-management "$@"; }

hub wait --for=condition=Established crd/terraformstacks.platform.example.io --timeout=90s
hub explain terraformstack.spec.sourceRef
hub explain terraformstack.spec.sourceRef.version
hub explain terraformstack.spec.context
kubectl kustomize control-plane/terraformstack/examples/simulator |
  hub apply --dry-run=server --validate=strict -f -
```

Repita o dry-run com um caminho inválido. O resultado esperado é `Invalid` em `spec.sourceRef.path`. Mantenha o CRD instalado para a Etapa 05.

```bash
sed 's|foundation/stacks/simulator|foundation/../simulator|g' \
  control-plane/terraformstack/examples/simulator/stacks.yaml |
  hub apply --dry-run=server --validate=strict -f -
```

## O que observar

A API publica os campos de fonte, contexto, ciclo de vida e reconciliação. O dry-run valida as duas solicitações sem criar CRs. Se o tipo estiver ausente, confira a condição `Established` e o resultado do sync da Application.

[← Etapa 03](03-argo-workflows.md) · [Etapa 05 →](05-terraformstack-runtime.md)
