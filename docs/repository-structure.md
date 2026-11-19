# Estrutura e responsabilidades do monorepo

A estrutura do repositório evidencia **Control Plane, Workflow Plane, Observability Plane e Data Plane**, com UAT e PROD como ambientes distintos. Os diretórios existentes com README são marcadores de escopo, não componentes implantados.

```text
platform-engineering-poc/
├── AGENTS.md
├── README.md
├── control-plane/
│   ├── crds/                    # AWSEnvironment, Domain e Capabilities
│   └── controllers/             # reconciliação, políticas e status
├── workflow-plane/
│   ├── argo-workflows/          # execução e aprovações
│   ├── terraform-executor/      # plan/apply controlado
│   ├── argocd/                  # GitOps global
│   └── crossplane/              # integração progressiva
├── observability-plane/         # OTel, métricas, logs, traces, auditoria
├── platform/hub/                # bootstrap do Platform Management Cluster
├── data-plane/
│   ├── uat/
│   │   ├── management-eks/      # Argo CD, operators, OTel local
│   │   ├── eks-pci/             # workloads PCI
│   │   ├── eks-non-pci/         # workloads não PCI
│   │   └── aws-runtime/         # serviços AWS da conta UAT
│   └── prod/
│       ├── management-eks/
│       ├── eks-pci/
│       ├── eks-non-pci/
│       └── aws-runtime/
├── docs/
│   ├── adr/
│   ├── architecture.md
│   ├── repository-structure.md
│   └── learning-roadmap.md
├── labs/
└── tests/
```

## Critérios de organização
- Código de CRDs e controllers pertence a `control-plane/`; instalação do Hub pertence a `platform/hub/`.
- Definições compartilhadas de workflows, Terraform runners, Argo CD e Crossplane pertencem a `workflow-plane/`.
- Telemetria compartilhada fica em `observability-plane/`; overlays específicos do ambiente ficam no respectivo Spoke.
- Configurações e recursos por conta ficam em `data-plane/uat/` ou `data-plane/prod/`. Não duplicar manifests compartilhados.
- **Experience Layer** descreve canais de consumo, não exige pasta separada para cada cliente enquanto não houver código.
- Hub-and-Spoke permanece conceito explicativo na documentação, sem substituir as camadas nem impor uma divisão adicional no Git.
- Criar subdiretórios de implementação apenas quando houver conteúdo real.
- Toda a documentação, comentários didáticos e guias devem usar **pt-BR**.
- Não usar uma API FastAPI própria como requisito; a interface principal é a Kubernetes API.

## Critérios de revisão
Isolamento UAT/PROD, contratos versionados, referências válidas, ownership do `spec` por GitOps, status reconciliado, privilégio mínimo, limites PCI, ausência de segredos, testes e instruções de limpeza.
