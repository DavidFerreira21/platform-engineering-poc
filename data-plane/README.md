# ☁️ Data Plane

O Data Plane é composto de **Entities**, com Environments UAT/PROD e duas áreas de infraestrutura: **Core Infrastructure** (compartilhada) e **Domains** de negócio/aplicação. Ambos pertencem ao Data Plane.

**Environment é a conta AWS inteira**, não sinônimo de Core. A divisão padrão de ownership é **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**. Uma Entity pode ter diversos Domains; cada Domain é identidade única que possui Capabilities e aplicações em UAT e/ou PROD.

## Management EKS e AWS Runtime (topologia-alvo)

Em cada conta UAT/PROD, a topologia prevê **Management EKS**, **EKS PCI** e **EKS Non-PCI**. O Management EKS hospeda o **Argo CD local**, Operators a escolher e os backends de logs e traces do ambiente. Cada cluster coleta sua própria telemetria e envia métricas direto ao **AMP da conta**. Logs e traces seguem ao Gateway do Management EKS e são armazenados por Loki/Tempo no S3 da conta. O perfil de telemetria de PCI depende da validação de isolamento. [Fluxo de observabilidade](../docs/flows/observability.md)

O **AWS Runtime** engloba Core Infrastructure (recursos compartilhados) e Domain Infrastructure (recursos de cada domínio). **Terraform** cria Foundation e Entity Core como rede, dependências e EKS; **Crossplane** reconcilia Capabilities AWS de Domain, e pode operar Capabilities comuns quando o ownership for explícito. Um recurso não terá Terraform e Crossplane como owners simultâneos.

**Domain Applications:** código, GitHub Actions **e manifests Kubernetes da aplicação** ficarão **juntos no repositório próprio da squad, previsto para uma etapa posterior**; Argo CD local aplica esses manifests em EKS PCI/Não PCI autorizado. O **Application Team** é owner desse repositório.

## Personas e consumo da infraestrutura

- **Platform Team:** responsável pelos mecanismos da plataforma, contratos, Foundation e Entity Core (inclusive EKS compartilhado), aprovações/guardrails e Golden Paths.
- **Federated DevOps Team:** ponte entre plataforma e squad; responsável por propor e gerenciar Capabilities de **Domain Infrastructure** (Database, ObjectStorage, Messaging e upgrades) via manifests no monorepo.
- **Application Team:** código e ciclo de **Domain Applications** num repositório próprio; consome Golden Paths/Starter Kits sem precisar conhecer os motores internos.

**Aplicação de demonstração.** O smoke test HTTP do Non-PCI entra na Etapa 18. Nas Etapas 22 a 25, a squad implementa o registro de recebimento de encomendas com FastAPI, Jinja2, Aurora PostgreSQL e evidência JSON no S3. O Golden Path reutiliza essa aplicação nas Etapas 34 e 35 com a solicitação federada de Capabilities, CI em GitHub Actions, Argo CD local e telemetria. [ADR-0007](../docs/adr/0007-personas-and-developer-experience.md)

## Organização proposta

```text
data-plane/
└── entity-a/
    ├── core-infrastructure/
    │   ├── uat/                 # 3 EKS e infraestrutura comum
    │   └── prod/
    └── domains/
        └── rastreamento-encomendas/
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
