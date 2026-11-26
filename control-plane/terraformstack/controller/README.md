# TerraformStack Controller

Controller Go baseado em controller-runtime (framework do Kubebuilder). O CRD versionado em ../crd é a fonte canônica do schema. Tipos Go do spec permitem validação e resolução da intenção, enquanto objetos unstructured mantêm compatibilidade com o JSON de execução e os Workflows Argo.

## Runtime

- Watches de TerraformStack e Workflow, índice de dependências e fila com rate limiting.
- Workers concorrentes (4 por padrão), Leader Election via Lease, duas réplicas.
- Leitura direta da stack, Workflow e dependências para autorização; status Update usa resourceVersion e impede sobrescrever claim concorrente.
- RequeueAfter agenda drift/retry sem varrer stacks periodicamente.
- Métricas controller-runtime em 8080; healthz/readyz em 8081. Não expor métricas publicamente.
- Namespace e catálogo configuráveis por flags; RBAC deve acompanhar mudanças de namespace.

## Testes

```bash
go mod tidy -diff
go test -race ./...
go vet ./...
```

O CI executa envtest com API Kubernetes/etcd e testes unitários. A Etapa 06 valida integração com Argo/OPA/Terraform em Kind.

[Lab 05](../../../labs/05-terraformstack-runtime.md) · [Fluxo](../../../docs/flows/terraformstack.md)

## Contrato de integração

[controller-config.yaml](controller-config.yaml) é a configuração de instalação
do Platform Team. Define namespaces, workers, Leader Election, referência do
WorkflowTemplate, nomes de parâmetros de entrada/saída e referências aos
templates de policy e aprovação. O catálogo continua sendo a fonte autorizada
de identidades, backend, volumes e perfis OPA. O WorkflowTemplate controla a
execução interna.

A configuração é validada estritamente no startup (campos desconhecidos,
nomes inválidos, duplicações e campos obrigatórios). Kustomize gera ConfigMap
imutável e rollout por hash. Uma tentativa ativa guarda o snapshot do contrato;
alterações incompatíveis bloqueiam reconciliação até restaurar a configuração.
Drene execuções antes de mudar o contrato. Não é um mecanismo de revogação de
um apply já autorizado.

Os nomes podem mudar; os valores e semântica do protocolo permanecem fixos:
intent JSON, volume, decisão OPA, hash de autorização, resultado de aprovação
e verificação. O runner precisa respeitar o claim com resourceVersion. Não há
plugins ou escolha de template pelo CR. Ao mudar namespaces ou template,
ajuste também scripts do runner (namespace fixo na referência), RBAC, admissão, destino GitOps e WorkflowTemplate; a configuração
não concede permissões nem renomeia automaticamente objetos externos.

## Fronteira de execução

O package `internal/controller` mantém reconciliação, lifecycle, dependências e autorização. `internal/execution` define uma interface interna mínima de criação, leitura normalizada, cancelamento e snapshot. `internal/execution/argo` concentra objetos Workflow, parâmetros, resultados e mapeamento de eventos. O manager conecta o watch do adapter ao controller. Argo Workflows continua sendo a única integração suportada; não há plugins nem seleção de engine pelo usuário. O runner, OPA e WorkflowTemplate são uma implementação de referência da POC, não dependências compiladas do controller. A implementação de referência deve cumprir o protocolo de autorização; mantenha scripts, RBAC e admissão compatíveis com os namespaces e nomes escolhidos.
