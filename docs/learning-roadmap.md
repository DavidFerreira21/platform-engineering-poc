# 🧪 Roteiro prático da POC

[← Início](../README.md) · [Arquitetura](architecture.md) · [Fundamentos](fundamentos/README.md)

Construção incremental e observável: cada lab terá pré-requisitos, comandos, evidências, falha controlada, diagnóstico, recuperação, custo e limpeza. **Primeira implementação: Lab 01 (kind, namespaces, rede local e Traefik)**, seguido por **Lab 02 (Argo CD central)**, **Lab 02B (Argo Workflows instalado via GitOps)** e **Lab 03 (Crossplane instalado via GitOps)**, depois `ObjectStorage` e **`Database` com Aurora PostgreSQL**. A conta inicial será apenas UAT. A decisão de autenticação AWS fica para uma fase posterior; não é necessária para criar o kind. **A ordem pedagógica difere da sequência de provisionamento de uma entidade completa.**

| Fase | Tema | Evidência |
| :--- | :--- | :--- |
| **00** | Documentação e monorepo | Conceitos e decisões consistentes |
| **01** | Hub `kind` + Traefik | API, CoreDNS, namespaces, mapeamento de portas, Ingress HTTP e verificações locais |
| **02** | [Argo CD central (GitOps)](../labs/02-argocd-central.md) | Bootstrap do Argo CD, acesso local, Git como fonte da verdade e aplicações da plataforma |
| **02B** | Argo Workflows no Hub por GitOps | Argo CD instala Argo Workflows e CRDs; executar um Workflow local inofensivo, observar Pods/logs/status e validar reexecução e limpeza, sem AWS ou Terraform |
| **03** | [Crossplane Core por GitOps](../labs/03-crossplane-core.md) | Argo CD instala o Crossplane; observar CRDs, XRDs, Compositions e status sem AWS |
| **03B** | UAT: registro de conta e autenticação AWS | Configurar arquivo de Environment e método de autenticação a definir; validar STS **antes de qualquer operação** |
| **04** | `ObjectStorage` real | XRD/Composition/Provider AWS, bucket de laboratório, `Ready`, drift e teardown |
| **04B** | `Database` com Amazon Aurora PostgreSQL | Contrato de Database e cluster Aurora PostgreSQL em UAT, sujeito a custo, rede, IAM, plano e teardown autorizados |
| **05** | Entity, Environment e Domain (registros lógicos) | Schema e validação de identidade/ownership/destino, **sem depender de Domain Bootstrap** |
| **06** | Argo CD central + primeira Capability de Domain | GitOps de `ObjectStorage` → Crossplane → AWS, com destino autorizado e isolamento |
| **07** | Foundation Terraform reutilizável | Config central por Entity/Environment, plan/state isolado |
| **08** | Integração Argo Workflows + gates + Terraform Executor | Reutilizar o Lab 02B; validar plan/apply, state seguro, aprovação e auditoria |
| **08B** | `TerraformRun` CRD + Controller (**propostos**) | Validação, status, idempotência, disparo único do Argo Workflow e gates (primeiro com executor simulado) |
| **09** | Entity Core e EKS Auto Mode incremental | Um EKS por vez, verificando custo, identidade e PCI |
| **10** | Argo CD local + Operators / gateway de telemetria | Management EKS faz GitOps para EKS PCI/Non-PCI autorizados; Operators a escolher e OTel Collector/gateway |
| **10B** | CI/CD de Domain Applications (repo da squad) | GitHub Actions faz build/test/scan/publica; **manifests Kubernetes ficam no mesmo repo da aplicação** e Argo CD local realiza deploy GitOps |
| **10C** | Ownership por persona e permissões | Simular Platform Team, Federated DevOps e Application Team; PRs positivos/negativos por caminho, proteção e destinos GitOps |
| **13A** | Java Service Golden Path / Starter Kit | Gerar repositório de aplicação Java com build, testes, Dockerfile, GitHub Actions, manifests K8s, telemetria e policies; desenvolver sem conhecer internals da plataforma |
| **11** | Duas contas AWS UAT/PROD | Credenciais e recursos segregados |
| **12** | Observabilidade OpenTelemetry / Grafana / Loki / Tempo | Gateway nos spokes, logs/traces centralizados, métricas, alertas e segregação de acesso |
| **13** | Backstage / CLI / AI-MCP | Consumir Capabilities e Golden Paths por canais independentes, sem API paralela |
| **14** | Extensão opcional: criação de contas | Justificar contrato/controlador de lifecycle se necessário |
| **14B** | Reavaliar Domain Bootstrap (opcional) | ADR preservado; decidir posteriormente se este gerador terá utilidade, **sem implementá-lo agora** |
| **15** | Segurança, custo e recuperação | Testes de isolamento e cleanup |

**Ordem dos componentes no Hub:** `kind` e Traefik → Argo CD central → **Argo Workflows via Argo CD** → Crossplane Core via Argo CD. O Lab 02B instala e observa o motor com um fluxo de teste local; integração com Terraform, gates humanos e permissões AWS fica para o Lab 08. Não confundir instalar o workflow controller com habilitar execução de infraestrutura real.

**Não implementar `AWSEnvironment` CRD ou Environment Controller próprio como pré-requisito inicial.** Environment e Domain são registros lógicos; somente Capabilities precisam de XRD no primeiro ciclo. **Domain Bootstrap está em reavaliação e adiado**, não é pré-requisito dos labs. **O Argo CD local** aplica namespaces/policies e workloads Kubernetes. `Database` seguirá `ObjectStorage`, com Aurora PostgreSQL, custos e teardown aprovados. `TerraformRun` e CI/CD externo são propostas, **não componentes implantados**.

**Três personas e Golden Paths:** o **Platform Team** mantém a plataforma e sua infraestrutura compartilhada; o **Federated DevOps Team** propõe e mantém os manifests de **Domain Infrastructure** no monorepo; o **Application Team** desenvolve e opera código/CI/manifests Kubernetes em **repositório separado**. O primeiro Starter Kit planejado é um **Java Service**. Ver [ADR-0007](adr/0007-personas-and-developer-experience.md). O Lab 10C valida fronteiras de ownership, e o 13A valida a experiência de desenvolvimento. Não criar repositório de aplicação sem ele ser informado.

**Conta UAT:** registrar o destino em `foundation/config/entity-a/uat/environment.example.yaml` (template sem dados reais). Conta, região, credenciais e método de autenticação ainda precisam ser definidos e validados no início dos labs AWS. Nenhuma chave ou segredo deve ser versionado.

**Segurança:** toda operação AWS precisa de conta/região confirmadas via STS, IAM mínimo, estimativa de custos, aprovação específica e plano de remoção. O Hub permanece em `kind` inicialmente e os EKS reais favorecem Auto Mode, sujeitos a compatibilidade e PCI. Não duplicar ownership Terraform/Crossplane, versionar segredos nem executar `destroy` indiscriminadamente.

[Arquitetura](architecture.md) · [Decisões](adr/README.md)
