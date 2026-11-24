# Roteiro prático da POC

[← Início](../README.md) · [Arquitetura](architecture.md) · [Guias das etapas](../labs/README.md) · [Fundamentos](fundamentos/README.md)

O roteiro tem **sete fases** e **25 etapas, de 00 a 24**. Cada fase agrupa uma entrega verificável. Os guias mostram o objetivo, a decisão, como reproduzir e o que observar. [Convenção dos guias](../labs/README.md#como-ler-e-reproduzir-uma-etapa)

Os guias disponíveis cobrem a **Fase 1** e o **registro de UAT** na Etapa 07. A próxima entrega é o CRD `TerraformStack` na Etapa 05.

## Fase 1 — Preparação do Hub local

Preparar o Kubernetes local e exercitar Argo CD, Argo Workflows e Crossplane no Hub `kind`, sem acesso à AWS.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **00** | [Documentação e monorepo](repository-structure.md) | Arquitetura, decisões, estrutura e ownership alinhados |
| **01** | [Hub kind e Traefik](../labs/01-platform-management-cluster.md) | API Kubernetes, namespaces, CoreDNS, portas e Ingress HTTP local verificados |
| **02** | [Argo CD central e catálogo de add-ons](../labs/02-argocd-central.md) | `ApplicationSet` gera as Applications de Argo Workflows e Crossplane Core a partir do Git. A demonstração comprova drift e restauração de um ConfigMap |
| **03** | [Argo Workflows](../labs/03-argo-workflows.md) | A Application do catálogo instala controller e CRDs. Workflow local observado em sucesso, falha controlada e reexecução |
| **04** | [Crossplane Core](../labs/04-crossplane-core.md) | A Application do catálogo instala o Crossplane Core. XRD e Composition de teste reconciliam um ConfigMap local |

## Fase 2 — TerraformStack e primeira execução na AWS

Exercitar o contrato no Control Plane e a execução simulada no Workflow Plane. Em seguida, configurar UAT para aplicar a primeira stack real com aprovação humana.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **05** | CRD Kubernetes `TerraformStack` via GitOps | A Application `terraformstack-crd` publica o CRD. A API aceita um exemplo válido, rejeita um exemplo inválido e expõe o subrecurso de status. O teste usa dados fictícios e segue o [ADR-0006](adr/0006-terraformrun-and-application-ci.md) |
| **06** | Controller, Argo Workflows e Terraform Executor simulado | A Application `terraformstack-controller` entrega o controller e o `WorkflowTemplate`. Após sincronizar a solicitação `TerraformStack`, observar o Workflow criado pelo controller, o status, a falha e a reexecução no Hub |
| **07** | [Registro e validação de UAT](../labs/07-registro-ambientes-aws.md) | O registro de Environment vincula `entity-a/uat` à conta e à região esperadas. O acesso humano é conferido via IAM Identity Center e STS |
| **08** | Identidade do Terraform Executor e backend de estado | O Platform Team prepara a role e inicializa o backend com autorização, state segregado e lock. O executor resolve o destino pela Foundation, renova a sessão temporária e recusa identidades ou destinos indevidos |
| **09** | Rede UAT via `TerraformStack` com aprovação | Argo Workflows gera o plano da rede, pausa para aprovação individual na UI, revalida e aplica o mesmo plano. Status, auditoria e sub-redes privadas verificadas |

## Fase 3 — Provisionamento dos clusters e smoke test

Usar o fluxo Terraform já validado para provisionar um EKS por vez em UAT. Confirmar custo, compatibilidade com Auto Mode e operadores antes da execução.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **10** | Management EKS | Stack Terraform aplicada em UAT e cluster acessível com a identidade autorizada |
| **11** | EKS Non-PCI e Argo CD local | EKS Non-PCI provisionado em UAT. Argo CD local do Management EKS sincroniza apenas os destinos autorizados |
| **12** | Smoke test com aplicação HTTP mínima | App Python + Flask de repositório separado entregue em namespace de teste no Non-PCI, com health check e rollback verificados |

## Fase 4 — Capabilities AWS com Crossplane

Usar o Crossplane já instalado no Hub para validar a identidade do Provider AWS e entregar as primeiras Capabilities em UAT, sob responsabilidade do Platform Team. O consumo por uma squad entra no Golden Path.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **13** | Identidade do Provider AWS no Hub | O Provider assume uma role dedicada em UAT com credenciais temporárias. O teste confirma renovação e rejeição de identidade ou destino indevido |
| **14** | `ObjectStorage` via Crossplane | Validar o XRD e a Composition localmente, depois reconciliar um bucket de teste em UAT e conferir prontidão, drift e política de remoção |
| **15** | `Database` com Aurora PostgreSQL via Crossplane | Validar o XRD e a Composition, depois reconciliar um cluster Aurora em sub-redes privadas com IAM, conectividade, custo e teardown aprovados |

## Fase 5 — Observabilidade e operação

Instrumentar a plataforma e a aplicação de teste. Diagnosticar falhas, evoluir a reconciliação do `TerraformStack` e comprovar os limites de acesso antes de expandir.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **16** | Coleta de logs, métricas e traces | OpenTelemetry no Data Plane e visualização central com Grafana, Loki e Tempo, correlacionando sinais da aplicação e dos controllers |
| **17** | Diagnóstico, recuperação e custos | Falha controlada detectada, investigada e recuperada. Evidências de alertas, auditoria, custos e reexecução |
| **18** | Detecção de drift do `TerraformStack` | Plan periódico sobre uma stack UAT já aplicada distingue alterações externas de mudanças declaradas e publica condição ou alerta, sem remediação |
| **19** | Reconciliação do `TerraformStack` por políticas | Políticas versionadas distinguem auto apply permitido, aprovação obrigatória e bloqueio, incluindo cenários destrutivos |
| **20** | [Registro de PROD](../labs/07-registro-ambientes-aws.md#uso-do-registro-prod-na-etapa-20) e isolamento UAT/PROD e PCI/Non-PCI | Validar PROD como destino separado antes dos testes entre contas. Testar negativas de IAM e de destinos GitOps com o que já está disponível. Validar GitOps ponta a ponta em outros clusters quando houver provisionamento autorizado |

## Fase 6 — Experience Layer e Developer Experience

Preparar os caminhos de consumo da plataforma para as três personas, com escopo proporcional à POC.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **21** | Responsabilidades e permissões por persona | Platform Team, DevOps Federado e Application Team exercitam PRs por path e repositório, IAM, RBAC, Admission, AppProjects e restrições de Environment |
| **22** | Canais de consumo, catálogo e Starter Kit | Contratos e exemplos acessíveis às personas. Backstage, CLI e Git são canais possíveis. O Starter Kit contém apenas o necessário para a aplicação de teste |

## Fase 7 — Golden Path de ponta a ponta

Reutilizar a aplicação da Etapa 12 para demonstrar a jornada completa de um Domain, da solicitação de infraestrutura à entrega observável.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **23** | Primeiro Domain e Capabilities via GitOps | Validar o contrato de Domain, o ownership e os Environments permitidos. Criar namespace e políticas no destino autorizado. DevOps Federado solicita Capabilities por PR e observa a reconciliação pelo Argo CD central e Crossplane |
| **24** | Aplicação de Domain e validação ponta a ponta | Repositório da squad demonstra PR, GitHub Actions, imagem, promoção GitOps, Argo CD local, telemetria e rollback |

## Regras transversais

- **Registros e contratos.** Entity, Environment e Domain são registros lógicos. As Capabilities usam XRDs Crossplane. `TerraformStack` é o CRD de Foundation e Entity Core. O controller valida a intenção e coordena Workflow, política, plano aprovado e status. [Contratos](adr/0001-platform-contracts.md) · [TerraformStack](adr/0006-terraformrun-and-application-ci.md)
- **Segurança e custos.** Conferir identidade e conta via STS, região, IAM mínimo, custo, autorização por operação e plano de remoção. Manter credenciais e Terraform state fora do Git. Aplicar isolamento desde o primeiro acesso AWS e consolidar os testes na Etapa 20.
- **Ownership.** Platform Team mantém Foundation e Core. DevOps Federado gerencia Domain Infrastructure. Application Team mantém código, CI e manifests Kubernetes no repositório da aplicação. O Argo CD local entrega o workload no destino autorizado. [Personas](adr/0007-personas-and-developer-experience.md)

## Extensões opcionais

- **Automação de criação de contas.** Avaliar apenas se surgir necessidade de automatizar o lifecycle de contas da organização.
