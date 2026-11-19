# 🧪 Roteiro prático da POC

[← Página inicial](../README.md) · [📖 Fundamentos](fundamentos/README.md) · [🏗️ Arquitetura](architecture.md)

**Antes de implementar:** consultar a [base de Fundamentos de Platform Engineering](fundamentos/README.md), com 24 artigos conceituais em pt-BR. Os labs não substituem o estudo de problemas, produto, Golden Paths, contratos e reconciliação.

**Organização da POC:** as quatro camadas **Foundation → Entity Core Infrastructure → Domain Infrastructure → Applications** são representadas em pastas do mesmo monorepo, com ownership diferente. O Crossplane é central; Argo CD central sincroniza contratos e Argo CD local entrega aplicações. Ver [arquitetura](architecture.md) e [ADR-0012](adr/0012-federated-ownership-and-gitops.md).

**Objetivo:** construir a plataforma progressivamente, com Hub local em `kind` e **provisionamento AWS real**. A criação de contas é uma fase posterior; os primeiros exercícios usam conta preexistente autenticada. Cada lab apresenta finalidade, pré-requisitos, comandos, evidências de execução, falha controlada, recuperação, custo e limpeza.

| Fase | Tema | Evidência e critérios de saída |
| --- | --- | --- |
| 00 | Baseline documental, ADRs, Git e segurança | Estrutura e links coerentes; nenhuma credencial versionada |
| 01 | Hub `kind`, Kubernetes API, RBAC e observação | Acesso controlado, eventos e logs inspecionados |
| 02 | Conectividade AWS real: perfis/roles, STS e pré-voo | Conta/região explicitadas, permissões mínimas, custos e plano de limpeza |
| 03 | Crossplane Core, XRD, Composition e Provider AWS | Criar e consultar Custom Resource pela mesma Kubernetes API; reconciliar de verdade |
| 04 | Capability inicial `ObjectStorage` com AWS real | Bucket de laboratório real, status, drift e limpeza controlada |
| 05 | Capability `Database` e demais recursos gradualmente | Provisionamento real com limites de custo, confirmação e teardown |
| 06 | CRD `AWSEnvironment` e Environment Controller próprio | Adotar conta preexistente, status.conditions, reconciliação idempotente |
| 07 | Argo Workflows, Terraform Executor, gates e aprovações | Workflow vinculado ao Environment, plan/apply autorizado e auditável |
| 08 | GitOps e Argo CD no Hub | PR → CR → reconciliador → status; correção de drift |
| 09 | XRD `Domain`: Namespace e políticas | Composition cria recursos no destino permitido; restrições de clusters remotos testadas |
| 10 | UAT e PROD com **duas contas AWS existentes** | Permissões, workloads e recursos isolados; conta-alvo validada |
| 11 | **EKS Auto Mode**: Management EKS e PCI/Non-PCI, incremental | Provisionar um cluster real por vez; validar NodePools, integração com Argo CD, operators, custos e isolamento |
| 12 | Observability Plane ponta a ponta | Correlação de logs, métricas, traces e reconciliação |
| 13 | Backstage, CLI, AI/MCP | Consumo dos mesmos contratos sem API paralela |
| 14 | **Criação automatizada de conta AWS** via AWSEnvironment | Argo Workflows coordena criação, rede, segurança e dependências humanas |
| 15 | PCI, identidade, custos e recuperação | Revisão de privilégios, restauração e teardown seguro |

**Nota sobre ordem:** a sequência dos labs demonstra primeiro um recurso AWS de baixo custo por Crossplane e só depois implementa o lifecycle completo Foundation → Core → Domain → Applications. A topologia de uma entidade completa não precisa existir para validar a primeira Capability.

## 🔐 Regras técnicas e pedagógicas
- Crossplane **é parte do Control Plane** desde os primeiros labs; o Workflow Plane possui Argo Workflows, Terraform Executor e Argo CD.
- `AWSEnvironment` é CRD nativo com controller próprio, inicialmente operando sobre **contas existentes**; `Domain` e Capabilities são XRDs com Compositions.
- A AWS é real: executar alterações apenas com conta/região conferidas via STS, credenciais temporárias, avaliação de custos e aprovação específica por operação. Preferir recursos baratos e descartáveis antes de banco/compute.
- Separar `plan` de `apply`, não versionar Terraform state ou credenciais e documentar `destroy` e remoção dos recursos gerenciados com cautela.
- Ter testes de mocks para lógica interna não significa substituir os labs AWS por simulação.
- Não pressupor integração de chamados, criação de conta na fase inicial ou acesso administrativo compartilhado PCI/Non-PCI.

## ☁️ Padrão EKS
A criação dos três EKS de cada ambiente pertence a **Entity Core Infrastructure** (Terraform). O Lab 11 utiliza **Amazon EKS Auto Mode**, não node groups autogerenciados como padrão. O Hub continua em `kind` local. Verificar compatibilidade com DaemonSets, requisitos PCI, identidade de Pods, rede, storage e cobrança antes do provisionamento. Escalar gradualmente de um cluster para a topologia-alvo, sem criação automática de todos os clusters.


---

[← Página inicial](../README.md) · [Fundamentos](fundamentos/README.md) · [Arquitetura](architecture.md)
