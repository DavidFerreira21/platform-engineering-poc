# ADR-0011: Amazon EKS Auto Mode como padrão dos clusters nos Data Planes

- **Situação:** Aceito como padrão inicial da arquitetura-alvo, sujeito à validação técnica por tipo de cluster
- **Data:** 18/11/2026
- **Relacionados:** ADR-0003, ADR-0010

## Contexto
A POC provisionará recursos AWS reais, mas a intenção é aprender contratos, automação, GitOps e lifecycle sem desviar o foco para operação manual de nós Kubernetes. O Hub pode permanecer em kind; os clusters reais dos Spokes precisam ser fáceis de criar e manter.

## Decisão
1. Adotar **Amazon EKS Auto Mode** por padrão nos clusters de **UAT** e **PROD**: **Environment Management EKS**, **EKS PCI** e **EKS Non-PCI**.
2. O **Platform Management Cluster** global continua inicialmente em **kind local**; não confundir kind com EKS Auto Mode.
3. Os módulos de provisionamento devem habilitar Auto Mode explicitamente e usar os recursos e interfaces adequados para computação, rede, balanceamento e storage gerenciados pela AWS. Customizações usam configurações suportadas, incluindo NodePools/NodeClasses quando aplicável.
4. Não instalar ou gerenciar duplicadamente serviços que o Auto Mode já fornece; conferir a compatibilidade de add-ons/operators e evitar pressupor acesso via SSH/SSM, alteração livre de AMI ou modificações host-level nos nós.
5. Avaliar de forma explícita a compatibilidade do **Management EKS** com Argo CD, operators e gateways; nos EKS PCI/Non-PCI avaliar controles de segurança, agentes, isolamento e requisitos PCI. A escolha **não implica certificação PCI automática**.
6. Provisionar gradualmente, um cluster por vez conforme os labs, usando contas AWS autenticadas, controles de custo e aprovação da execução. Não criar seis EKS por simplesmente adotar o ADR.

## Consequências
- Menor manutenção de infraestrutura de nodes e de componentes AWS integrados, com trade-offs em personalização e acesso ao host.
- Podem existir custos adicionais do EKS Auto Mode além dos custos usuais de EKS e capacidade. Estimar e acompanhar despesas antes de cada lab.
- Caso algum operator, add-on ou política não seja compatível, documentar exceção e avaliar outra modalidade de compute para o cluster específico; não desviar silenciosamente do padrão.
- Validação concreta ocorre no Lab 11, incluindo criação, acesso, Auto Mode, scheduling, identidade, observabilidade e teardown.

## Referências
- [AWS — EKS Auto Mode](https://docs.aws.amazon.com/eks/latest/userguide/automode.html)
- [AWS — Boas práticas EKS Auto Mode](https://docs.aws.amazon.com/eks/latest/best-practices/automode.html)
