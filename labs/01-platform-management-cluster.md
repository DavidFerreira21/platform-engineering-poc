# Etapa 01 — Preparação do Platform Management Cluster: kind, rede e Traefik

[← Guias das etapas](README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos: planos](../docs/fundamentos/17-planos-e-responsabilidades.md) · [Roadmap](../docs/learning-roadmap.md)

## Objetivo

Preparar um Hub Kubernetes local reproduzível e verificar o caminho de uma requisição HTTP do host até um Pod.

## O que será feito

- Criar ou reutilizar o cluster kind `platform-management` e preparar os namespaces `platform-system`, `traefik`, `argocd` e `crossplane-system`. Os dois últimos ficam reservados para as próximas etapas.
- Instalar o Traefik via Helm e publicar o Deployment, o Service e o Ingress do serviço de demonstração `echo`.
- Validar `demo.localhost:8080`, provocar um `404` com um host desconhecido e recuperar usando o host correto.
- Reexecutar a instalação no mesmo Hub ou, na limpeza opcional, remover e recriar o cluster local.

O exercício roda no Hub local e tem custo AWS zero.

## Pré-requisitos

- Docker funcional, `kind`, `kubectl`, `helm` e `curl` no PATH.
- Conexão para baixar imagens e o Helm chart oficial do Traefik.
- Portas **8080 e 8443 TCP** disponíveis no host para o kind. O bind é exclusivamente em **127.0.0.1** (não expõe o laboratório na LAN).
- Sem AWS, sem secrets e sem alteração do contexto Kubernetes padrão.
- Ambiente local com recursos suficientes para kind, Traefik e pod de teste.

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

**Escolha didática:** usamos `networking.k8s.io/v1/Ingress` em vez de desenvolver um CRD de plataforma ou usar IngressRoute. O Helm chart Traefik pode instalar CRDs de terceiros, mas não criamos controllers/contratos próprios. O dashboard não é publicado.

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

A porta 8443 fica preparada para HTTPS futuro, mas **não configuramos CA/certificado TLS confiável** nesta etapa. Não expor dashboard, secrets ou endpoints administrativos.

## Etapa 5 — Observar, provocar erro e recuperar

```bash
kubectl --context kind-platform-management -n platform-system get ingress,svc,deploy,pods
kubectl --context kind-platform-management -n traefik logs deploy/traefik --tail=100
curl --noproxy '*' -i --resolve desconhecido.localhost:8080:127.0.0.1 http://desconhecido.localhost:8080/
```

**Falha controlada:** a última requisição deve retornar **404** do Traefik, pois não há uma regra Ingress para `desconhecido.localhost`. Em seguida faça a requisição para `demo.localhost` e observe o sucesso. Isso prova o roteamento baseado no header HTTP `Host`, não apenas que o pod respondeu. Se surgir `connection refused`, confira o mapping/porta. Se houver `503`, confira endpoints e health do backend. Se houver `404` para a rota correta, confira `IngressClass` e regras.

## Etapa 6 — Limpeza e reexecução

Para **repetir a etapa sem apagar o cluster**, `create.sh` e `install-traefik.sh` podem ser chamados novamente (manifests declarativos / Helm upgrade). Para **zerar todo o Hub**, inclusive configurações futuras, é necessário confirmar explicitamente:

```bash
bash platform/hub/scripts/delete.sh --confirm
bash platform/hub/scripts/create.sh
bash platform/hub/scripts/install-traefik.sh
bash platform/hub/scripts/verify.sh
```

**Atenção:** o `delete.sh` apaga **todo o cluster kind** e seus objetos, incluindo eventuais instalações posteriores de Argo CD/Crossplane. Não afeta AWS, mas jamais assuma que recursos externos gerenciados por esses componentes foram removidos.

## Critérios de aceite

- [ ] API Kubernetes e CoreDNS funcionam, nó está `Ready`.
- [ ] Namespaces esperados criados, sem componentes Argo CD/Crossplane instalados.
- [ ] Traefik está saudável e usando NodePorts 30080/30443.
- [ ] `demo.localhost:8080` retorna `platform-management-ok`.
- [ ] Host incorreto retorna `404` e é possível explicar a rota do tráfego.
- [ ] Scripts podem ser executados novamente sem duplicar recursos.
- [ ] Limpeza destrutiva requer `--confirm`.

## Próximas etapas

Siga para a [Etapa 02 — Argo CD central](02-argocd-central.md), a [Etapa 03 — Argo Workflows](03-argo-workflows.md) e a [Etapa 04 — Crossplane Core](04-crossplane-core.md). Ao adotar Traefik no Argo CD, documentar uma **transferência de ownership Helm → Argo CD**, sem dois reconciliadores gerenciando o release simultaneamente. Não executar o bootstrap do Traefik manual depois de adotado pelo GitOps.

Fontes: [kind — Configuration](https://kind.sigs.k8s.io/docs/user/configuration/) · [Traefik — Setup on Kubernetes](https://doc.traefik.io/traefik/setup/kubernetes/) · [Traefik Helm Chart](https://github.com/traefik/traefik-helm-chart).
