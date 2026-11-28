# GitOps Flow — Platform Engineering POC

A Etapa 01 provisiona o Hub EKS com a capability Argo CD. A Etapa 02 valida o primeiro sync no próprio Hub. O Argo CD autogerenciado dos Management EKS atende à entrega de workloads nos destinos locais autorizados. [Roadmap](../learning-roadmap.md) · [ADR-0008](../adr/0008-managed-argocd-connectivity.md)

![Fluxo GitOps: Platform Addons, Crossplane, TerraformStack e Domain Applications](./gitops-flow.png)

[Visualizar o diagrama em tamanho completo](./gitops-flow.png)

## 1. Objetivo e princípio

O **Git é a fonte de verdade das configurações desejadas**. Argo CD sincroniza manifests com as APIs Kubernetes; cada controller especializado é responsável por transformar a intenção em recursos e acompanhar seu estado. O desenho separa **como a intenção é entregue** de **como a infraestrutura ou a aplicação é reconciliada**.

O GitOps abrange quatro fluxos complementares:

| Fluxo | Entrada declarativa | Responsável pela execução/reconciliação | Resultado |
| --- | --- | --- | --- |
| **Platform GitOps** | Catálogo de add-ons e Applications | Argo CD central e ApplicationSet | Componentes Kubernetes da plataforma |
| **Infrastructure GitOps — Crossplane** | CRs de Capabilities | Crossplane central e Provider AWS | S3, Aurora PostgreSQL e demais Capabilities |
| **Infrastructure GitOps — Terraform** | CR `TerraformStack` | TerraformStack Controller → Argo Workflows → Terraform Executor | ConfigMap no Simulator. Foundation, rede, IAM e EKS no alvo AWS |
| **Domain GitOps** | Manifests de aplicações no repositório da squad | Argo CD local | Aplicações nos clusters PCI/Non-PCI autorizados |

> **Importante:** o diagrama é **lógico**, não prova de conectividade entre VPCs/landing zones nem de recursos AWS já implantados. O Argo CD central não executa diretamente Terraform nem substitui o Crossplane.

## 2. Bootstrap do Control Plane

A [Etapa 01](../../labs/01-platform-management-cluster.md) usa a stack Terraform `platform-hub` para provisionar o cluster e a **EKS Capability for Argo CD** com IAM Identity Center. A [Etapa 02](../../labs/02-argocd-central.md) registra o Hub por ARN, configura permissões e sincroniza o primeiro recurso do Git.

O catálogo em `platform/hub/addons/catalog/` organiza uma Application por componente. A instalação no EKS depende da adaptação dos AppProjects, do destino e das configurações de cada add-on à capability. A descoberta de Applications pelo ApplicationSet e o sync são operações distintas. Inspecionar revisão, destino e diff antes do sync manual.

## 3. Platform GitOps — governança dos add-ons

O **Platform Team** mantém o padrão de componentes Kubernetes da plataforma ao longo de seu ciclo de vida: instalação, atualização, configuração e reconciliação.

O catálogo pode incluir componentes como External Secrets Operator, OTel Collector e add-ons de segurança. Os **recursos AWS associados**, como IAM Roles, são provisionados pelo motor de infraestrutura designado, e não automaticamente pelo Helm Chart.

### Descoberta e seleção

- **Git Generator:** descobre as entradas versionadas do catálogo.
- **Cluster Generator:** seleciona clusters **registrados e autorizados** por labels e outras informações de destino.
- **Matrix Generator:** pode combinar os dois para gerar uma Application por combinação aprovada de componente e cluster.

O laboratório já possui **Git File Generator** no `hub-addons`; a seleção por clusters e a combinação Matrix fazem parte da evolução-alvo, não da implementação atual.

Perfis previstos:

| Perfil | Escopo de add-ons |
| --- | --- |
| **Hub** | Argo Workflows, Crossplane, controllers e coleta dos sinais operacionais do Hub. A EKS Capability for Argo CD é provisionada no bootstrap do Hub. |
| **Management** | Argo CD local, OTel Collector, OTel Gateway, Loki e Tempo |
| **PCI** | Add-ons do perfil PCI, incluindo Collector quando a fronteira de telemetria for aprovada |
| **Non-PCI** | Add-ons do perfil padrão, incluindo OTel Collector |

O [fluxo de observabilidade](observability.md) detalha os destinos e acessos de cada perfil.

Cada cluster poderá receber labels como `entity`, `environment`, `cluster-role` e `compliance`. Essas labels são **metadados de seleção**, não mecanismo de autorização por si sós. O onboarding estabelece previamente acesso, identidade, RBAC e confiança do cluster.

**Overrides por Entity/Environment** são permitidos quando definidos e aprovados pelo Platform Team. O Platform Team mantém o ownership do add-on após a customização de parâmetros.

## 4. Infrastructure GitOps — Crossplane

