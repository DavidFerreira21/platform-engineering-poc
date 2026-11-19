# Platform Engineering POC

Laboratório pessoal para estudar uma plataforma **Kubernetes-native**, com API declarativa, automação e ambientes de execução separados.

## Arquitetura-alvo

1. **Experience Layer:** Backstage, CLI, GitOps e AI/MCP.
2. **Control Plane:** Kubernetes API Server, CRDs da plataforma e Platform Controllers.
3. **Workflow Plane:** Argo Workflows, Terraform Executor, Argo CD, aprovações e Crossplane como integração progressiva.
4. **Observability Plane:** OpenTelemetry, métricas, logs, traces, auditoria e alertas.
5. **Data Plane:** ambientes **UAT** e **PROD**, cada um representando uma conta AWS, com Environment Management EKS, EKS PCI, EKS Non-PCI e AWS Runtime.

O Control, Workflow e Observability Planes são responsabilidades lógicas hospedadas no **Platform Management Cluster**. O Data Plane contém os ambientes de execução. O padrão **Hub-and-Spoke** ajuda a explicar a relação entre gerenciamento central e ambientes, mas não substitui os nomes das camadas e não significa, por si só, uma topologia de redes AWS.

O desenho físico-alvo contempla sete clusters, mas os labs começam com simulação local em `kind` e serviços mockados. **A documentação não significa que essa infraestrutura já foi implantada.**

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
- Proibido versionar segredos, kubeconfigs, Terraform state ou identificadores reais. Operações AWS reais exigem autorização explícita.

**Situação atual:** documentação e estrutura inicial para implementação futura.
