# ADR-0003: Camadas lógicas e dois ambientes de execução

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
A POC precisa representar gerenciamento central e operação por ambiente sem referências a empresas, países ou entidades.

## Decisão
Adotar as camadas Experience Layer, Control Plane, Workflow Plane, Observability Plane e Data Plane. O Platform Management Cluster hospeda Control, Workflow e Observability, como divisões lógicas. No Data Plane, UAT e PROD são contas AWS distintas, cada qual com Environment Management EKS, EKS PCI, EKS Non-PCI e AWS Runtime. Hub-and-Spoke descreve a topologia de gerenciamento, sem substituir o nome das camadas.

## Consequências e pontos de atenção
Topologia física-alvo de sete clusters; implementação incremental com kind/mocks. O padrão não determina Transit Gateway nem outra arquitetura de rede.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.

## Complemento posterior — ADR-0011
Os EKS reais dos Data Planes (management, PCI e Non-PCI) adotarão inicialmente **Amazon EKS Auto Mode**. O Hub local segue em `kind`. Consultar [ADR-0011](0011-eks-auto-mode.md).
