# ☁️ Data Plane

O Data Plane é composto de **Entities**, com Environments UAT/PROD e duas áreas de infraestrutura: **Core Infrastructure** (compartilhada) e **Domains** de negócio/aplicação. Ambos pertencem ao Data Plane.

**Environment é a conta AWS inteira**, não sinônimo de Core. Uma Entity pode ter diversos Domains; cada Domain é identidade única que possui Capabilities e aplicações em UAT e/ou PROD. **A ativação por Environment requer Domain Bootstrap**: um perfil versionado define Capabilities/namespace/policies obrigatórios, gerados como manifests Git e reconciliados por Crossplane central e Argo CD local.

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
            │   ├── infrastructure/
            │   └── applications/
            └── prod/
                ├── infrastructure/
                └── applications/
```

Por enquanto, os diretórios históricos [UAT](uat/README.md) e [PROD](prod/README.md) são **marcadores documentais**, sem recursos implantados. Migraremos para a organização por entidade durante os labs, evitando duplicação de configurações.

Três clusters EKS por Environment na arquitetura-alvo: Management EKS com Argo CD local, EKS PCI e EKS Non-PCI; Auto Mode é preferência sob validação de custo/segurança.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Estrutura](../docs/repository-structure.md)
