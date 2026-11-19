# Registros de Decisão Arquitetural (ADRs)

Baseline da POC pessoal, em **18/11/2026**. As decisões descrevem a **arquitetura-alvo**, não componentes necessariamente implementados.

| ADR | Decisão |
| --- | --- |
| [0001](0001-kubernetes-native-platform-api.md) | API Kubernetes-native com CRDs e controllers |
| [0002](0002-contract-families.md) | Famílias AWSEnvironment, Domain e Capabilities |
| [0003](0003-hub-spoke-topology.md) | Camadas arquiteturais e topologia de gerenciamento |
| [0004](0004-workflow-and-gitops.md) | GitOps, orquestração e mecanismos de execução |
| [0005](0005-long-running-environment-lifecycle.md) | Provisionamento longo e dependências manuais |
| [0006](0006-observability-and-local-operators.md) | Observabilidade transversal e operators locais |
| [0007](0007-security-and-isolation.md) | Segurança, isolamento PCI e proteção contra exclusão |
| [0008](0008-monorepo-and-learning.md) | Monorepo e aprendizado incremental |

## Convenções
Todos os textos explicativos são escritos em **português brasileiro**. Os nomes **Control Plane, Workflow Plane, Observability Plane e Data Plane** identificam as camadas. **Hub-and-Spoke** é apenas o padrão complementar de topologia de gerenciamento. Mudanças de arquitetura devem ser registradas em novos ADRs ou revisões explícitas via PR. Nunca incluir dados sensíveis, contas reais ou referências corporativas.

## Decisões adicionais
- [ADR-0009 — Crossplane e responsabilidades de reconciliação](0009-crossplane-e-reconciliacao.md)
- [ADR-0010 — AWS real e adoção progressiva de ambientes](0010-aws-real-e-ambientes.md)

Esses ADRs atualizam decisões anteriores sobre Crossplane e ordem dos labs, sem alterar a nomenclatura dos planes.
