# TerraformStack 11A — Argo Workflows UI com Client Auth e RBAC

[← Arquitetura](architecture.md) · [ADR-0006](adr/0006-terraformrun-and-application-ci.md) · [Roadmap](learning-roadmap.md)

## Objetivo e escopo

Validar, no Hub `kind`, a experiência de **aprovação operacional** de Workflow suspenso pela Argo Workflows UI, com identidades Viewer/Approver e RBAC Kubernetes. Os manifests deste guia são **configuração e teste locais**, não um Terraform Executor real. Nenhuma operação AWS, `terraform apply` ou `destroy` é autorizada por este documento.

A Etapa 03 usa o chart Helm `argo-workflows` 2.0.18 com `singleNamespace: true` e `server.enabled: false`. Para exercitar a UI na 11A, habilitar **Argo Server** no mesmo namespace `argo-workflows`, mantendo o controller existente, via alteração revisada da Application GitOps (sem patch manual persistente):

```yaml
# trecho a integrar em:
# workflow-plane/argo-workflows/bootstrap/application.yaml
spec:
  source:
    helm:
      valuesObject:
        server:
          enabled: true
          authModes: [client]
```

**Atenção:** este trecho é **parcial**; preservar todas as demais chaves de `valuesObject` da Application original, em especial `crds`, `singleNamespace` e `workflow.serviceAccount`. Esta branch **não altera a Application existente**; a habilitação da UI fica para a execução controlada da Etapa 11A. Não habilitar `authModes: [server]` em paralelo, para não permitir execução com a identidade do servidor em vez do cliente.

No modo `client`, o Argo Server usa o bearer token Kubernetes apresentado pelo cliente para autorizar requisições. A versão do Kubernetes deve atender aos requisitos do modo client do Argo Server (incluindo `SelfSubjectReview`). A UI será acessada apenas por `kubectl port-forward` no laboratório, sem Ingress público. Não salvar tokens em Git, shell history, arquivos de CI ou screenshots.

## Manifests

- [RBAC Viewer e Approver](../workflow-plane/argo-workflows/rbac/terraform-approval.yaml): duas ServiceAccounts, duas Roles e dois RoleBindings **namespace-scoped**, sem acesso AWS nem permissões de administração do cluster.
- [Workflow de teste](../workflow-plane/argo-workflows/demo/approval-smoke.yaml): `plan` simulado → `suspend: {}` → `apply` simulado; apenas mensagens `echo`, sem executar Terraform.

O `ServiceAccount` de **execução** do Workflow continua sendo `argo-workflow`, diferente das ServiceAccounts usadas para **acessar a UI**. O controller da futura Stack terá identidade própria, e só ele criará workflows Terraform reais.

## Procedimento manual (Etapa 11A)

Após habilitar o Argo Server via Argo CD, verificar:

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub -n argo-workflows get deploy,svc | grep -E 'argo-workflows.*server'
hub -n argo-workflows rollout status deployment/argo-workflows-server --timeout=120s
hub apply -f workflow-plane/argo-workflows/rbac/terraform-approval.yaml
```

Confirmar a matriz de RBAC **antes** de testar a UI:

```bash
ns=argo-workflows
viewer=system:serviceaccount:$ns:terraform-workflows-viewer
approver=system:serviceaccount:$ns:terraform-workflows-approver

hub auth can-i get workflows.argoproj.io -n "$ns" --as="$viewer"         # yes
hub auth can-i patch workflows.argoproj.io -n "$ns" --as="$viewer"       # no
hub auth can-i update workflows.argoproj.io -n "$ns" --as="$viewer"      # no
hub auth can-i patch workflows.argoproj.io -n "$ns" --as="$approver"     # yes
hub auth can-i update workflows.argoproj.io -n "$ns" --as="$approver"    # yes
hub auth can-i get secrets -n "$ns" --as="$approver"                     # no
hub auth can-i create pods -n "$ns" --as="$approver"                     # no
hub auth can-i create workflows.argoproj.io -n "$ns" --as="$approver"    # no
```

Esses resultados pressupõem que as ServiceAccounts de teste não tenham outras permissões herdadas por bindings adicionais; se houver, investigar os bindings efetivos. `update/patch` são permissões **amplas**, que podem modificar o conteúdo de Workflows e não apenas retomá-los. Esta configuração é um compromisso didático aceito para o namespace local, **não é um perfil seguro de aprovação para workloads não confiáveis em produção**.

Em terminais separados, criar os tokens temporários, sem gravá-los em disco:

```bash
# Usar um terminal autorizado, evitando logs ou compartilhamento de sessão:
hub -n argo-workflows create token terraform-workflows-viewer --duration=15m
hub -n argo-workflows create token terraform-workflows-approver --duration=15m

