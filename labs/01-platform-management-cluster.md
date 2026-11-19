# Lab 01 — Platform Management Cluster com kind

[← Labs](README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos: planos](../docs/fundamentos/17-planos-e-responsabilidades.md) · [Próximo passo: roadmap](../docs/learning-roadmap.md)

## Objetivo e limite

Criar o **Hub local** que futuramente hospedará Control Plane, Workflow Plane e Observability Plane. Neste laboratório **não** instalamos Crossplane, Argo CD, Terraform, Providers AWS, EKS ou Capabilities. O objetivo é **ver a infraestrutura nascendo e funcionando**, sem surpresas na conta AWS.

## Pré-requisitos

- Docker instalado e funcionando; verificar `docker info`.
- `kind` e `kubectl` no `PATH`, com versões suportadas e compatíveis.
- Recursos locais suficientes para um nó control-plane; internet para baixar a imagem do kind quando necessário.
- **Nenhuma credencial AWS é necessária.** Caso haja outros clusters no kubeconfig, os scripts usam explicitamente `--context kind-platform-management`.

Do diretório raiz do projeto:

```bash
kind version
kubectl version --client
docker info
```

## Etapa 1 — Ler a definição

Abra [`kind-config.yaml`](../platform/hub/kind-config.yaml). O cluster se chama `platform-management` e contém **um nó control-plane**, suficiente para o primeiro lab. Esse **não é um EKS** nem um cluster por plano: é o Hub único que receberá as ferramentas por etapas.

## Etapa 2 — Criar e observar

```bash
bash platform/hub/scripts/create.sh
```

O script valida Docker/`kind`/`kubectl`, só cria o cluster se não existir, espera o nó ficar `Ready` e mostra os nós. Observe também:

```bash
kubectl --context kind-platform-management cluster-info
kubectl --context kind-platform-management get nodes -o wide
kubectl --context kind-platform-management get pods -n kube-system
kubectl --context kind-platform-management get namespaces
```

**O que observar:** API Server acessível, control-plane `Ready`, pods de sistema iniciados e namespaces padrão. Os planos da plataforma ainda são **responsabilidades arquiteturais**, não componentes instalados.

## Etapa 3 — Verificar e diagnosticar

```bash
bash platform/hub/scripts/verify.sh
```

Se o nó não ficar pronto, verifique `docker info`, `kind get clusters` e `kubectl --context kind-platform-management get events -A`. Para logs do cluster:

```bash
kind export logs --name platform-management
```

**Falha controlada sem alterações:** tente inspecionar um contexto inexistente:

```bash
kubectl --context contexto-inexistente get nodes
```

O erro é esperado. Compare com `kubectl --context kind-platform-management get nodes` para compreender por que os scripts usam um contexto explícito.

## Etapa 4 — Limpeza

A remoção do cluster é **opcional**; se for prosseguir para o Lab Crossplane, mantenha-o. Para recriar do zero, execute:

```bash
bash platform/hub/scripts/delete.sh --confirm
bash platform/hub/scripts/create.sh
bash platform/hub/scripts/verify.sh
```

**Aviso:** `delete.sh` remove o cluster **local `platform-management`** e todos os objetos dentro dele. Não toca na AWS. Sem `--confirm`, ele se recusa a excluir.

## Critérios de conclusão

- [ ] `kind` cria o Hub e um nó fica `Ready`.
- [ ] `kubectl` mostra API, namespaces e pods `kube-system`.
- [ ] A verificação executa sem erros e é possível identificar o contexto correto.
- [ ] Foi observada uma falha controlada e entendida a recuperação.
- [ ] O procedimento de remoção/recriação foi compreendido (executar apenas se desejado).

## Próximo laboratório

Instalar e inspecionar o **Crossplane no Hub** antes de configurar qualquer conta AWS. Depois provisionaremos `ObjectStorage` e `Database` com **Amazon Aurora PostgreSQL**, usando **somente UAT inicialmente**. O método de autenticação AWS será definido na etapa específica; o [template UAT](../foundation/config/entity-a/uat/environment.example.yaml) não contém credenciais.
