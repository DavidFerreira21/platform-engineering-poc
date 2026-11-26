# TerraformStack Controller

Controller Go baseado em controller-runtime (framework do Kubebuilder). O CRD versionado em ../crd é a fonte canônica do schema. Tipos Go do spec permitem validação e resolução da intenção, enquanto objetos unstructured mantêm compatibilidade com o JSON de execução e os Workflows Argo.

## Runtime

- Watches de TerraformStack e Workflow, índice de dependências e fila com rate limiting.
- Workers concorrentes (4 por padrão), Leader Election via Lease, duas réplicas.
- Leitura direta da stack e Workflow para autorização; status Update usa resourceVersion e impede sobrescrever claim concorrente.
- RequeueAfter agenda drift/retry sem varrer stacks periodicamente.
- Métricas controller-runtime em 8080; healthz/readyz em 8081. Não expor métricas publicamente.
- Namespace e catálogo configuráveis por flags; RBAC deve acompanhar mudanças de namespace.

## Testes

```bash
go mod tidy
go test -race ./...
go vet ./...
```

O CI executa envtest com API Kubernetes/etcd e testes unitários. A Etapa 06 valida integração com Argo/OPA/Terraform em Kind.

[Lab 05](../../../labs/05-terraformstack-runtime.md) · [Fluxo](../../../docs/flows/terraformstack.md)
