# Etapa 06B — TerraformStack de ponta a ponta no kind

## Objetivo

Reconciliar uma stack Terraform real no Hub local e comprovar pelo status do `TerraformStack` as decisões `AutoApply`, `RequireApproval` e `Deny`, o drift do ambiente e o teardown governado.

## O que será feito

O Argo CD sincroniza a intenção em Git. O controller fixa as revisões de código e configuração e abre um Workflow. O Terraform grava o mapa efetivo de Network no campo `network.json` de um ConfigMap no namespace `terraformstack-lab06b`. O exercício usa a hierarquia de Configuration de `entity-a/uat` e identidade Kubernetes local.

## Por que fizemos assim

O mesmo caminho de reconciliação atende o primeiro plan, as mudanças declaradas e a observação periódica do ambiente. Mantemos o state em um PVC por stack, com locking do backend local, e o plano binário em um diretório por execução. O OPA decide sobre o JSON derivado do plano. O apply revalida o hash, a política, a intenção e o efeito previsto antes de executar o binário salvo. O status registra a identidade da execução, a revisão, o resultado da política e o sinal de drift.

O catálogo local autoriza os dois nomes do exercício, a raiz Terraform do repositório e o destino lógico `entity-a/uat`. O código Terraform das revisões aceitas é mantido pelo Platform Team. A política Rego acompanha a imagem do executor e o status publica seu SHA-256. [Contrato](../docs/flows/terraformstack.md) · [Política OPA](../workflow-plane/terraform-executor/lab06b/policy.rego)

## 1. Preparar o Hub e sincronizar

Em um Hub descartável, use o bootstrap com auditoria no lugar da criação inicial do cluster da Etapa 01. Instale Traefik e Argo CD conforme as Etapas [01](01-platform-management-cluster.md) e [02](02-argocd-central.md) e sincronize Argo Workflows da [Etapa 03](03-argo-workflows.md).

```bash
bash platform/hub/scripts/create-lab06b.sh
hub() { kubectl --context kind-platform-management "$@"; }

docker build -t terraformstack-controller:lab06b control-plane/terraformstack/controller
docker build -t terraformstack-executor:lab06b -f workflow-plane/terraform-executor/lab06b/Dockerfile .
docker image inspect terraformstack-controller:lab06b --format '{{.Id}}'
kind load docker-image --name platform-management terraformstack-controller:lab06b terraformstack-executor:lab06b
hub apply -f platform/hub/addons/bootstrap.yaml
hub get storageclass
```

Registre o ID da imagem local, a revisão do build e, na publicação em registry, o digest do manifesto da imagem a promover. Confirme que a StorageClass local é padrão. Aguarde o CRD da [Etapa 05](05-terraformstack-crd.md) com o schema atualizado, incluindo `status.execution` e `status.policy`.

Abra **terraformstack-local** no Argo CD, confirme o destino Hub, a revisão e o diff com controller, WorkflowTemplate, dois PVCs, namespace, RBAC e políticas de admissão. Sincronize manualmente e aguarde o Deployment. Abra **terraformstack-local-requests** e confira os dois CRs e a anotação `platform.example.io/revision` com o SHA resolvido antes do sync. A stack dependente começa suspensa.

Para prévia de uma branch publicada, crie Applications temporárias com sufixo `-preview` usando os mesmos projetos e caminhos das entradas do catálogo, fixando `targetRevision` no SHA desejado. Faça o mesmo com a Application do CRD se o schema ainda não estiver em `main`. Em cada caso use uma única Application para gerenciar os mesmos objetos. [Entrada do runtime](../platform/hub/addons/catalog/terraformstack-local.yaml) · [Entrada das solicitações](../platform/hub/addons/catalog/terraformstack-local-requests.yaml)

