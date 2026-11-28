# Etapa 02 — Primeiro sync GitOps no Hub EKS

[← Etapa 01](01-platform-management-cluster.md) · [Guias das etapas](README.md)

## Objetivo

Validar que o Argo CD gerenciado consegue ler o Git e sincronizar um recurso no próprio Hub EKS.

## O que será feito

- Validar o login no Argo CD com o IAM Identity Center.
- Registrar o Hub por ARN e configurar as permissões para o teste.
- Sincronizar manualmente o ConfigMap de demonstração a partir de `main` e conferir a revisão e o valor aplicado.

## Por que fizemos assim

A stack da Etapa 01 já cria a capability. Nesta etapa, comprovamos a entrega de ponta a ponta. O registro por ARN permite que o Argo CD reconheça o EKS como destino. O AppProject limita o repositório, o namespace e o tipo de recurso permitidos. A role da capability precisa de autorização Kubernetes para executar a sincronização. [Registro de clusters](https://docs.aws.amazon.com/eks/latest/userguide/argocd-register-clusters.html) · [Permissões do Argo CD](https://docs.aws.amazon.com/eks/latest/userguide/argocd-permissions.html)

## 1. Conferir a capability e o destino

Execute da raiz do repositório com o perfil `platform-hub` autenticado. Confira a conta e o destino antes dos comandos que alteram o cluster.

```bash
aws sts get-caller-identity --profile platform-hub --no-cli-pager
hub_arn="$(aws eks describe-cluster --profile platform-hub --region us-east-1 \
  --name platform-hub --query 'cluster.arn' --output text)"
argo_role_arn="$(aws eks describe-capability --profile platform-hub --region us-east-1 \
  --cluster-name platform-hub --capability-name argocd --query 'capability.roleArn' --output text)"
hub() { kubectl --context "$hub_arn" "$@"; }

aws eks describe-capability --profile platform-hub --region us-east-1 \
  --cluster-name platform-hub --capability-name argocd \
  --query 'capability.{Status:status,Issues:health.issues,URL:configuration.argoCd.serverUrl}'
hub -n argocd get applications,applicationsets,appprojects
```

Com a capability em `ACTIVE`, abra a URL retornada, autentique-se pelo IAM Identity Center e confirme que o grupo autorizado consegue abrir a interface. Se houver falha no login, confira o mapeamento do grupo e o estado retornado em `Issues`.

## 2. Registrar o Hub e autorizar o sync

Criar `platform-system` se ainda estiver ausente. Os bindings abaixo usam o grupo da Access Entry da capability. O teste concede leitura para a descoberta de recursos no cluster e limita a escrita a ConfigMaps em `platform-system`.

```bash
hub create namespace platform-system --dry-run=client -o yaml | hub apply -f -

hub apply -f - <<'EOF_RBAC'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: lab02-argocd-discovery
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["get", "list", "watch"]
EOF_RBAC
hub create clusterrolebinding lab02-argocd-discovery --clusterrole=lab02-argocd-discovery \
  --group="eks-access-entry:$argo_role_arn" --dry-run=client -o yaml | hub apply -f -
hub -n platform-system create role lab02-argocd-deploy --resource=configmaps \
  --verb=get,list,watch,create,update,patch,delete --dry-run=client -o yaml | hub apply -f -
hub -n platform-system create rolebinding lab02-argocd-deploy --role=lab02-argocd-deploy \
  --group="eks-access-entry:$argo_role_arn" --dry-run=client -o yaml | hub apply -f -

hub apply -f - <<EOF_HUB
apiVersion: v1
kind: Secret
metadata:
  name: platform-hub
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: cluster
stringData:
  name: platform-hub
  server: "$hub_arn"
EOF_HUB
```

Se o Hub já estiver registrado, reutilize o cadastro existente. Confira o destino em **Settings → Clusters** na interface. Os ARNs são resolvidos na sessão do operador e ficam fora do Git. [Configuração do destino local](https://docs.aws.amazon.com/eks/latest/userguide/argocd-register-clusters.html#argocd-register-local)

## 3. Sincronizar o ConfigMap

Usar os manifests canônicos do [AppProject](../workflow-plane/argocd/bootstrap/project.yaml) e da [Application](../workflow-plane/argocd/bootstrap/application.yaml), ajustando o destino e autorizando o namespace de origem na sessão atual. A Application lê `workflow-plane/argocd/demo` na branch `main`.

```bash
sed -e "s|https://kubernetes.default.svc|$hub_arn|g" \
  -e '/^spec:$/a\  sourceNamespaces: [argocd]' \
  workflow-plane/argocd/bootstrap/project.yaml | hub apply -f -

sed "s|https://kubernetes.default.svc|$hub_arn|g" \
  workflow-plane/argocd/bootstrap/application.yaml | hub apply -f -
```

Abra `lab02-gitops-demo` na UI. Confira o commit, o destino `platform-system` e o diff do ConfigMap. Selecione **Sync → Synchronize** e consulte o resultado:

```bash
hub -n argocd get application lab02-gitops-demo \
  -o 'custom-columns=SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
hub -n platform-system get configmap lab02-gitops-demo \
  -o 'jsonpath={.data.mensagem}{"\n"}'
```

Se a Application não resolver o Git ou o destino, confira `hub -n argocd describe application lab02-gitops-demo`. Em caso de `Forbidden`, verifique o ARN da role, os bindings e as restrições do AppProject.

## O que observar

- A interface abre com o login SSO. Em **Settings → Clusters**, o Hub aparece registrado pelo ARN.
- A Application fica `Synced` e `Healthy`. A revisão corresponde ao commit de `main`.
- O ConfigMap no namespace `platform-system` contém `definido-no-git`.

## Limpar o exercício

Remover a Application, o ConfigMap, o AppProject e os bindings exclusivos do teste. Preservar o cadastro do Hub e o namespace para os próximos componentes.

```bash
hub -n argocd delete application lab02-gitops-demo --ignore-not-found
hub -n platform-system delete configmap lab02-gitops-demo --ignore-not-found
hub -n argocd delete appproject lab02 --ignore-not-found
hub -n platform-system delete rolebinding,role lab02-argocd-deploy --ignore-not-found
hub delete clusterrolebinding,clusterrole lab02-argocd-discovery --ignore-not-found
```

O catálogo de add-ons entra no Hub com AppProjects e destinos adaptados à capability. A [Etapa 03](03-argo-workflows.md) acompanha a instalação e o aceite do primeiro add-on.
