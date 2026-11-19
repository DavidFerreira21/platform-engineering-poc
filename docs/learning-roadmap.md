# 🧪 Roteiro prático da POC

[← Início](../README.md) · [Arquitetura](architecture.md) · [Fundamentos](fundamentos/README.md)

Construção incremental e observável: cada lab terá pré-requisitos, comandos, evidências, falha controlada, diagnóstico, recuperação, custo e limpeza. **Primeira implementação: Lab 01 do Hub com kind**, seguido por Crossplane, `ObjectStorage` e **`Database` com Aurora PostgreSQL**. A conta inicial será apenas UAT. A decisão de autenticação AWS fica para uma fase posterior; não é necessária para criar o kind. **A ordem pedagógica difere da sequência de provisionamento de uma entidade completa.**

| Fase | Tema | Evidência |
| :--- | :--- | :--- |
| **00** | Documentação e monorepo | Conceitos e decisões consistentes |
| **01** | Hub `kind` | API Kubernetes e RBAC funcionando |
| **02** | Crossplane Core no Hub local | Instalar Crossplane e explorar CRDs, XRDs, Compositions e status sem AWS |
| **03** | UAT: registro de conta e autenticação AWS | Configurar arquivo de Environment e método de autenticação a definir; validar identidade com STS **antes de qualquer operação** |
| **04** | `ObjectStorage` real | XRD/Composition/Provider AWS, bucket de laboratório, `Ready`, drift e teardown |
| **04B** | `Database` com Amazon Aurora PostgreSQL | Contrato de Database e cluster Aurora PostgreSQL em UAT, sujeito a custo, rede, IAM, plano e teardown autorizados |
| **05** | Entity, Environment e criação padronizada de Domain | Registro único, perfil versionado, destinos selecionados e geração idempotente de estrutura/manifests |
| **06** | Argo CD central + primeiro Domain Bootstrap | Criar Domain gera manifest de `ObjectStorage` conforme perfil de lab → Crossplane → AWS; demonstrar validação e isolamento |
| **07** | Foundation Terraform reutilizável | Config central por Entity/Environment, plan/state isolado |
| **08** | Argo Workflows, gates e Terraform Executor | Aprovação e execução auditáveis |
| **09** | Entity Core e EKS Auto Mode incremental | Um EKS por vez, verificando custo, identidade e PCI |
| **10** | Argo CD local + perfil de Bootstrap completo | Criação de Domain gera Database, ObjectStorage, namespace e policies nos destinos escolhidos; evidências de entrega antes do Deployment |
| **11** | Duas contas AWS UAT/PROD | Credenciais e recursos segregados |
| **12** | Observabilidade | Status, logs, métricas, traces, auditoria |
| **13** | Backstage / CLI / AI-MCP | Consumo dos contratos sem API paralela |
| **14** | Extensão opcional: criação de contas | Justificar contrato/controlador de lifecycle se necessário |
| **15** | Segurança, custo e recuperação | Testes de isolamento e cleanup |

**Não implementar `AWSEnvironment` CRD ou Environment Controller próprio como pré-requisito inicial.** Environment e Domain são registros lógicos; somente Capabilities precisam de XRD no primeiro ciclo. O **Domain Bootstrap** é o fluxo padrão de **criação de Domain**: gera estrutura e manifests iniciais conforme perfil versionado e Environments selecionados; **o Argo CD local**, e não um controller do Domain, aplica namespaces/policies. `Database` usará **Aurora PostgreSQL** como implementação inicial do produto, após `ObjectStorage`, e será implementado progressivamente e nunca pode ser marcado pronto antes de sua entrega real.

**Conta UAT:** registrar o destino em `foundation/config/entity-a/uat/environment.example.yaml` (template sem dados reais). Conta, região, credenciais e método de autenticação ainda precisam ser definidos e validados no início dos labs AWS. Nenhuma chave ou segredo deve ser versionado.

**Segurança:** toda operação AWS precisa de conta/região confirmadas via STS, IAM mínimo, estimativa de custos, aprovação específica e plano de remoção. O Hub permanece em `kind` inicialmente e os EKS reais favorecem Auto Mode, sujeitos a compatibilidade e PCI. Não duplicar ownership Terraform/Crossplane, versionar segredos nem executar `destroy` indiscriminadamente.

[Arquitetura](architecture.md) · [Decisões](adr/README.md)
