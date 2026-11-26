# Etapa 06 — Terraform Stack Simulator de ponta a ponta no kind

## Objetivo

Reconciliar uma stack Terraform real no Hub local e comprovar pelo status do `TerraformStack` as decisões `AutoApply`, `RequireApproval` e `Deny`, o drift do ambiente e o teardown governado.

## O que será feito

O Argo CD sincroniza a intenção em Git. O controller fixa as revisões de código e configuração e abre um Workflow. O Terraform grava o mapa efetivo de Network no campo `network.json` de um ConfigMap no namespace `terraformstack-simulator`. O exercício usa a hierarquia de Configuration de `entity-a/uat` e identidade Kubernetes local.

## Por que fizemos assim

O mesmo caminho de reconciliação atende o primeiro plan, as mudanças declaradas e a observação periódica do ambiente. Mantemos o state em um PVC por stack, com locking do backend local, e o plano binário em um diretório por execução. O OPA decide sobre o JSON derivado do plano. O apply revalida o hash, a política, a intenção e o efeito previsto antes de executar o binário salvo. O status registra a identidade da execução, a revisão, o resultado da política e o sinal de drift.

O catálogo local autoriza os dois nomes do exercício, a raiz Terraform do repositório e o destino lógico `entity-a/simulator`. O Simulator lê Configuration com o perfil de dados `entity-a/uat`. O código Terraform das revisões aceitas é mantido pelo Platform Team. O Workflow usa uma versão imutável das regras Rego e o status publica seu SHA-256. [Contrato](../docs/flows/terraformstack.md) · [Política OPA](../workflow-plane/argo-workflows/terraformstack/policy.rego)

## 1. Preparar o Hub e sincronizar

Reutilize o Hub com auditoria da [Etapa 01](01-platform-management-cluster.md) e o runtime da [Etapa 05](05-terraformstack-runtime.md). No Argo CD, abra **terraformstack-simulator**. Confira o destino `platform-system`, os dois CRs e a anotação `platform.example.io/revision` com o SHA resolvido. A dependente começa suspensa. Para um preview, fixe a Application no SHA de exercício acessível ao Hub. [Entrada do Simulator](../platform/hub/addons/catalog/terraformstack-simulator.yaml)

```bash
cluster="${HUB_KIND_NAME:-platform-management}"
hub() { kubectl --context "kind-$cluster" "$@"; }
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
hub -n argo-workflows get workflowtemplate terraformstack
stack() { hub -n platform-system get terraformstack simulator-network "$@"; }
workflow() { stack -o jsonpath='{.status.execution.workflow}'; }
network() {
  hub -n terraformstack-simulator get configmap simulator-network \
    -o jsonpath='{.data.network\.json}'
  printf '\n'
}
stack -o yaml
hub -n argo-workflows get workflow "$(workflow)"
network
```

Se o Workflow ficar pendente, confira eventos dos Pods e dos PVCs. Para falha no plan, consulte a fase e os logs do Workflow. Confirme que o SHA publicado é acessível dentro do Hub e que os Pods podem obter o provider Kubernetes. Em `ImagePullBackOff`, confira as imagens locais carregadas no `kind` e a conectividade com o registry do OPA.

## 2. Exercitar as decisões OPA pelo GitOps

Use commits em uma branch de exercício para alterar apenas `network.vpc.cidr` de [Network em UAT](../foundation/config/entity-a/uat/network.yaml). Em cada cenário, selecione o SHA publicado na Application de solicitações, confirme a nova anotação no diff e sincronize manualmente.

| Cenário | Decisão esperada | Evidência |
| :--- | :--- | :--- |
| Primeiro sync com `10.81.0.0/16` | `AutoApply` | `Ready` após apply e verificação. O ConfigMap também contém os defaults de DNS, NAT e endpoints |
| Revisão com `10.82.0.0/16` | `RequireApproval` | `ApprovalRequired`. O ConfigMap mantém o CIDR anterior enquanto a aprovação está pendente |
| Revisão com `0.0.0.0/0` | `Deny` | `ReconcileBlocked`, motivo `PolicyDenied` e ConfigMap preservado |

Para aprovar, o administrador do Hub provisiona uma ServiceAccount no namespace de teste com nome individual. Cada aprovador usa apenas seu token temporário. A política de admissão limita a edição ao gate ativo do Simulator.

