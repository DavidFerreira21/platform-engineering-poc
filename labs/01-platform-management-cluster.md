# Etapa 01 — Preparação do Platform Management Cluster: kind, rede e Traefik

[← Guias das etapas](README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos: planos](../docs/fundamentos/17-planos-e-responsabilidades.md) · [Roadmap](../docs/learning-roadmap.md)

## Objetivo

Preparar um Hub Kubernetes local reproduzível e verificar o caminho de uma requisição HTTP do host até um Pod.

## O que será feito

- Criar ou reutilizar o cluster kind `platform-management` e preparar os namespaces `platform-system`, `traefik`, `argocd` e `crossplane-system`. Os dois últimos ficam reservados para as próximas etapas.
- Instalar o Traefik via Helm e publicar o Deployment, o Service e o Ingress do serviço de demonstração `echo`.
- Validar `demo.localhost:8080` e acompanhar o caminho da requisição até o Pod.
- Conferir como revisitar o exercício em um Hub que já tem etapas posteriores.

O exercício roda no Hub local e tem custo AWS zero.

## Por que fizemos assim

Começamos com um cluster `kind` e uma rota HTTP local para verificar a base que receberá os componentes da plataforma. O serviço `echo` permite seguir uma requisição do host ao Ingress, ao Service e ao Pod antes de instalar o Argo CD.

## Pré-requisitos

- Docker funcional, `kind`, `kubectl`, `helm` e `curl` no PATH.
- Conexão para baixar imagens e o Helm chart oficial do Traefik.
- Portas **8080 e 8443 TCP** disponíveis no host para o kind. O bind é exclusivamente em **127.0.0.1** (não expõe o laboratório na LAN).
- Os comandos de interação com o Hub selecionam `kind-platform-management` explicitamente.
- Ambiente local com recursos suficientes para kind, Traefik e pod de teste.

Se o Hub já estiver em uso, use **Reexaminar o Hub existente** para observar a base e revisitar os testes HTTP.

Verifique antes:

```bash
docker info
kind version
kubectl version --client
helm version
```

## Etapa 1 — Entender a rede e criar o cluster

Abra [`kind-config.yaml`](../platform/hub/kind-config.yaml): **host 127.0.0.1:8080 → node container:30080**, e **host 127.0.0.1:8443 → node container:30443**. Esses números correspondem aos **NodePorts** do Service Traefik definidos em [`values.yaml`](../platform/hub/traefik/values.yaml).

```bash
bash platform/hub/scripts/create.sh
```

O script reutiliza o cluster existente, caso haja um; nesse caso **não consegue alterar os mapeamentos de portas**. Se o cluster foi criado com uma configuração anterior, remova-o e recrie-o explicitamente (consulte a seção de limpeza). Em seguida, observe:

```bash
kubectl --context kind-platform-management cluster-info
kubectl --context kind-platform-management get nodes -o wide
kubectl --context kind-platform-management get pods -n kube-system
```

**Resultado esperado:** nó `Ready`, API acessível, CoreDNS e componentes Kubernetes básicos operacionais.

## Etapa 2 — Entender os namespaces

Abra [`namespaces.yaml`](../platform/hub/namespaces.yaml). Os namespaces `platform-system`, `traefik`, `argocd` e `crossplane-system` foram aplicados pelo `create.sh`. Os dois últimos são **somente reservas lógicas**: não há Argo CD nem Crossplane instalados.

```bash
kubectl --context kind-platform-management get namespaces --show-labels
```

**Conceito:** namespace não é um Domain. O Hub central e os Domains do Data Plane possuem responsabilidades diferentes.

## Etapa 3 — Instalar Traefik, com script independente

Examine as configurações [`values.yaml`](../platform/hub/traefik/values.yaml) e [versão do Helm chart](../platform/hub/traefik/chart-version.txt). O chart usa uma **versão fixa** para reprodução. O Service é `NodePort`, não `LoadBalancer`; no kind, não é necessário cloud load balancer nem MetalLB.

```bash
bash platform/hub/scripts/install-traefik.sh
kubectl --context kind-platform-management -n traefik get deployment,pods,service
```

O script instala/atualiza o Traefik via Helm, espera o Deployment e aplica um **Deployment + Service + Ingress Kubernetes nativo** de demonstração, definido em [`demo.yaml`](../platform/hub/traefik/demo.yaml).

O serviço `echo` usa o Ingress nativo do Kubernetes para demonstrar o roteamento HTTP. O Traefik também carrega seu provider de CRDs para publicar a rota HTTPS do Argo Workflows na Etapa 03.

## Etapa 4 — Validar o caminho de rede

```bash
bash platform/hub/scripts/verify.sh
curl --noproxy '*' --resolve demo.localhost:8080:127.0.0.1 http://demo.localhost:8080/
```

**Resposta esperada:** `platform-management-ok`. A URL também funciona no navegador em ambientes que resolvem `.localhost` como loopback: [http://demo.localhost:8080/](http://demo.localhost:8080/).

```text
Browser / curl (127.0.0.1:8080)
    → Docker port mapping (kind node :30080)
    → Traefik Service NodePort :30080
    → Traefik Pod (IngressClass traefik)
    → Ingress demo.localhost
    → Service echo :80
    → Deployment echo :5678
```

A porta 8443 recebe a rota HTTPS do Argo Workflows na Etapa 03. O serviço `echo` demonstra a entrada HTTP na porta 8080.

## Etapa 5 — Conferir a configuração e os logs

```bash
kubectl --context kind-platform-management -n platform-system get ingress,svc,deploy,pods
kubectl --context kind-platform-management -n traefik logs deploy/traefik --tail=100
```

O Ingress associa o host `demo.localhost` ao Service `echo`. Se houver `connection refused`, confira o mapeamento de portas. Para `503`, confira os endpoints e o estado do backend. Para `404` na rota esperada, confira a `IngressClass` e a regra de host.

## Etapa 6 — Reexaminar o Hub existente

Se você voltar a esta etapa depois de instalar outros componentes, use a verificação do Hub e repita os testes de requisição HTTP.

```bash
bash platform/hub/scripts/verify.sh
```

O serviço `echo` fica disponível para a verificação do Hub em etapas posteriores. A exclusão completa do cluster é uma operação de manutenção separada. [Operações do Hub](../platform/hub/README.md)

## O que observar

- [ ] API Kubernetes e CoreDNS funcionam, nó está `Ready`.
- [ ] Namespaces esperados disponíveis para os componentes do Hub.
- [ ] Traefik está saudável e usando NodePorts 30080/30443.
- [ ] `demo.localhost:8080` retorna `platform-management-ok`.

## Próximas etapas

Siga para a [Etapa 02 — Argo CD central](02-argocd-central.md), a [Etapa 03 — Argo Workflows](03-argo-workflows.md) e a [Etapa 04 — Crossplane Core](04-crossplane-core.md). Ao adotar Traefik no Argo CD, documentar uma **transferência de ownership Helm → Argo CD**, sem dois reconciliadores gerenciando o release simultaneamente. Não executar o bootstrap do Traefik manual depois de adotado pelo GitOps.

Fontes: [kind — Configuration](https://kind.sigs.k8s.io/docs/user/configuration/) · [Traefik — Setup on Kubernetes](https://doc.traefik.io/traefik/setup/kubernetes/) · [Traefik Helm Chart](https://github.com/traefik/traefik-helm-chart).