```bash
hub -n platform-system rollout status deployment/terraformstack-local-controller --timeout=120s
hub -n argo-workflows get workflowtemplate terraformstack-local
stack() { hub -n platform-system get terraformstack lab06b-network "$@"; }
workflow() { stack -o jsonpath='{.status.execution.workflow}'; }
network() {
  hub -n terraformstack-lab06b get configmap lab06b-network     -o jsonpath='{.data.network\.json}'
  printf '\n'
}
stack -o yaml
hub -n argo-workflows get workflow "$(workflow)"
network
```

Se o Workflow ficar pendente, confira eventos dos Pods e dos PVCs. Para falha no plan, consulte a fase e os logs do Workflow. Confirme que o SHA publicado é acessível dentro do Hub e que os Pods podem obter o provider Kubernetes. Em `ErrImageNeverPull`, confira os nomes carregados no `kind`.

## 2. Exercitar as decisões OPA pelo GitOps

Use commits em uma branch de exercício para alterar apenas `network.vpc.cidr` de [Network em UAT](../foundation/config/entity-a/uat/network.yaml). Em cada cenário, selecione o SHA publicado na Application de solicitações, confirme a nova anotação no diff e sincronize manualmente.

| Cenário | Decisão esperada | Evidência |
| :--- | :--- | :--- |
| Primeiro sync com `10.81.0.0/16` | `AutoApply` | `Ready` após apply e verificação. O ConfigMap também contém os defaults de DNS, NAT e endpoints |
| Revisão com `10.82.0.0/16` | `RequireApproval` | `ApprovalRequired`. O ConfigMap mantém o CIDR anterior enquanto a aprovação está pendente |
| Revisão com `0.0.0.0/0` | `Deny` | `ReconcileBlocked`, motivo `PolicyDenied` e ConfigMap preservado |

Para aprovar, o administrador do Hub provisiona uma ServiceAccount no namespace de teste com nome individual. Cada aprovador usa apenas seu token temporário. A política de admissão limita a edição ao gate ativo do 06B.

```bash
approver=aprovador-david
hub -n terraformstack-lab06b create serviceaccount "$approver"
printf 'Bearer %s\n' "$(hub -n terraformstack-lab06b create token "$approver" --duration=30m)"
stack -o jsonpath='{.status.policy.hash}{"\n"}'
```