```bash
approver=aprovador-local
hub -n terraformstack-simulator create serviceaccount "$approver"
hub -n argo-workflows create rolebinding "$approver" \
  --role=terraformstack-approvers \
  --serviceaccount="terraformstack-simulator:$approver"
printf 'Bearer %s\n' "$(hub -n terraformstack-simulator create token "$approver" --duration=30m)"
stack -o jsonpath='{.status.policy.hash}{"\n"}'
```

Na [UI do Argo Workflows](https://workflows.localhost:8443), autentique com o token, selecione `argo-workflows`, abra o Workflow e o nó **approval**. No **Resume**, selecione `Approve` e informe o hash completo apresentado no status. Confira `status.policy.cidrs` para ver os CIDRs do plano antes e depois. Confira `Ready` e `network` com o novo valor. Repita com `Reject`, com hash incorreto e com a identidade somente leitura da Etapa 03. Cada tentativa que altera a decisão precisa de um Workflow novo, acionado por uma revisão aprovada. A resposta padrão do gate é `Reject`.

```bash
docker exec "$cluster-control-plane" cat /var/local/terraformstack-audit/requests.jsonl |
  python3 -c 'import json,sys
for line in sys.stdin:
    event = json.loads(line)
    print(event["user"]["username"], event.get("objectRef", {}).get("name"),
          event.get("responseStatus", {}).get("code"))'
```

Correlacione a entrada de auditoria bem-sucedida com o Workflow e o hash do plano. A política foi criada somente para o Hub descartável. Os eventos completos ficam no nó do `kind`.

Para provar falha fechada no Hub descartável, altere o Rego para uma expressão inválida em uma revisão de teste. Sincronize **terraformstack-workflows** nessa revisão e depois sincronize o Simulator com uma mudança de CIDR. Observe o step OPA falhar, o status indicar `PolicyError` e o ConfigMap preservar o valor anterior. Restaure o WorkflowTemplate com o sync da revisão válida. Os ConfigMaps de scripts e política são imutáveis e têm nomes derivados do conteúdo. Deixe a limpeza das versões antigas para depois que seus Workflows concluírem.

Para provar integridade, monte o PVC `simulator-network` em um Pod de diagnóstico temporário no namespace `argo-workflows`. Use a imagem do executor, comando `sleep 600`, `automountServiceAccountToken: false` e monte o PVC em `/state`. Enquanto o gate aguarda aprovação, altere o binário no volume.

```bash
hub -n argo-workflows exec simulator-plan-diagnostic -- \
  sh -c 'printf x >> "/state/$1/plan.bin"' sh "$(workflow)"
hub -n argo-workflows delete pod simulator-plan-diagnostic
```

Aprove com o hash original e confira a falha antes do apply. O retry configurado abre outro Workflow com um plano novo e aguarda outra aprovação.

## 3. Observar drift e reconciliação

Volte ao CIDR de baseline por GitOps e aprove pelo mesmo gate. Observe a próxima checagem devolver `NoChanges`. Em seguida, altere só o CIDR do ConfigMap. Este patch simula a alteração externa ao Terraform.

```bash
payload="$(network | python3 -c 'import json,sys
data = json.load(sys.stdin)
data["vpc"]["cidr"] = "10.83.0.0/16"
print(json.dumps({"data":{"network.json":json.dumps(data)}}))')"
hub -n terraformstack-simulator patch configmap simulator-network --type merge -p "$payload"
hub -n platform-system get terraformstack simulator-network -w \
  -o custom-columns=GEN:.metadata.generation,WORKFLOW:.status.execution.workflow,GATILHO:.status.execution.trigger,CONDICAO:.status.conditions[0].type,DRIFT:.status.lastDriftTime
```

Sem novo commit ou sync, aguarde `DriftCheck` na mesma revisão, a condição `DriftDetected` ou o horário de drift persistido e o retorno do CIDR ao baseline. Confira o Workflow periódico e a próxima checagem sem mudanças. Um sync repetido na mesma revisão e o reinício do controller conservam o Workflow ativo. O intervalo padrão desta solicitação é 45 segundos.

No ensaio de concorrência, abra um gate e sincronize uma nova revisão antes de aprovar. Confira o Workflow antigo terminado e um novo hash e Workflow. Durante um apply já iniciado, sincronize uma segunda revisão e confira que ela aguarda a verificação da primeira. Use `status.execution.applying` como sinal para iniciar essa segunda sincronização.

Para testar dependência, libere `suspend` de `simulator-dependent` com a Network ainda suspensa ou bloqueada e confira `DependencyBlocked`. Coloque a Network em `Ready` e verifique a execução da dependente e seus state e ConfigMap próprios. Para observar retry operacional, interrompa um Pod de plan em uma execução de teste e confira a falha, o intervalo de 10 segundos e a nova tentativa com Workflow e plan próprios. Restaure as solicitações pelo GitOps depois dos testes.

## 4. Exercitar destroy e limpar

1. Com ambas as stacks ativas, solicite `spec.lifecycle.destroy: true` na Network. Confirme bloqueio pela dependência. Destrua a dependente primeiro ou retire sua dependência por uma revisão autorizada.
2. Confira o destroy explícito da Network chegando ao gate. Aprove o hash, aguarde `Destroyed`, confirme a ausência do ConfigMap e a permanência do CR.
3. Recrie a stack por uma revisão com `destroy: false`. Remova o CR com o comportamento padrão e confira retenção do ConfigMap e PVC. Sincronize novamente para retomar o mesmo state.
4. Em outra execução, finalize o destroy da dependente e remova o CR dela antes do teste de cleanup da Network. Sincronize `destroyResourcesOnDeletion: true` na Network, confirme que a intenção está vigente e exclua o CR. Confirme `deletionTimestamp` e a pausa para aprovação. Ao negar, preserve o ConfigMap e o CR em exclusão. Ao aprovar em um ensaio separado, confirme a ausência do ConfigMap e a remoção do CR somente depois da verificação.

Limpe primeiro as solicitações. Exclua o Workflow e os PVCs do Simulator apenas quando não houver apply em execução e o state não for mais necessário. Para um ensaio de negativa com finalizer pendente, mantenha as evidências e descarte o Hub dedicado ao encerrar o exercício. Em um Hub compartilhado, aplique apenas o procedimento de recuperação autorizado para aquela stack.

## O que observar

- `status.observedGeneration`, `resolvedSource` e `execution.workflow` ligam intenção, commit e execução. O GitOps mostra a sincronização do CR. `lastCheckTime` e `lastDriftTime` mostram a observação do recurso pelo Terraform.
- `status.policy` guarda decisão, motivo, hash, resumo das ações, fase e identidade da aprovação. O log de auditoria do Hub confirma a identidade que retomou o gate. As evidências são o status, o ConfigMap e a ausência de mudança em um segundo plan.
- Guarde revisão, ID ou digest de imagem e resultados do exercício. O PVC contém os planos e states locais.

## Evidências da versão anterior

No ensaio anterior à consolidação da pipeline, em 25/11/2026, o Hub `kind` com Kubernetes 1.35 recebeu pelo Argo CD os SHAs de uma prévia Git privada. O controller local em execução teve ID de imagem `sha256:6468df47cae38721e1f4191c38e2cb972467b36c0ee985fc1b4070885e20e32c`. O baseline dos manifestos de teste partiu de `e363da42a3cd2120be9b463a790bb20c79060944`. As revisões de alteração do CIDR ficaram no espelho local.

| Cenário | Resultado observado |
| :--- | :--- |
| `AutoApply` | ConfigMap com CIDR `10.81.0.0/16` e valores herdados de DNS, NAT e endpoints |
| `RequireApproval` | **Resume** no nó da UI aplicou `10.82.0.0/16` com o hash aprovado. O status e a auditoria registraram a identidade |
| `Deny` e erro OPA | `PolicyDenied` bloqueou `0.0.0.0/0`. `PolicyError` registrou Rego inválido. Os dois cenários preservaram o ConfigMap |
| Drift e idempotência | A checagem periódica detectou `10.83.0.0/16` e restaurou o baseline na mesma revisão. Sync repetido e reinício conservaram o Workflow que aguardava aprovação |
| Concorrência e integridade | A revisão nova cancelou um gate obsoleto; o apply que já havia começado concluiu. Hash incorreto bloqueou aprovação. Plano adulterado falhou e o retry gerou outro plano |
| Lifecycle | Suspensão e dependência ficaram visíveis no status. Destroy explícito removeu os dois ConfigMaps após aprovação. A exclusão padrão reteve o recurso. O cleanup opt-in aprovado liberou o finalizer; a negativa pela UI preservou o finalizer e o ConfigMap |

`SimulatedPlanReady=True` confirmou o antigo controller simulado naquele ensaio. O runtime e o CRD passaram pelo dry-run na API. Os sete testes Rego passaram.

[← Etapa 05](05-terraformstack-runtime.md) · [Roteiro](../docs/learning-roadmap.md) · [Etapa 07 →](07-crossplane-core.md)