1. O DevOps Federado propõe a criação/alteração de uma **Capability aprovada** por PR.
2. O Argo CD central sincroniza o CR correspondente na API Kubernetes do Hub.
3. O Crossplane central observa o recurso, aplica XRD/Composition e utiliza Provider AWS e identidade autorizada.
4. O Crossplane cria/atualiza e continua reconciliando os recursos AWS; suas condições e status indicam o resultado.

As duas primeiras Capabilities do roadmap são **`ObjectStorage` (S3)** e **`Database` (Aurora PostgreSQL)**. Crossplane não precisa executar um Terraform Workflow para reconciliá-las. **`Synced` no Argo CD confirma a entrega do manifesto, não necessariamente o recurso AWS pronto.**

## 5. Infrastructure GitOps — TerraformStack

1. O Platform Team (ou persona autorizada) propõe um `TerraformStack` em Git.
2. Após validação e revisão, o Argo CD central sincroniza **o CR de intenção** no Hub.
3. O **TerraformStack Controller** valida a estrutura da intenção, verifica dependências e agenda cada tentativa. O CR identifica a fonte em `spec.sourceRef` e pode informar `spec.context` e `spec.workflowRef.name`. O alias seleciona uma integração cadastrada, ou usa o default. Há no máximo um Workflow ativo por CR.
4. A pipeline de referência no **Argo Workflows** resolve a revisão Git, o destino e o state conforme o perfil confiável. O Terraform Executor disponibiliza os binários para `plan`, avaliação OPA e, quando exigida, validação da aprovação humana.
5. Com o handshake habilitado (default), o controller concede autorização e o executor registra o claim antes do apply. Uma integração compatível pode desabilitar esse handshake; nesse caso, o controller aguarda o término da execução ativa antes de agendar outra. A pipeline publica o resultado tipado e o controller atualiza `TerraformStack.status`. Logs JSON, Events e métricas expõem a saúde da reconciliação independentemente de coleta externa.

**Não há `terraform apply` em todo reconcile da API Kubernetes.** A Etapa 06 exercita Terraform, OPA, aprovação e drift no Simulator. A Etapa 13 leva o apply aprovado à Network UAT. As Etapas 28–29 validam drift e políticas na AWS.

A exclusão padrão do CR retém os recursos. O teardown habilitado explicitamente exige execução verificada; a referência governada aplica plano, política e aprovação conforme a decisão da pipeline. O fluxo detalha também concorrência, state e retry. [Fluxo do TerraformStack](terraformstack.md) · [ADR-0006](../adr/0006-terraformrun-and-application-ci.md)

## 6. Domain GitOps — aplicações das entidades

Cada **Environment** da Entity possui, na topologia-alvo, três clusters: **Management**, **PCI** e **Non-PCI**.

O Argo CD **central** instala e mantém a configuração-base do **Argo CD local** no Management Cluster, quando esse modelo for autorizado. O Argo CD **local** gerencia as Applications das squads nos clusters PCI e Non-PCI permitidos, além de extensões locais não conflitantes e expressamente autorizadas.

A squad mantém **código, testes, pipeline CI/CD e manifests Kubernetes em repositório próprio**, separado do monorepo da POC. O CI publica uma imagem e propõe uma mudança declarativa de versão. Após revisão, o Argo CD local sincroniza a aplicação; a pipeline não realiza deployment imperativo como caminho padrão.

**Nenhum recurso Kubernetes pode ser administrado pelos dois Argo CDs simultaneamente.** Os limites são aplicados por AppProjects, RBAC, admission policies, permissões de repositório e identidades. Em particular, Argo CD local não administra os add-ons corporativos cujo owner é o Argo central.

## 7. Reconciliação, segurança e drift

| Tipo de recurso | Quem detecta divergência | Comportamento |
| --- | --- | --- |
| Componentes Kubernetes da plataforma | Argo CD central | Detecta OutOfSync; `selfHeal` corrige quando habilitado e autorizado |
| Domain Applications | Argo CD local | Sync e rollback conforme políticas de aplicação |
| Capabilities AWS | Crossplane | Reconciliação contínua conforme configuração/provider |
| Foundation e EKS via Terraform | TerraformStack Controller + Terraform | Plan/avaliação de drift e remediação conforme política |

**GitOps não substitui controles preventivos.** IAM, RBAC, admission, segregação PCI/Non-PCI, revisão de PR e autorização de destinos continuam obrigatórios.

## 8. Onboarding dos clusters

O onboarding futuro tem três operações independentes:

1. **Provisionar:** TerraformStack cria EKS e suas dependências AWS.
2. **Registrar e autorizar:** automação estabelece identidade, permissões, endpoint/acesso e labels do destino no Argo CD apropriado. **Publicar outputs no Parameter Store, por si só, não registra o cluster.**
3. **Configurar:** ApplicationSet identifica o cluster autorizado e gera Applications para os perfis aprovados.