Na [UI do Argo Workflows](https://workflows.localhost:8443), autentique com o token, selecione `argo-workflows`, abra o Workflow e o nó **approval**. No **Resume**, selecione `Approve` e informe o hash completo apresentado no status. Confira `Ready` e `network` com o novo valor. Repita com `Reject`, com hash incorreto e com a identidade somente leitura da Etapa 03. Cada tentativa que altera a decisão precisa de um Workflow novo, acionado por uma revisão aprovada. A resposta padrão do gate é `Reject`.

```bash
docker exec platform-management-control-plane cat /var/local/lab06b-audit/requests.jsonl |
  python3 -c 'import json,sys
for line in sys.stdin:
    event = json.loads(line)
    print(event["user"]["username"], event.get("objectRef", {}).get("name"),
          event.get("responseStatus", {}).get("code"))'
```

Correlacione a entrada de auditoria bem-sucedida com o Workflow e o hash do plano. A política foi criada somente para o Hub descartável. Os eventos completos ficam no nó do `kind`.

Para provar falha fechada, carregue uma imagem de executor de teste com Rego sintaticamente inválido e aponte temporariamente as imagens dos steps do WorkflowTemplate para essa tag. Sincronize uma nova revisão permitida e confira `PolicyError` sem alteração do ConfigMap. Restaure o template pela Application do runtime. Para provar integridade, altere um byte de `/state/<workflow>/plan.bin` num Pod de diagnóstico autorizado pelo administrador enquanto o gate aguarda aprovação. Aprove com o hash original e confirme a falha antes de apply. Remova o Pod de diagnóstico.

## 3. Observar drift e reconciliação

Volte ao CIDR de baseline por GitOps e aprove pelo mesmo gate. Observe a próxima checagem devolver `NoChanges`. Em seguida, altere só o CIDR do ConfigMap. Este patch simula a alteração externa ao Terraform.

```bash
payload="$(network | python3 -c 'import json,sys
data = json.load(sys.stdin)
data["vpc"]["cidr"] = "10.83.0.0/16"
print(json.dumps({"data":{"network.json":json.dumps(data)}}))')"
hub -n terraformstack-lab06b patch configmap lab06b-network --type merge -p "$payload"
hub -n platform-system get terraformstack lab06b-network -w \
  -o custom-columns=GEN:.metadata.generation,WORKFLOW:.status.execution.workflow,GATILHO:.status.execution.trigger,CONDICAO:.status.conditions[0].type,DRIFT:.status.lastDriftTime
```

Sem novo commit ou sync, aguarde `DriftCheck` na mesma revisão, a condição `DriftDetected` ou o horário de drift persistido e o retorno do CIDR ao baseline. Confira o Workflow periódico e a próxima checagem sem mudanças. Um sync repetido na mesma revisão e o reinício do controller conservam o Workflow ativo. O intervalo padrão desta solicitação é 45 segundos.

No ensaio de concorrência, abra um gate e sincronize uma nova revisão antes de aprovar. Confira o Workflow antigo terminado e um novo hash e Workflow. Durante um apply já iniciado, sincronize uma segunda revisão e confira que ela aguarda a verificação da primeira. Use `status.execution.applying` como sinal para iniciar essa segunda sincronização.

Para testar dependência, libere `suspend` de `lab06b-dependent` com a Network ainda suspensa ou bloqueada e confira `DependencyBlocked`. Coloque a Network em `Ready` e verifique a execução da dependente e seus state e ConfigMap próprios. Para observar retry operacional, interrompa um Pod de plan em uma execução de teste e confira a falha, o intervalo de 10 segundos e a nova tentativa com Workflow e plan próprios. Restaure as solicitações pelo GitOps depois dos testes.

## 4. Exercitar destroy e limpar

1. Com ambas as stacks ativas, solicite `spec.lifecycle.destroy: true` na Network. Confirme bloqueio pela dependência. Destrua a dependente primeiro ou retire sua dependência por uma revisão autorizada.
2. Confira o destroy explícito da Network chegando ao gate. Aprove o hash, aguarde `Destroyed`, confirme a ausência do ConfigMap e a permanência do CR.
3. Recrie a stack por uma revisão com `destroy: false`. Remova o CR com o comportamento padrão e confira retenção do ConfigMap e PVC. Sincronize novamente para retomar o mesmo state.
4. Em outra execução, sincronize `destroyResourcesOnDeletion: true` antes de excluir o CR. Confirme `deletionTimestamp` e a pausa para aprovação. Ao negar, preserve o ConfigMap e o CR em exclusão. Ao aprovar em um ensaio separado, confirme a ausência do ConfigMap e a remoção do CR somente depois da verificação.

Limpe primeiro as solicitações. Exclua o Workflow e os PVCs do 06B apenas quando não houver apply em execução e o state não for mais necessário. Para um ensaio de negativa com finalizer pendente, mantenha as evidências e descarte o Hub dedicado ao encerrar o exercício. Em um Hub compartilhado, aplique apenas o procedimento de recuperação autorizado para aquela stack.

## O que observar

- `status.observedGeneration`, `resolvedSource` e `execution.workflow` ligam intenção, commit e execução. O GitOps mostra a sincronização do CR. `lastCheckTime` e `lastDriftTime` mostram a observação do recurso pelo Terraform.
- `status.policy` guarda decisão, motivo, hash, resumo das ações e fase da aprovação. O log de auditoria do Hub confirma a identidade que retomou o gate. As evidências são o status, o ConfigMap e a ausência de mudança em um segundo plan.
- Guarde revisão, ID ou digest de imagem e resultados do exercício. O PVC contém os planos e states locais.

[← Etapa 06](06-terraformstack-controller.md) · [Roteiro](../docs/learning-roadmap.md) · [Etapa 07 →](07-registro-ambientes-aws.md)
