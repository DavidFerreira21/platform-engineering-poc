# ☁️ Data Plane

O Data Plane é composto de **Entities**, com Environments UAT/PROD e duas áreas de infraestrutura: **Core Infrastructure** (compartilhada) e **Domains** de negócio/aplicação. Ambos pertencem ao Data Plane.

**Environment é a conta AWS inteira**, não sinônimo de Core. A divisão padrão de ownership é **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**. Uma Entity pode ter diversos Domains; cada Domain é identidade única que possui Capabilities e aplicações em UAT e/ou PROD. **A criação de um Domain utiliza Domain Bootstrap**: o perfil versionado gera sua estrutura e os manifests iniciais de Capabilities/namespace/policies para os Environments selecionados; Crossplane central e Argo CD local reconciliam os recursos.

## Management EKS e AWS Runtime (topologia-alvo)

Em cada conta UAT/PROD, a topologia prevê **Management EKS**, **EKS PCI** e **EKS Non-PCI**. O Management EKS será o ponto local de **Argo CD**, **Operators ainda a estudar** e **OpenTelemetry Collector/gateway** para encaminhar logs, métricas e traces. A visão centralizada ficará no Observability Plane do Hub; não pressupor uma instalação de Loki em cada ambiente.

O **AWS Runtime** engloba Core Infrastructure (recursos compartilhados) e Domain Infrastructure (recursos de cada domínio). **Terraform** cria Foundation e Entity Core como rede, dependências e EKS; **Crossplane** reconcilia Capabilities AWS de Domain, e pode operar Capabilities comuns quando o ownership for explícito. Um recurso não terá Terraform e Crossplane como owners simultâneos.

**Domain Applications:** GitHub Actions será usado para CI/CD **em outro repositório de aplicação de teste, ainda não fornecido**; Argo CD local aplica os manifests de workloads em EKS PCI ou Não-PCI autorizado.

**Domain Bootstrap** continua registrado como hipótese, mas a implementação foi **adiada e está em reavaliação**.

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