# Executar em outro terminal e manter ativo:
hub -n argo-workflows port-forward svc/argo-workflows-server 2746:2746
```

Abrir `https://localhost:2746` (ou o endereço indicado pelo Argo Server, conforme TLS configurado) e informar `Bearer <TOKEN>` no formulário de login, para cada identidade em sessões separadas. Não adicionar exceção de TLS para acesso remoto. O token de curta duração é escolhido para o laboratório; o servidor Kubernetes pode aplicar um limite diferente. Encerrar as sessões após os testes.

Criar uma execução **simulada** e capturar seu nome:

```bash
run=$(hub create -f workflow-plane/argo-workflows/demo/approval-smoke.yaml -o jsonpath='{.metadata.name}')
echo "$run"
hub -n argo-workflows get workflow "$run" -w
```

Na UI, um Viewer deve enxergar o Workflow suspenso, porém não retomá-lo. Um Approver deverá conseguir retomar a execução. Se a UI exigir verbos adicionais, registrar o erro e avaliar cada verbo — **não atribuir `*`, ClusterRole administrativo, `server` auth mode ou acesso a Secrets como correção rápida**.

Confirmar `Succeeded` e os logs do passo de `apply` **simulado**. O teste de negação pode também ser reproduzido por `kubectl auth can-i`; o teste de botão da UI é obrigatório para fechar o aceite. Criar outro Workflow e deixá-lo suspenso para confirmar que não avança sem retomada; eliminá-lo no cleanup.

Limpar os Workflows de teste e as permissões quando não forem mais necessários:

```bash
hub -n argo-workflows delete workflows -l lab.platform.example/exercise=terraform-approval-smoke --ignore-not-found
hub delete -f workflow-plane/argo-workflows/rbac/terraform-approval.yaml
```

A desativação do Argo Server, se desejada após o lab, deve ocorrer pelo manifesto GitOps revisado, não por `kubectl patch` persistente.

## Sem prazo automático e nova tentativa

O passo `suspend: {}` não tem `duration`: não há timeout de aprovação na 11A. **O token de acesso à UI continua temporário**, independentemente de quanto tempo o Workflow fica suspenso. Se o token expirar, obter outro token para entrar na UI; isso não cancela o Workflow.

Se o Workflow for encerrado/cancelado antes do apply, o controller registrará a execução como cancelada sem iniciar outra automaticamente. A Stack permanece declarada no Git e sincronizada pelo Argo CD. **Não é preciso clicar Sync novamente.** Uma nova tentativa deverá ser solicitada explicitamente ao controller com `platform.example.io/retry-request-id` de UUID único no TerraformStack, conforme [guia de retry](terraformstack-11a-retry-annotation.md); ela cria novo Workflow, novo plan e nova suspensão. Repetir Sync sem mudança na intenção não deve criar Workflow novo. Encerrar uma operação que já esteja aplicando infraestrutura não equivale a rollback.

## Critérios de aceite

| Teste | Resultado esperado |
| :--- | :--- |
| Argo Server | `client` exclusivamente, acessível por port-forward local; `server` desabilitado |
| Viewer | Visualiza Workflow; `patch/update` negados; não consegue `resume` pela UI |
| Approver | Consegue `resume` na UI; não consegue criar Workflow, ler Secrets ou criar Pods |
| Suspensão | `apply` simulado não inicia até a retomada manual; sem `duration` nem TTL de aprovação |
| Continuidade | Ao retomar, Workflow termina `Succeeded` e registra logs de simulação |
| Escopo | Bindings em `argo-workflows`, sem RBAC cluster-wide para os usuários de aprovação |
| Segurança | Tokens temporários não são versionados; UI não é exposta publicamente |
| Limitação documentada | `update/patch` não são permissões exclusivas de aprovação; não declarar auditoria individual forte |

**Próximo passo da implementação real:** integrar ao `TerraformStack Controller` a geração do Workflow e a validação técnica de hash, generation, alvo e plano antes do `terraform apply`. Este smoke test **não** prova essas garantias e não substitui os critérios de aceite completos da 11A.
