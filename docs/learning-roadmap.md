# Roteiro prático da POC

[← Início](../README.md) · [Arquitetura](architecture.md) · [Guias das etapas](../labs/README.md) · [Fundamentos](fundamentos/README.md)

O roteiro tem **sete fases** e uma sequência principal de **25 etapas, de 00 a 24**. A Etapa 11 tem três entregas ordenadas (**11A, 11B e 11C**) sem alterar a numeração principal. Cada fase agrupa uma entrega verificável. Cada guia começa com **Objetivo** e **O que será feito**, seguidos de pré-requisitos, comandos, evidências, falha controlada, diagnóstico, recuperação, custo e limpeza.

A **Fase 1** reúne os guias disponíveis. As Fases 2 a 7 descrevem as próximas entregas da POC.

## Fase 1 — Preparação do Hub e dos ambientes AWS

Preparar e testar as ferramentas no Kubernetes local. Registrar as contas UAT e PROD e confirmar o acesso humano a cada destino.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **00** | [Documentação e monorepo](repository-structure.md) | Arquitetura, decisões, estrutura e ownership alinhados |
| **01** | [Hub kind e Traefik](../labs/01-platform-management-cluster.md) | API Kubernetes, namespaces, CoreDNS, portas e Ingress HTTP local verificados |
| **02** | [Argo CD central](../labs/02-argocd-central.md) | Sincronização GitOps, drift e restauração de um ConfigMap |
| **03** | [Argo Workflows](../labs/03-argo-workflows.md) | Controller e CRDs instalados por Argo CD. Workflow local observado em sucesso, falha controlada e reexecução |
| **04** | [Crossplane Core](../labs/04-crossplane-core.md) | XRD e Composition de teste reconciliam um ConfigMap local |
| **05** | [Registro de UAT e PROD](../labs/05-registro-ambientes-aws.md) | Dois registros de Environment versionados. Conta e região conferidas e acesso humano comprovado via STS |

## Fase 2 — Control Plane e contratos

Definir contratos declarativos e provar a validação local da intenção antes de ligar os executores às contas AWS.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **06** | Contratos de Entity, Environment e Domain | Registros lógicos validados por schema, ownership e destinos permitidos |
| **07** | CRD Kubernetes `TerraformStack` | A API valida entradas e expõe status para stacks de Foundation e Entity Core. O contrato de execução segue o [ADR-0006](adr/0006-terraformrun-and-application-ci.md) |
| **08** | XRDs e Compositions de `ObjectStorage` e `Database` | Contratos Crossplane, seleção de Composition e condições de status testados localmente |

## Fase 3 — Workflow Plane e execução autorizada

Ligar as solicitações declarativas a identidades e execuções rastreáveis. O Argo CD central sincroniza os CRs. O `TerraformStack` Controller coordena o Argo Workflows e o Terraform Executor.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **09** | Identidades de workload do Terraform Executor e do Provider AWS | Cada workload assume a role prevista na UAT com credenciais temporárias. O teste confirma renovação de sessão e rejeição de identidade ou destino indevido |
| **10** | Controller, Argo Workflows e Terraform Executor simulado | Um `TerraformStack` sincronizado do Git dispara um Workflow e atualiza status. Reexecução segura e falha observável |
| **11** | TerraformStack incremental (11A → 11B → 11C) | Três entregas verificáveis, sem ativar reconciliação automática antes de validar deploy e drift |
| **11A** | Deploy ponta a ponta com aprovação na Argo Workflows UI | `TerraformStack` → Workflow → plan → `suspend`/retomada autorizada por RBAC na UI → revalidação técnica do plano → apply do mesmo plano; state segregado, lock, status e auditoria |
| **11B** | Detecção de drift | Plan periódico identifica alteração externa e publica condição/alerta sem remediação |
| **11C** | Reconciliação por políticas | Políticas versionadas distinguem auto apply permitido, aprovação obrigatória e bloqueio, incluindo cenários destrutivos |

**Autenticação e RBAC da 11A:** Argo Server em `client` mode, Viewer/Approver com permissões namespace-scoped e testes de suspensão/retomada pela UI em um [Workflow simulado](terraformstack-11a-argo-client-rbac.md). A interface permanece desabilitada na Etapa 03 e será habilitada via GitOps durante a 11A.

