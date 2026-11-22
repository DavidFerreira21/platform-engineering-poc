# ☁️ Data Plane

O Data Plane é composto de **Entities**, com Environments UAT/PROD e duas áreas de infraestrutura: **Core Infrastructure** (compartilhada) e **Domains** de negócio/aplicação. Ambos pertencem ao Data Plane.

**Environment é a conta AWS inteira**, não sinônimo de Core. A divisão padrão de ownership é **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**. Uma Entity pode ter diversos Domains; cada Domain é identidade única que possui Capabilities e aplicações em UAT e/ou PROD. **Domain Bootstrap é uma proposta documentada, atualmente adiada/em reavaliação**. Não exigir um gerador para que Capabilities e aplicações sejam entregues pelos reconciliadores Crossplane central e Argo CD local.

## Management EKS e AWS Runtime (topologia-alvo)

Em cada conta UAT/PROD, a topologia prevê **Management EKS**, **EKS PCI** e **EKS Non-PCI**. O Management EKS será o ponto local de **Argo CD**, **Operators ainda a estudar** e **OpenTelemetry Collector/gateway** para encaminhar logs, métricas e traces. A visão centralizada ficará no Observability Plane do Hub; não pressupor uma instalação de Loki em cada ambiente.

O **AWS Runtime** engloba Core Infrastructure (recursos compartilhados) e Domain Infrastructure (recursos de cada domínio). **Terraform** cria Foundation e Entity Core como rede, dependências e EKS; **Crossplane** reconcilia Capabilities AWS de Domain, e pode operar Capabilities comuns quando o ownership for explícito. Um recurso não terá Terraform e Crossplane como owners simultâneos.

**Domain Applications:** código, GitHub Actions **e manifests Kubernetes da aplicação** ficarão **juntos no repositório separado da squad, ainda não fornecido**; Argo CD local aplica esses manifests em EKS PCI/Não PCI autorizado. O **Application Team** é owner desse repositório.

**Domain Bootstrap** continua registrado como hipótese, mas a implementação foi **adiada e está em reavaliação**.

## Personas e consumo da infraestrutura

- **Platform Team:** responsável pelos mecanismos da plataforma, contratos, Foundation e Entity Core (inclusive EKS compartilhado), aprovações/guardrails e Golden Paths.
- **Federated DevOps Team:** ponte entre plataforma e squad; responsável por propor e gerenciar Capabilities de **Domain Infrastructure** (Database, ObjectStorage, Messaging e upgrades) via manifests no monorepo.
- **Application Team:** código e ciclo de **Domain Applications** num repositório próprio; consome Golden Paths/Starter Kits sem precisar conhecer os motores internos.

**App mínima de demonstração (preferência Python + Flask):** primeiro executar **smoke test no EKS após Argo CD local (Lab 10B)**; após observabilidade, segurança e construção da **Developer Experience (Lab 15)**, reutilizá-la no **Golden Path / integração final (Lab 16)** com GitHub Actions e telemetria. **Não há necessidade de desenvolver aplicação complexa ou gerador completo de Starter Kits**. O conceito de Golden Path continua válido para evolução futura. A criação de Database/S3 permanece demanda do DevOps Federado. [ADR-0007](../docs/adr/0007-personas-and-developer-experience.md).

## Organização proposta

```text
data-plane/
└── entity-a/
    ├── core-infrastructure/
    │   ├── uat/                 # 3 EKS e infraestrutura comum
    │   └── prod/
    └── domains/
        └── payments/
            ├── domain.yaml      # owners e políticas comuns
            ├── uat/
            │   └── infrastructure/      # DevOps Federado
            └── prod/
                └── infrastructure/      # DevOps Federado
```

O repositório da aplicação, **separado deste monorepo**, concentrará `src/`, `tests/`, `Dockerfile`, `.github/workflows/` e `k8s/` (base/overlays UAT e PROD), sob gestão do Application Team.

Por enquanto, os diretórios históricos [UAT](uat/README.md) e [PROD](prod/README.md) são **marcadores documentais**, sem recursos implantados. Migraremos para a organização por entidade durante os labs, evitando duplicação de configurações.

Três clusters EKS por Environment na arquitetura-alvo: Management EKS com Argo CD local, EKS PCI e EKS Non-PCI; Auto Mode é preferência sob validação de custo/segurança.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Estrutura](../docs/repository-structure.md)
