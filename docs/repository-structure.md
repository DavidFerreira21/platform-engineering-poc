# Estrutura e responsabilidades do monorepo

A estrutura do repositório evidencia **Control Plane, Workflow Plane, Observability Plane e Data Plane**, com UAT e PROD como ambientes distintos. Os diretórios existentes com README são marcadores de escopo, não componentes implantados.

```text
platform-engineering-poc/
├── AGENTS.md
├── README.md
├── control-plane/
│   ├── environment/             # CRD AWSEnvironment + controller próprio
│   └── crossplane/              # Core, XRDs, Compositions, Providers
│       ├── domains/             # XRD Domain
│       └── capabilities/        # XRDs Database, Storage, Compute, Messaging
├── workflow-plane/
│   ├── argo-workflows/          # execução e aprovações
│   ├── terraform-executor/      # plan/apply controlado
│   ├── argocd/                  # GitOps global
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
- Código do controller AWSEnvironment e XRDs/Compositions/Providers Crossplane pertencem ao `control-plane/`; instalação do Hub pertence a `platform/hub/`.
- Argo Workflows, Terraform Executor, gates e Argo CD pertencem a `workflow-plane/`; Crossplane pertence ao **Control Plane**.
- Telemetria compartilhada fica em `observability-plane/`; overlays específicos do ambiente ficam no respectivo Spoke.
- Configurações e recursos por conta ficam em `data-plane/uat/` ou `data-plane/prod/`. Não duplicar manifests compartilhados.
- **Experience Layer** descreve canais de consumo, não exige pasta separada para cada cliente enquanto não houver código.
- Hub-and-Spoke permanece conceito explicativo na documentação, sem substituir as camadas nem impor uma divisão adicional no Git.
- Criar subdiretórios de implementação apenas quando houver conteúdo real.
- Toda a documentação, comentários didáticos e guias devem usar **pt-BR**.
- Não usar uma API FastAPI própria como requisito; a interface principal é a Kubernetes API.

## Critérios de revisão
Isolamento UAT/PROD, contratos versionados, referências válidas, ownership do `spec` por GitOps, status reconciliado, privilégio mínimo, limites PCI, ausência de segredos, testes e instruções de limpeza.

## AWS real e separação de fases

Os diretórios `data-plane/uat` e `data-plane/prod` representarão contas AWS **reais** distintas quando disponibilizadas. No início, usar uma única conta preexistente autorizada e evoluir para duas contas. Somente depois implementar a criação automática de contas pelo Environment Controller. Templates reutilizáveis ficam na camada proprietária, não duplicados nos overlays UAT/PROD. O bootstrap `kind` do Hub não implica que os recursos AWS sejam simulados.

## Provisionamento EKS
Os seis clusters EKS previstos nos Spokes serão modelados com **Amazon EKS Auto Mode** por padrão, conforme [ADR-0011](adr/0011-eks-auto-mode.md). Configurações compartilhadas e módulos reutilizáveis ficam em sua camada proprietária; overlays UAT/PROD expressam apenas diferenças de ambiente. O `platform/hub/` local utiliza `kind`.

## Trilha de fundamentos

O diretório `docs/fundamentos/` organiza aulas conceituais, separadas de `labs/`: a primeira explica **por que/para quem/como pensar**; o segundo demonstra **como implementar, observar, falhar e recuperar**. Aulas numeradas com links anterior/próxima devem manter pt-BR, exercícios e exemplos da POC. O livro é referência de temas, não texto a reproduzir. [Índice](fundamentos/README.md).