**Contrato de status da 11A:** `status.phase` é um resumo derivado de `status.conditions` (fonte de verdade), com `observedGeneration`, `currentExecution`, `lastPlan` e `lastSuccessfulExecution`. Validar transições, autorização/expiração, falhas, idempotência e ausência de dados sensíveis. [ADR-0006](adr/0006-terraformrun-and-application-ci.md#decisão-4--contrato-de-status-phase--conditions).

## Fase 4 — Provisionamento e smoke test do Data Plane

Exercitar recursos reais primeiro em UAT. Terraform tem ownership de rede e EKS. Crossplane tem ownership das Capabilities AWS delegadas. Provisionar um EKS por vez e validar custo, compatibilidade com Auto Mode e operadores.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **12** | Rede da UAT via Terraform | Módulo Foundation reutilizável. Plano aprovado, rede aplicada e sub-redes privadas verificadas |
| **13** | `ObjectStorage` via Crossplane | Bucket de teste em UAT, condições de prontidão, drift e política de remoção verificados |
| **14** | `Database` com Aurora PostgreSQL via Crossplane | Cluster Aurora em sub-redes privadas com IAM, conectividade, custo e teardown aprovados |
| **15** | Management EKS | Stack Terraform aplicada em UAT e cluster acessível com a identidade autorizada |
| **16** | EKS Non-PCI e Argo CD local | EKS Non-PCI provisionado em UAT. Argo CD local do Management EKS sincroniza apenas os destinos autorizados |
| **17** | Smoke test com aplicação HTTP mínima | App Python + Flask de repositório separado entregue em namespace de teste no Non-PCI, com health check e rollback verificados |

## Fase 5 — Observabilidade e operação

Instrumentar a plataforma e a aplicação de teste. Usar evidências para diagnosticar falhas, recuperar o serviço e validar os limites antes de expandir.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **18** | Coleta de logs, métricas e traces | OpenTelemetry no Data Plane e visualização central com Grafana, Loki e Tempo, correlacionando sinais da aplicação e dos controllers |
| **19** | Diagnóstico, recuperação e custos | Falha controlada detectada, investigada e recuperada. Evidências de alertas, auditoria, custos e reexecução |
| **20** | Isolamento UAT/PROD e PCI/Non-PCI | Identidade, GitOps, acesso e destinos testados com casos permitidos e negados. Expansão de clusters condicionada aos resultados, custo e autorização |

## Fase 6 — Experience Layer e Developer Experience

Preparar os caminhos de consumo da plataforma para as três personas, com escopo proporcional à POC.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **21** | Responsabilidades e permissões por persona | Platform Team, DevOps Federado e Application Team exercitam PRs por path e repositório, IAM, RBAC, Admission, AppProjects e restrições de Environment |
| **22** | Canais de consumo, catálogo e Starter Kit | Contratos e exemplos acessíveis às personas. Backstage, CLI e Git são canais possíveis. O Starter Kit contém apenas o necessário para a aplicação de teste |

## Fase 7 — Golden Path de ponta a ponta

Reutilizar a aplicação da Etapa 17 para demonstrar a jornada completa de um Domain, da solicitação de infraestrutura à entrega observável.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **23** | Primeiro Domain e Capabilities via GitOps | Registro de Domain com ownership, namespace e políticas em destino autorizado. DevOps Federado solicita Capabilities por PR e observa a reconciliação pelo Argo CD central e Crossplane |
| **24** | Aplicação de Domain e validação ponta a ponta | Repositório da squad demonstra PR, GitHub Actions, imagem, promoção GitOps, Argo CD local, telemetria e rollback |

## Regras transversais

- **Registros e contratos.** Entity, Environment e Domain são registros lógicos. As Capabilities usam XRDs Crossplane. `TerraformStack` é o CRD de Foundation e Entity Core. O controller valida a intenção e coordena Workflow, política, plano aprovado e status. [Contratos](adr/0001-platform-contracts.md) · [TerraformStack](adr/0006-terraformrun-and-application-ci.md)
- **Segurança e custos.** Conferir identidade e conta via STS, região, IAM mínimo, custo, autorização por operação e plano de remoção. Manter credenciais e Terraform state fora do Git. Aplicar isolamento desde o primeiro acesso AWS e consolidar os testes na Etapa 20.
- **Ownership.** Platform Team mantém Foundation e Core. DevOps Federado gerencia Domain Infrastructure. Application Team mantém código, CI e manifests Kubernetes no repositório da aplicação. O Argo CD local entrega o workload no destino autorizado. [Personas](adr/0007-personas-and-developer-experience.md)

## Extensões opcionais

- **Automação de criação de contas.** Avaliar apenas se surgir necessidade de automatizar o lifecycle de contas da organização.
