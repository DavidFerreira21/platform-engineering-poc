# ⚙️ Workflow Plane

O Workflow Plane inclui o **Domain Bootstrap**, que valida o perfil versionado e gera a estrutura e os manifests iniciais de cada novo Domain, para os Environments selecionados, em PR; ele não provisiona recursos diretamente e não concorre com Argo CD/Crossplane. A verificação da prontidão deve agregar evidências de ambos. 

O Workflow Plane mantém processos, GitOps central, aprovação e execução, **sem Crossplane**. **Argo Workflows é instalado logo após Argo CD ([Lab 02B](../labs/02b-argo-workflows.md))**, antes de Crossplane; o Lab 08 aprofunda integração com Terraform Executor e aprovações.

| Componente | Papel |
| :--- | :--- |
| **Argo CD central** | Sincroniza CRs Kubernetes de Capability na API do Hub |
| **Argo Workflows** | Instalado via GitOps no Hub no Lab 02B; começa com workflow local sem efeito em AWS e evolui para passos, gates, espera e auditoria de Foundation/Core no Lab 08 |
| **Terraform Executor** | `plan`, aprovação e `apply` de Foundation/EKS, com state seguro |

`Entity`, `Environment` e `Domain` são registros lógicos no Git. Não exigem Environment Controller próprio ou XRD de Domain nesta fase. A criação de contas será reconsiderada no futuro quando houver necessidade de lifecycle automatizado. O **Argo CD local** roda em cada Management EKS autorizado, junto do Data Plane.

## Aplicações e execuções declarativas (planejadas)

- **TerraformRun (proposto):** o controller fica no Control Plane; **Argo Workflows** executa o Workflow criado após validação do CRD, com `plan → gate → apply`, status e state seguro. Lab 08B complementa o Executor do Lab 08.
- **Domain Applications:** GitHub Actions fará o **CI/CD em um repositório de aplicação de teste separado**, ainda a fornecer. Build/test/scan/publicação → atualização GitOps revisada → Argo CD **local** no Management EKS faz deploy nos clusters autorizados. Não adicionar código ou workflow CI da aplicação neste monorepo.
- **Domain Bootstrap:** conceito preservado no ADR-0005, mas **adiado e em reavaliação**. Não torná-lo pré-requisito dos labs.

[ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

Os labs do Workflow Plane usam o Hub local. A integração com Terraform Executor segue no roteiro do Lab 08.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)