As landing zones podem ser isoladas, sem VPC Peering nem Transit Gateway. O padrão GitOps não deve pressupor conectividade irrestrita. Na AWS, as três VPCs (Platform, UAT e PROD) permanecem isoladas e têm saída via NAT. A EKS Capability for Argo CD acessa as APIs privadas por conectividade gerenciada AWS após registro por ARN e Access Entries. Essa conectividade não se estende a Pods, kubectl ou telemetria. [ADR-0008](../adr/0008-managed-argocd-connectivity.md)

## 9. Limites e GitOps Bridge

Git define **configuração e intenção declarativas**; não é obrigatório armazenar nele outputs dinâmicos, credenciais ou Terraform state.

Na Etapa 17, o GitOps Bridge publica outputs selecionados do Terraform no Parameter Store e o External Secrets Operator os materializa em Secrets nos namespaces consumidores. O ApplicationSet central consulta o cadastro criado pelo onboarding do cluster. [GitOps Bridge](gitops-bridge.md)

## 10. Implementação incremental

| Etapa | Evidência principal |
| --- | --- |
| 01–03 | Hub EKS, Argo CD gerenciado e Argo Workflows |
| 04–06 | CRD, controller, pipeline e Simulator com Terraform e OpenTofu |
| 07–08 | Crossplane Core com recurso local e aceite GitOps |
| 09–13 | Aceites do Hub, UAT, identidade, state e primeiro apply de Network |
| 14–18 | Management e Non-PCI, GitOps multicluster, Bridge e smoke test HTTP |
| 19–21 | Identidade Crossplane e Capabilities AWS |
| 22–25 | Aplicação FastAPI com Aurora/S3, GitHub Actions, ECR e Argo CD local |
| 26–31 | Observabilidade, drift, reconciliação por políticas e isolamento |
| 32–35 | Experience Layer e Golden Path integrado |

O bootstrap manual do Hub e o primeiro sync precedem a instalação dos add-ons.

## 11. Princípios arquiteturais

1. Git é a fonte de verdade da **intenção aprovada**.
2. Argo CD sincroniza contratos; motores especializados reconciliam seus respectivos domínios.
3. Os add-ons corporativos pertencem permanentemente ao Platform Team.
4. Aplicações e extensões locais autorizadas têm lifecycle independente.
5. Cada recurso Kubernetes possui **um único owner de reconciliação**.
6. Bootstrap, registro de cluster, configuração e operação são atividades distintas.
7. `Synced` na API Kubernetes não implica infraestrutura AWS convergida.
8. O desenho lógico preserva os contratos entre o Hub EKS e os destinos autorizados.

## 12. Documentos relacionados

- [Arquitetura principal](../architecture.md)
- [Roadmap](../learning-roadmap.md)
- [TerraformStack Flow](terraformstack.md) · [GitOps Bridge Flow](gitops-bridge.md)
- [ADR-0003 — Reconciliação e GitOps](../adr/0003-reconciliation-and-gitops.md)

## 13. Inspirações externas

Os exemplos abaixo mostram **formas diferentes de resolver problemas parecidos**. Eles ajudaram a orientar a separação de responsabilidades e os limites de cada fluxo, mas não representam uma arquitetura copiada integralmente.

**[RELEX Solutions — Where State Meets Sync (PlatformCon 2026)](https://2026.platformcon.com/sessions/where-state-meets-sync-scaling-platforms-without-the-hype).** A RELEX apresentou a evolução de uma plataforma AKS em que infraestrutura e releases Helm eram acoplados ao Terraform. A empresa passou a gerenciar infraestrutura com Terraform Cloud e aplicações com Argo CD, dando a cada camada seu próprio ciclo de vida. **Inspiração na POC:** separar o provisionamento da infraestrutura da entrega GitOps de componentes e aplicações, sem começar obrigatoriamente por um portal.

**[Quince — Scaling Across Continents with Crossplane](https://tech.quince.com/from-whiteboard-to-production-scaling-quince-across-continents-with-crossplane-41dfb51d37a2).** A equipe descreveu a divisão da plataforma entre **Foundation em Terraform** — redes, EKS e IAM compartilhado — e **recursos específicos das aplicações em Crossplane**, como S3, filas e bancos. **Inspiração na POC:** distinguir Foundation/Entity Core de Domain Infrastructure, dando a cada camada um motor e um ownership apropriados. Na nossa arquitetura, o Crossplane é centralizado; a implementação da Quince não precisa ser igual.

**[AWS EKS Blueprints — Multi-Cluster Hub & Spoke](https://aws-ia.github.io/terraform-aws-eks-blueprints/patterns/gitops/gitops-multi-cluster-hub-spoke-argocd/).** O exemplo da AWS provisiona clusters EKS e utiliza um Argo CD no Hub para descobrir destinos e distribuir add-ons por ApplicationSets. **Inspiração na POC:** governança central e instalação declarativa de Platform Addons em Management, PCI e Non-PCI, preservando a autonomia do Argo CD local para aplicações.

**Síntese:** a inspiração está nos **princípios de separação e reconciliação**, não na adoção das mesmas ferramentas, topologia, permissões ou políticas de cada projeto.
