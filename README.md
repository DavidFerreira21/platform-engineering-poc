# Platform Engineering POC

Nesta POC, construímos uma plataforma interna para testar padrões de **Engenharia de Plataforma** e consolidar aprendizados sobre APIs de infraestrutura, GitOps, governança e experiência do desenvolvedor. O repositório registra as decisões, os resultados dos experimentos e os guias para reproduzir a construção por etapas.

## Arquitetura

![Arquitetura de referência da POC com Hub EKS, contas UAT e PROD, três personas e quatro camadas de ownership](docs/arquitetura%20Plataform.png)

O desenho mostra a **topologia-alvo**. O Hub EKS concentra o Control Plane e o Workflow Plane. UAT e PROD têm contas AWS separadas, com infraestrutura compartilhada e destinos autorizados para workloads. A observabilidade mantém coleta e armazenamento por ambiente e centraliza a consulta na conta da plataforma.

O **Platform Team** mantém a base e os contratos. O **DevOps Federado** gerencia a infraestrutura dos Domains por meio desses contratos. As **squads** mantêm o código, o CI e os manifests das aplicações em seus próprios repositórios. Uma aplicação de referência vai exercitar o consumo de Capabilities e a entrega por GitOps. [Diagrama em tamanho completo](docs/arquitetura%20Plataform.png) · [Arquitetura detalhada](docs/architecture.md) · [Personas e responsabilidades](docs/adr/0007-personas-and-developer-experience.md)

## Onde a POC está hoje

- **Hub EKS.** O bootstrap inicial registrou acesso à API, NodePools em `Ready` e a capability Argo CD em `ACTIVE`. O próximo marco é validar o primeiro workload e o primeiro sync gerenciado. [Bootstrap](labs/01-platform-management-cluster.md#o-que-observar) · [Primeiro sync](labs/02-argocd-central.md)
- **Control Plane.** Os ensaios iniciais registraram a execução do TerraformStack com o Simulator e a reconciliação de um ConfigMap pelo Crossplane. Os aceites desses fluxos no Hub EKS seguem nas Etapas 03–08. [Evidências do Simulator](labs/06-terraformstack-simulator.md#evidências-de-aceite) · [Evidências do Crossplane](labs/08-crossplane-gitops.md#evidências-de-aceite)
- **Spokes e observabilidade.** A integração com UAT, os primeiros recursos gerenciados por Capabilities e o aceite da telemetria estão no roteiro. [Estado e sequência das etapas](docs/learning-roadmap.md#estado-da-poc)

## Padrões em avaliação

| Padrão | Pergunta que a POC investiga | Estado |
| --- | --- | --- |
| **GitOps central e local** | Como distribuir add-ons e intenções de infraestrutura pelo Hub e entregar aplicações nos destinos autorizados? | Capability Argo CD ativa no bootstrap inicial. Primeiro sync e acesso às spokes a validar. [Fluxo GitOps](docs/flows/gitops.md) |
| **Terraform por API declarativa** | Como um CRD pode representar a intenção de infraestrutura e coordenar plano, aprovação e reconciliação por uma pipeline? | CRD, controller e pipeline versionados. Simulator exercitado com recursos de teste. [TerraformStack](docs/flows/terraformstack.md) · [Contrato de execução](docs/flows/terraformstack-execution-contract.md) |
| **Crossplane como API de infraestrutura** | Como expor Capabilities especializadas e manter o lifecycle sob um único reconciliador? | XRD e Composition exercitados com um ConfigMap. `ObjectStorage` e `Database` são os casos AWS previstos. [Crossplane no GitOps](docs/flows/gitops.md#4-infrastructure-gitops--crossplane) |
| **GitOps Bridge** | Como publicar outputs verificados do Terraform para consumo por componentes Kubernetes no ambiente? | Raízes e publicador versionados. Publicação e consumo no ambiente a validar na Etapa 17. [Fluxo do Bridge](docs/flows/gitops-bridge.md) |
| **Observabilidade por ambiente** | Como concentrar a consulta no Grafana e manter a ingestão de métricas durante uma falha do Gateway? | Topologia definida. Aceite de ingestão, consulta e falha controlada previsto em UAT nas Etapas 26–27. [Fluxo de observabilidade](docs/flows/observability.md) |
| **Developer Experience e Golden Paths** | O que uma squad precisa conhecer para entregar uma aplicação e o que a plataforma pode oferecer por contratos e padrões reutilizáveis? | Personas e responsabilidades definidas. Starter Kit e jornada ponta a ponta previstos nas Etapas 32–35. [Personas e Developer Experience](docs/adr/0007-personas-and-developer-experience.md) |

## Acompanhar ou reproduzir

- **Acompanhar a construção.** O roteiro mostra a sequência, as dependências e os critérios de aceite por etapa. [Roteiro](docs/learning-roadmap.md)
- **Reproduzir um marco.** O índice distingue guias executáveis de etapas que ainda têm critérios de aceite. O percurso começa no bootstrap do Hub EKS. [Laboratórios](labs/README.md)
- **Entender as decisões e encontrar os componentes.** [ADRs](docs/adr/README.md) · [Estrutura do monorepo](docs/repository-structure.md) · [Fundamentos de Platform Engineering](docs/fundamentos/README.md)

Antes de operar na AWS, confirme identidade, conta, região, permissões, custo e autorização da operação. [Segurança e entrega incremental](docs/adr/0004-aws-security-and-roadmap.md)
