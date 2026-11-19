# 🧪 Roteiro prático da POC

[← Início](../README.md) · [Arquitetura](architecture.md) · [Fundamentos](fundamentos/README.md)

Construção incremental e observável: cada lab terá pré-requisitos, comandos, evidências, falha controlada, diagnóstico, recuperação, custo e limpeza. **A ordem pedagógica difere da sequência de provisionamento de uma entidade completa.**

| Fase | Tema | Evidência |
| :--- | :--- | :--- |
| **00** | Documentação e monorepo | Conceitos e decisões consistentes |
| **01** | Hub `kind` | API Kubernetes e RBAC funcionando |
| **02** | Credenciais AWS temporárias | Identidade/conta/região via STS e orçamento |
| **03** | Crossplane Core, XRD, Composition e Provider AWS | Produto tipado na Kubernetes API |
| **04** | `ObjectStorage` real | Bucket de laboratório, `Ready`, drift e teardown |
| **05** | Entity, Environment, Domain + perfis de Bootstrap no Git | Identidade única, versão do perfil, ativação UAT/PROD, geração idempotente dos manifests obrigatórios |
| **06** | Argo CD central + Domain Bootstrap inicial | PR → manifests obrigatórios de `ObjectStorage` → Crossplane → AWS; isolamento entre Domains; domínio não fica `Ready` sem requisitos do perfil |
| **07** | Foundation Terraform reutilizável | Config central por Entity/Environment, plan/state isolado |
| **08** | Argo Workflows, gates e Terraform Executor | Aprovação e execução auditáveis |
| **09** | Entity Core e EKS Auto Mode incremental | Um EKS por vez, verificando custo, identidade e PCI |
| **10** | Argo CD local + Bootstrap completo por Environment | Namespace, quotas/policies, `Database` e `ObjectStorage` obrigatórios verificados; depois Deployment em destino permitido |
| **11** | Duas contas AWS UAT/PROD | Credenciais e recursos segregados |
| **12** | Observabilidade | Status, logs, métricas, traces, auditoria |
| **13** | Backstage / CLI / AI-MCP | Consumo dos contratos sem API paralela |
| **14** | Extensão opcional: criação de contas | Justificar contrato/controlador de lifecycle se necessário |
| **15** | Segurança, custo e recuperação | Testes de isolamento e cleanup |

**Não implementar `AWSEnvironment` CRD ou Environment Controller próprio como pré-requisito inicial.** Environment e Domain são registros lógicos; somente Capabilities precisam de XRD no primeiro ciclo. O **Domain Bootstrap** deve gerar manifests obrigatórios versionados, por Environment e por perfil; **o Argo CD local**, e não um controller do Domain, aplica namespaces/policies. `Database` do perfil completo será implementado progressivamente e nunca pode ser marcado pronto antes de sua entrega real.

**Segurança:** toda operação AWS precisa de conta/região confirmadas via STS, IAM mínimo, estimativa de custos, aprovação específica e plano de remoção. O Hub permanece em `kind` inicialmente e os EKS reais favorecem Auto Mode, sujeitos a compatibilidade e PCI. Não duplicar ownership Terraform/Crossplane, versionar segredos nem executar `destroy` indiscriminadamente.

[Arquitetura](architecture.md) · [Decisões](adr/README.md)
