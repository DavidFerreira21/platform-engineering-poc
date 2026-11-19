# ADR-0012: Ownership federado, quatro camadas e GitOps central/local

- **Situação:** Aceito como decisão arquitetural da POC; implementação pendente
- **Data:** 19/11/2026
- **Relacionado a:** [ADR-0008](0008-monorepo-and-learning.md), [ADR-0009](0009-crossplane-e-reconciliacao.md), [ADR-0010](0010-aws-real-e-ambientes.md), [ADR-0011](0011-eks-auto-mode.md)

## Contexto

A POC precisa reproduzir o modelo de plataforma com **Foundation de responsabilidade central** e **autonomia de DevOps federados** para os recursos dos Domains e da entidade. A arquitetura já estabelece Kubernetes API central, Crossplane para Capabilities e Terraform para fundação/EKS. Era necessário definir como separar Git, aplicação de manifests, autorização e ownership sem construir muitos repositórios ou instalar Crossplane em cada ambiente.

## Decisão

1. **Manter monorepo:** usar diretórios para representar fronteiras de responsabilidade. Não abrir repositórios por Domain/entidade na POC. Possíveis nomes desses repositórios são apenas analogias organizacionais.
2. **Quatro camadas de ownership:**
   - **Foundation:** conta AWS, landing zone, rede, IAM e baseline; owner central; módulos/configuração por entidade e ambiente; Terraform.
   - **Entity Core Infrastructure:** Management EKS, EKS PCI, EKS Non-PCI e infraestrutura comum realmente necessária; alteração federada sob contratos e aprovações centrais; EKS via Terraform.
   - **Domain Infrastructure:** Capabilities AWS específicas dos Domains; DevOps federado; Crossplane central.
   - **Applications:** código e manifests dos workloads; times locais; Argo CD local.
3. **Centralizar Crossplane:** uma instalação no Platform Management Cluster, responsável pelos recursos AWS definidos em seus XRDs/Compositions/Providers. **Sem cascata de Crossplane** na primeira versão.
4. **Separar GitOps de execução:** Argo CD central sincroniza os CRs de infraestrutura do Git à Kubernetes API do Hub; Crossplane observa os CRs e reconcilia recursos AWS. Argo CD local no Environment Management EKS sincroniza aplicações/configurações nos clusters locais autorizados. Argo Workflows + Terraform executam Foundation/EKS com aprovações; Argo CD não executa Terraform.
5. **Preservar Domain como contrato central:** preferir entrega de namespaces/quotas/políticas via Argo CD local, validando o mecanismo de propagação e status, sem dupla gestão de objetos Kubernetes.
6. **Foundation é central:** um módulo Bootstrap reutilizável e configuração por entidade/ambiente, alinhados ao padrão que se pretende demonstrar. A equipe federada pode propor PR, mas não possuir o lifecycle da conta.
7. **Exclusão e limites:** contas nunca são destruídas por remoção de CR; recursos compartilhados não pertencem ao lifecycle de um Domain; recurso externo tem um único reconciliador; políticas de proteção/retention e gates são obrigatórios para componentes críticos.
8. **Sem MSK inicialmente:** foi apenas um exemplo de Capability compartilhada futura, não uma dependência de implementação.

## Alternativas consideradas

| Alternativa | Resultado |
| :--- | :--- |
| Repositório por Domain, entidade e Foundation | Desnecessário para o laboratório; aumenta manutenção e dificulta visão integrada |
| Crossplane central e Crossplanes locais em cascata | Mais autonomia de reconciliação, mas custo adicional de Providers, contratos, credenciais, versões e diagnóstico; adiado até existir necessidade |
| Um Argo CD central gerenciando todos os Deployments | Aumenta acoplamento e escopo administrativo do Hub; preferido GitOps local por ambiente |
| Crossplane central gerenciando também todas as aplicações | Mistura ownership de infraestrutura AWS com lifecycle local de workloads; não adotado |
| Foundation sob ownership federado | Contradiz o controle central de contas, landing zones e baseline de segurança; não adotado |

## Consequências

- **Positivas:** autonomia por Domain/ambiente, governança central de contratos, separação de motores, menor complexidade da POC e evolução posterior para vários repositórios sem redesenhar a API.
- **Trade-offs:** o Hub torna-se dependência para reconciliação de Capabilities AWS; credenciais e ProviderConfigs precisam de isolamento forte; a distribuição de configurações para Argo CD local e agregação de status exige integração explícita.
- **Atenção:** CODEOWNERS e pastas Git não são isolamento de runtime. Aplicar RBAC, políticas e IAM separados. Um Argo CD local com acesso PCI/Non-PCI necessita validação de fronteira administrativa.

## Plano de validação

1. Demonstrar duas áreas de Domain no monorepo com revisão/autorização distintas.
2. Sincronizar um `ObjectStorage` pela aplicação central do Argo CD para a Kubernetes API, verificando `Ready` no Crossplane e o recurso AWS em conta autorizada.
3. Simular tentativa indevida de consumir credencial/recurso de outra área e comprovar negativa em política/RBAC/IAM.
4. Executar Terraform para um componente de Core com `plan`, gate, aprovação e limpeza segura.
5. Instalar GitOps local quando existir Management EKS e demonstrar que um workload é reconciliado localmente sem Crossplane.
6. Confirmar que exclusão de Domain não elimina Foundation, EKS ou recursos compartilhados.

**Documento de referência:** [Arquitetura consolidada](../architecture.md).
