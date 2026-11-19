# ADR-0009: Crossplane no Control Plane e reconciliadores especializados

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026
- **Atualiza:** ADR-0002 e ADR-0004

## Contexto

Nossa API oficial é a Kubernetes API. A arquitetura precisa distinguir contratos públicos de mecanismos de implementação, evitando escrever controllers próprios onde o Crossplane já oferece reconciliação declarativa.

## Decisão

1. **Kubernetes API Server é a API única.** Recebe e persiste os Custom Resources; Crossplane e o controller próprio são reconciliadores que observam a API, não endpoints paralelos.
2. **CRD nativo AWSEnvironment:** reconciliado por **Environment Controller próprio** (Go/controller-runtime previsto). O controller coordenará processos longos via Argo Workflows e Terraform, com status, idempotência, gates e proteção contra exclusão.
3. **Domain:** **XRD Crossplane + Composition**. Inicialmente modela namespace e configurações de cluster, como quota, limites e políticas. Não implementar Domain Controller customizado sem necessidade comprovada.
4. **Capabilities:** XRDs independentes para `Database`, `ObjectStorage`, `ComputeInstance`, `Messaging`; Crossplane Compositions/Providers cuidam da entrega. Capability é uma *família lógica*, não um CRD Claim genérico.
5. Um **XRD (CompositeResourceDefinition)** é uma definição do Crossplane que gera a exposição de um recurso composto por um **CRD Kubernetes**; o usuário consome a instância do recurso pela Kubernetes API.
6. **Crossplane Core, XRDs, Compositions e Providers** pertencem logicamente ao **Control Plane**. **Workflow Plane** contém Argo Workflows, Terraform Executor, Argo CD e gates. Crossplane reconcilia continuamente e **não é um motor de workflow procedural**.
7. Um recurso externo possui **um só dono de reconciliação**. Evitar gestão concorrente Terraform/Crossplane sobre o mesmo banco, bucket ou outro recurso.

## Consequências e validações

- Validar schemas, condições de status, exclusão, referências `environmentRef`/`domainRef` e versionamento dos XRDs.
- A Composition de Domain no Hub **não gerencia automaticamente um EKS remoto**. Testar provider apropriado ou delegação GitOps, credenciais, RBAC e limites PCI/Non-PCI.
- Demonstrar em lab que AWSEnvironment, Domain e Database estão acessíveis pelo **mesmo API Server**, cada qual reconciliado pelo responsável definido.
- Priorizar Crossplane no roadmap antes dos labs de Domain/Capabilities; não desenvolver Domain Controller próprio inicialmente.
