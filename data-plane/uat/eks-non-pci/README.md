# UAT — eks-non-pci

Cluster EKS destinado a workloads não PCI, incluindo operators/agentes locais quando necessários.

**Situação:** componente previsto na arquitetura-alvo, não provisionado. Qualquer alteração real na AWS exige autorização explícita.

## Tipo de cluster
**Amazon EKS Auto Mode** é o padrão de provisionamento deste cluster real. Conferir compatibilidade de add-ons, operators, identidade, rede e isolamento antes da implantação; o uso não implica que o cluster já existe. Consulte [ADR-0011](../../../docs/adr/0011-eks-auto-mode.md).
