# ✅ Testes

A suíte atual cobre contratos Kubernetes, schemas, controller TerraformStack, runner, políticas, configuração Terraform e documentação. Os aceites reais de Crossplane e integrações AWS seguem os respectivos labs.

- **Padrão:** testes locais, mocks e verificações estáticas sem recursos AWS reais.
- **Integração AWS:** habilitada apenas quando houver configuração e autorização específicas.
- **Qualidade:** validar links Markdown, manifests e caminhos do repositório ao evoluir os labs.

## Regressão do TerraformStack Controller

A suíte Go em `control-plane/terraformstack/controller/internal/controller` e `internal/execution/argo` usa cliente fake e relógio determinístico; o envtest usa API Kubernetes e etcd reais. Não acessa AWS.

```bash
cd control-plane/terraformstack/controller
go mod tidy -diff
go test -race ./...
```

A CI executa testes unitários e envtest em PRs e pushes na main. O envtest verifica a disputa concorrente pelo `resourceVersion`. Os testes unitários cobrem seleção de integração, autorização opcional, exclusividade e perda de Workflow. `internal/observability` testa transições, drift pendente, remoção de séries e Events sanitizados. O runner também passa por testes de resolução de branch, tag, SHA e vínculo de destino em um repositório descartável.

```bash
docker build -t terraformstack-runner:local workflow-plane/terraform-executor
python3 tests/test_terraformstack_resolution.py
```

A [Etapa 06](../labs/06-terraformstack-simulator.md) verifica o efeito do Workflow, do OPA, do claim e da ferramenta de IaC no Hub.

---

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro prático](../docs/learning-roadmap.md)
