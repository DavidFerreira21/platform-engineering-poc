# Platform Engineering POC

Laboratório pessoal para estudar uma plataforma **Kubernetes-native**, com API declarativa, automação e ambientes de execução separados.

## Arquitetura-alvo

1. **Experience Layer:** Backstage, CLI, GitOps e AI/MCP.
2. **Control Plane:** Kubernetes API Server, CRD nativo `AWSEnvironment` com controller próprio, Crossplane Core com XRDs/Compositions/Providers para `Domain` e Capabilities.
3. **Workflow Plane:** Argo Workflows, Terraform Executor, Argo CD e aprovações. Crossplane pertence ao **Control Plane**.
4. **Observability Plane:** OpenTelemetry, métricas, logs, traces, auditoria e alertas.
5. **Data Plane:** ambientes **UAT** e **PROD**, cada um representando uma conta AWS, com Environment Management EKS, EKS PCI, EKS Non-PCI e AWS Runtime.

O Control, Workflow e Observability Planes são responsabilidades lógicas hospedadas no **Platform Management Cluster**. O Data Plane contém os ambientes de execução. O padrão **Hub-and-Spoke** ajuda a explicar a relação entre gerenciamento central e ambientes, mas não substitui os nomes das camadas e não significa, por si só, uma topologia de redes AWS.

A topologia-alvo contempla sete clusters, mas o **Hub pode começar localmente em `kind`**. O provisionamento dos labs utilizará **AWS real**, inicialmente em conta existente e autenticada, sem provisionar todos os clusters de uma vez. **A documentação não significa que essa infraestrutura já foi implantada.**

## Documentação
- [Arquitetura](docs/architecture.md)
- [Decisões arquiteturais](docs/adr/README.md)
- [Estrutura do repositório](docs/repository-structure.md)
- [Roteiro de aprendizado](docs/learning-roadmap.md)
- [Instruções para agentes](AGENTS.md)

## Regras fundamentais
- Documentação, ADRs, orientações e explicações em **português brasileiro**; preservar nomes técnicos, comandos e campos de API.
- Nenhuma associação a empresa, país ou entidade: apenas UAT e PROD.
- Git é a fonte da verdade do estado desejado; Kubernetes API é a API da plataforma e expõe o estado operacional.
- Três **famílias** de contratos: `AWSEnvironment`, `Domain` e Capabilities especializadas (`Database`, `ObjectStorage` etc.). Não adotar `CapabilityClaim` genérico inicialmente.
- Um `AWSEnvironment` corresponde a uma conta AWS; pode haver vários Domains por ambiente.
- Provisionamento combina automação, validações técnicas e aprovações humanas auditáveis.
- Proibido versionar segredos, kubeconfigs, Terraform state ou identificadores reais. AWS real faz parte do laboratório: confirmar identidade/conta/região, permissões, custos e plano antes de cada execução; exigir aprovação explícita por operação e limpeza controlada.

**Situação atual:** documentação e estrutura inicial para implementação futura.

## Fases de integração AWS
1. **Conta existente autenticada:** acesso AWS real para criar e observar recursos com Crossplane e Terraform, sem criar conta.
2. **UAT e PROD:** separar a implantação em duas contas AWS preexistentes, com credenciais e estados isolados.
3. **Criação de contas:** evoluir o `AWSEnvironment` para provisionamento/adoção automatizados, com Argo Workflows e aprovações.

Decisões: [ADR-0009](docs/adr/0009-crossplane-e-reconciliacao.md) e [ADR-0010](docs/adr/0010-aws-real-e-ambientes.md).
