# ✅ Testes

Os testes cobrirão contratos Kubernetes, validação de schemas, controllers, workflows, segurança e integrações.

- **Padrão:** testes locais, mocks e verificações estáticas sem recursos AWS reais.
- **Integração AWS:** habilitada apenas quando houver configuração e autorização específicas.
- **Qualidade:** validar links Markdown, manifests e caminhos do repositório ao evoluir os labs.

## Regressão do TerraformStack Controller

A suíte Go em `control-plane/terraformstack/controller/internal/controller` e `internal/execution/argo` usa cliente fake e relógio determinístico; o envtest usa API Kubernetes e etcd reais. Não acessa AWS.

```bash
cd control-plane/terraformstack/controller
go mod tidy
go test -race ./...
```

O CI executa os testes unitários e o envtest em cada PR e push na main. A validação do runner e OPA é complementar. Os testes não substituem o Simulator do Lab 06: a atomicidade do claim do runner, a persistência real da API e a execução OPA/Terraform precisam de validação de integração.

---

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro prático](../docs/learning-roadmap.md)
