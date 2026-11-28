# Etapa 06 — Terraform Stack Simulator de ponta a ponta no kind

## Objetivo

Reconciliar uma stack Terraform real no Hub local e comprovar pelo status do `TerraformStack` as decisões `AutoApply`, `RequireApproval` e `Deny`, o drift do ambiente e o teardown governado.

## O que será feito

O Argo CD sincroniza a intenção em Git. O controller abre um Workflow e a pipeline fixa o commit de código e configuração. O Terraform grava o mapa efetivo de Network no campo `network.json` de um ConfigMap no namespace `terraformstack-simulator`. O exercício usa a hierarquia de Configuration de `entity-a/uat` e identidade Kubernetes local.

## Por que fizemos assim

Este lab exercita a integração de referência `terraformstack`, com `contract.authorization: true`. Não altere esse campo para `false` no template atual: ele depende do handshake. O mesmo caminho de reconciliação atende o primeiro plan, as mudanças declaradas e a observação periódica do ambiente. Mantemos o state em um PVC por stack, com locking do backend local, e o plano binário em um diretório por execução. O OPA avalia o JSON derivado do plano. Antes do apply, a pipeline confere o hash do binário e compara o efeito previsto com um novo plan. A função `claim()` confere a intenção e a autorização vigentes no CR e registra `status.claim` por atualização condicional. O apply usa o binário salvo. [Autorização final do apply](../docs/flows/terraformstack.md#autorização-final-do-apply)

O perfil local da pipeline autoriza os dois nomes do exercício, a raiz Terraform do repositório e o destino lógico `entity-a/simulator`. O Simulator lê Configuration com o perfil de dados `entity-a/uat`. O código Terraform das revisões aceitas é mantido pelo Platform Team. O Workflow usa o perfil `simulator`, composto de `common` e `simulator`, em um pacote imutável; o status publica a decisão, o hash do plano e o SHA-256 do pacote. [Contrato](../docs/flows/terraformstack.md) · [Política OPA](../workflow-plane/policy-engine/README.md)

## Pré-requisitos

Runtime da Etapa 05 instalado. Execute os comandos na raiz do repositório e no mesmo terminal, com Python 3 para as consultas de configuração e auditoria. Os cenários GitOps usam commits publicados no repositório autorizado pelo perfil. [Etapa 05](05-terraformstack-runtime.md) · [Perfil local](../workflow-plane/argo-workflows/terraformstack/simulator.json)

## 1. Preparar o Hub e sincronizar

Reutilize o Hub `platform-management` da Etapa 01. Termine a instalação do runtime e confirme o rollout do controller antes do sync do Simulator. Se as imagens ainda precisarem ser carregadas ou atualizadas, execute o bloco de reconstrução da Etapa 05. [Hub local](01-platform-management-cluster.md) · [Reconstruir e carregar as imagens](05-terraformstack-runtime.md#1-construir-ou-reconstruir-as-imagens)

No Argo CD, abra **terraformstack-simulator**. Confira o destino `platform-system`, os dois CRs e a anotação `platform.example.io/revision` com o SHA resolvido e faça o sync manual. A dependente começa suspensa. Para testar um commit publicado em branch, informe o SHA em **Sync → Revision** da mesma Application. Aguarde o fim do sync antes de consultar o CR. [Entrada do Simulator](../platform/hub/addons/catalog/terraformstack-simulator.yaml)

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
hub -n argo-workflows get workflowtemplate terraformstack
stack() { hub -n platform-system get terraformstack simulator-network "$@"; }
workflow() { stack -o jsonpath='{.status.execution.workflow}'; }
network() {
  hub -n terraformstack-simulator get configmap simulator-network \
    -o jsonpath='{.data.network\.json}'
  printf '\n'
}
hub -n platform-system wait terraformstack/simulator-network \
  --for=condition=Ready --timeout=180s
stack -o yaml
run="$(workflow)"
hub -n argo-workflows get workflow "$run"
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow=$run" \
  -o 'custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[-1].image,IMAGEID:.status.containerStatuses[?(@.name=="main")].imageID'
network
```

O primeiro sync entrega os CRs. O ConfigMap aparece depois do apply. Após atualizar o runner, confira o image ID do container `main` em um step Terraform desse Workflow. Aguarde `Ready` para observar o resultado de `verify`. Se a espera expirar, execute `stack -o yaml` e abra o Workflow indicado por `status.execution.workflow`. Confira eventos dos Pods e PVCs, a fase e os logs. Confirme que o SHA publicado é acessível dentro do Hub e que os Pods podem obter o provider Kubernetes. Em `ImagePullBackOff` de imagem com tag `:local`, repita o build e o carregamento no mesmo Hub. Para falha ao baixar a imagem do OPA, confira a conectividade com o registry. [Comandos de build e carregamento](05-terraformstack-runtime.md#1-construir-ou-reconstruir-as-imagens)

## 2. Exercitar as decisões OPA pelo GitOps

Crie uma branch de exercício a partir da revisão instalada na Etapa 05.

```bash
git switch -c teste/simulator-cidr
```

Para cada mudança de CIDR, edite `network.vpc.cidr` no YAML de Network em UAT. Confira a mudança no commit, publique e copie o SHA para **Sync → Revision** na Application **terraformstack-simulator**. Após o sync, confira no CR a anotação `platform.example.io/revision`. [YAML de Network em UAT](../foundation/config/entity-a/uat/network.yaml)

```bash
git add foundation/config/entity-a/uat/network.yaml
git commit -m "Exercita mudança de CIDR no Simulator"
git show --stat HEAD
git push -u origin HEAD
git rev-parse HEAD
```

| Cenário | Decisão esperada | Evidência |
| :--- | :--- | :--- |
| Primeiro sync com `10.81.0.0/16` | `AutoApply` | `Ready` após apply e verificação. O ConfigMap também contém os defaults de DNS, NAT e endpoints |
| Revisão com `10.82.0.0/16` | `RequireApproval` | `ExecutionWaiting / Waiting`. O ConfigMap mantém o CIDR anterior enquanto a aprovação está pendente |
| Revisão com `0.0.0.0/0` | `Deny` | `ReconcileBlocked`, com diagnóstico `PolicyDenied` em `status.result.reason` e ConfigMap preservado |

Para aprovar, o administrador do Hub provisiona uma ServiceAccount no namespace de teste com nome individual. Cada aprovador usa apenas seu token temporário. A política de admissão limita a edição ao gate ativo do Simulator.

```bash
approver=aprovador-local
hub -n terraformstack-simulator create serviceaccount "$approver"
hub -n argo-workflows create rolebinding "$approver" \
  --role=terraformstack-approvers \
  --serviceaccount="terraformstack-simulator:$approver"
printf 'Bearer %s\n' "$(hub -n terraformstack-simulator create token "$approver" --duration=30m)"
hub -n platform-system wait terraformstack/simulator-network \
  --for=condition=ExecutionWaiting --timeout=180s
stack -o jsonpath='{.status.result.details.planHash}{"\n"}'
```

Na UI do Argo Workflows, autentique com o token e selecione `argo-workflows`. Abra o Workflow indicado por `workflow` e clique no nó **approval** no grafo. Use **Resume** no painel desse nó. No formulário, selecione `Approve` no campo `approval`, informe o hash completo em `approved-hash` e confirme com **OK**. Compare a alteração proposta no Git com o CIDR observado pela função `network`. Após o apply, confira `Ready / Verified` e o novo valor com a função `network`. [UI do Argo Workflows](https://workflows.localhost:8443)

Experimente o Resume com a identidade somente leitura da Etapa 03 enquanto o gate estiver aberto. A UI deve recusar a operação e manter o gate pendente. Para testar `Reject` e hash incorreto, volte a propor `10.81.0.0/16` depois de aplicar `10.82.0.0/16`. Cada resposta encerra a tentativa com `ReconcileBlocked / ExecutionBlocked` e diagnóstico `ApprovalRejected`. Para repetir com o mesmo CIDR, publique outro commit com `git commit --allow-empty -m "Repete o cenário de aprovação do Simulator"` e sincronize o novo SHA. A resposta padrão do gate é `Reject`.

```bash
docker exec platform-management-control-plane cat /var/local/terraformstack-audit/requests.jsonl |
  python3 -c 'import json,sys
for line in sys.stdin:
    event = json.loads(line)
    print(event["user"]["username"], event.get("objectRef", {}).get("name"),
          event.get("responseStatus", {}).get("code"))'
```

Correlacione a entrada de auditoria bem-sucedida com o Workflow e o hash do plano. Os eventos completos ficam no nó do Hub `platform-management`. Se o arquivo estiver ausente, confira a política de auditoria montada pelo [ambiente local do Simulator](../platform/hub/kind-config.yaml).

Para provar falha fechada no Hub local, altere o Rego para uma expressão inválida em uma revisão de teste. Sincronize **terraformstack-workflows** nessa revisão e depois sincronize o Simulator com uma mudança de CIDR. Observe o step OPA falhar, o status indicar `PolicyError` e o ConfigMap preservar o valor anterior. Restaure o WorkflowTemplate com o sync da revisão válida. Os ConfigMaps de scripts e política são imutáveis e têm nomes derivados do conteúdo. Deixe a limpeza das versões antigas para depois que seus Workflows concluírem.

Para provar integridade, sincronize outra revisão que exija aprovação. Com o gate aberto, monte o PVC da stack em um Pod temporário e altere o binário daquele Workflow. O Pod usa o mesmo usuário do runner para acessar o arquivo.

```bash
hub apply -f - <<'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: simulator-plan-diagnostic
  namespace: argo-workflows
spec:
  automountServiceAccountToken: false
  restartPolicy: Never
  securityContext:
    runAsUser: 10001
    runAsGroup: 10001
    fsGroup: 10001
  containers:
    - name: diagnostic
      image: terraformstack-runner:local
      imagePullPolicy: IfNotPresent
      command: [sleep, "600"]
      volumeMounts:
        - name: state
          mountPath: /state
  volumes:
    - name: state
      persistentVolumeClaim:
        claimName: simulator-network
YAML
hub -n argo-workflows wait pod/simulator-plan-diagnostic \
  --for=condition=Ready --timeout=90s
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
  -o 'custom-columns=GEN:.metadata.generation,WORKFLOW:.status.execution.workflow,GATILHO:.status.execution.trigger,CONDICAO:.status.conditions[?(@.status=="True")].type,DRIFT:.status.lastDriftTime'
```

Sem novo commit ou sync, aguarde `DriftCheck` na mesma revisão, a condição `DriftDetected` ou o horário de drift persistido e o retorno do CIDR ao baseline. Confira o Workflow periódico e a próxima checagem sem mudanças. O intervalo desta solicitação é 45 segundos. Com o runtime consolidado, confira também `status.driftPending` durante a remediação, `status.lastSuccessTime` após `Ready` e a métrica `terraformstack_drift_pending` pelo endpoint da Etapa 05. O campo pode estar ausente antes do primeiro drift reportado; a condição `DriftDetected` pode ser transitória.

Com outro Workflow aguardando aprovação, sincronize a mesma revisão e confira que o nome da execução permanece. Em seguida, confira a troca de líder com o gate ainda aberto.

```bash
run_before="$(workflow)"
holder="$(hub -n platform-system get lease terraformstack-controller.platform.example.io \
  -o jsonpath='{.spec.holderIdentity}')"
hub -n platform-system delete pod "${holder%%_*}"
hub -n platform-system get lease terraformstack-controller.platform.example.io -w
```

Quando a Lease mostrar outro holder, encerre a consulta com Ctrl+C e compare `workflow` com `run_before`. O mesmo Workflow deve continuar aguardando aprovação.

No ensaio de concorrência, abra um gate e sincronize uma nova revisão antes de aprovar. Confira o Workflow antigo terminado e um novo hash e Workflow. Para testar a chegada de uma revisão durante o apply, deixe o próximo SHA pronto no painel de Sync antes de aprovar o gate atual. Assim que a consulta abaixo confirmar o início do apply, sincronize o SHA e acompanhe as duas execuções. A nova começa após o término do apply e do verify anteriores.

```bash
hub -n platform-system wait terraformstack/simulator-network \
  --for='jsonpath={.status.execution.phase}=Applying' --timeout=180s
```

Para observar retry operacional, sincronize outra revisão de teste com o CIDR de baseline. Aguarde o controller publicar o nome do Workflow dessa revisão. Encurte o deadline do Pod de `plan` para provocar uma falha. A exclusão de um Pod ainda pendente pode fazer o Argo recriá-lo na mesma tentativa.

```bash
run="$(workflow)"
for attempt in {1..90}; do
  plan_pod="$(hub -n argo-workflows get pods \
    -l "workflows.argoproj.io/workflow=$run" \
    --field-selector 'status.phase!=Succeeded,status.phase!=Failed' -o name |
    awk '/-plan-/{print; exit}')"
  [[ -n "$plan_pod" ]] && break
  sleep 1
done
if [[ -n "$plan_pod" ]]; then
  hub -n argo-workflows patch "$plan_pod" --type merge \
    -p '{"spec":{"activeDeadlineSeconds":1}}'
else
  printf '%s\n' 'Pod de plan ativo não encontrado.'
fi
```

Confira a falha do primeiro Workflow e, após o intervalo de 10 segundos, outro Workflow com `status.execution.trigger: Retry` e `attempt: 1`.

Para testar dependência, libere `suspend` de `simulator-dependent` com a Network ainda suspensa e confira `DependencyBlocked`. Na revisão que retoma a Network, use `reconciliation.driftCheckInterval: 10m` na Network para dar tempo à primeira execução da dependente. Uma checagem periódica coloca a Network em `Reconciling` temporariamente e pode cancelar uma execução dependente antes do apply. Aguarde `Ready` nas duas stacks e confira o state e o ConfigMap próprios da dependente. Restaure o intervalo de 45 segundos após o exercício.

Para testar um ciclo, suspenda temporariamente as duas solicitações, configure cada stack como dependência da outra e entregue uma revisão com as duas ativas. O controller registra `ReconcilePending / DependencyCycle`. Retire a dependência da Network antes de continuar.

Para testar a proteção do destino, use uma revisão com um contexto diferente de `entity-a/simulator`. Observe `ReconcileBlocked / ExecutionBlocked` com diagnóstico `ContextNotAllowed`, antes de montar o state. Um vínculo divergente produz `BindingChanged` ou `StateBindingMismatch`, conforme a verificação que detectar a divergência. Preserve o PVC e restaure o perfil aprovado antes de gerar outro plano.

## 4. Exercitar destroy e limpar

Faça as alterações no manifesto das solicitações e sincronize cada SHA pela mesma Application do Simulator. [Manifesto das stacks](../control-plane/terraformstack/examples/simulator/stacks.yaml)

1. Com ambas as stacks ativas, solicite `spec.lifecycle.destroy: true` na Network. Confirme `ReconcilePending / DependencyBlocked`. Mantenha o pedido da Network e sincronize outra revisão com `destroy: true` também na dependente. Aprove o plano da dependente e aguarde `Destroyed / Verified`. Em seguida, a Network pode chegar ao gate.
2. No gate da Network, use `Reject` e confira que o ConfigMap permanece. Na próxima revisão, mantenha o destroy da Network e suspenda a dependente já destruída com `suspend: true`, retirando seu `dependsOn`. Sincronize para gerar outro plano da Network. Aprove o hash, aguarde `Destroyed` e confirme a ausência do ConfigMap e a permanência do CR.
3. Recrie a stack por uma revisão com `destroy: false` e mantenha `destroyResourcesOnDeletion: false`. Depois de `Ready`, exclua o CR e confira a retenção do ConfigMap e do PVC. Sincronize novamente para retomar o mesmo state.
4. Com o destroy da dependente verificado, remova o CR dela antes do teste de cleanup da Network. Sincronize `destroyResourcesOnDeletion: true` na Network e confirme o campo aplicado antes de excluir. Use `--wait=false` para observar `deletionTimestamp` e o Workflow enquanto o finalizer aguarda. Aprove o plano e confirme a remoção do ConfigMap e do CR após a verificação.

Para o passo 3, aguarde a retenção e confira os recursos.

```bash
hub -n platform-system delete terraformstack simulator-network --timeout=90s
network
hub -n argo-workflows get pvc simulator-network
```

Para o passo 4, após habilitar o teardown e confirmar a revisão, solicite a exclusão e acompanhe o gate.

```bash
hub -n platform-system delete terraformstack simulator-network --wait=false
hub -n platform-system wait terraformstack/simulator-network \
  --for=condition=ExecutionWaiting --timeout=180s
stack -o jsonpath='{.metadata.deletionTimestamp}{"\n"}{.status.result.details.planHash}{"\n"}'
```

Depois da aprovação, aguarde a conclusão do teardown.

```bash
hub -n platform-system wait terraformstack/simulator-network --for=delete --timeout=180s
hub -n terraformstack-simulator get configmap simulator-network --ignore-not-found
```

Limpe primeiro as solicitações. Exclua apenas os Workflows e PVCs do Simulator depois que as execuções concluírem e o state puder ser descartado. Uma negativa no destroy durante exclusão mantém o CR pendente com finalizer. Nesse caso, preserve as evidências e trate a recuperação daquela stack no mesmo Hub.

```bash
hub -n argo-workflows delete workflows \
  -l 'platform.example.io/stack-name in (simulator-network,simulator-dependent)'
hub -n argo-workflows delete pvc simulator-network simulator-dependent
hub -n argo-workflows delete rolebinding "$approver"
hub -n terraformstack-simulator delete serviceaccount "$approver"
```

## O que observar

- `status.observedGeneration`, `resolvedSource` e `execution.workflow` ligam intenção, commit e execução. O GitOps mostra a sincronização do CR. Leia a condição pelo par `type` e `reason`, como `ExecutionWaiting / Waiting`. `lastCheckTime` e `lastDriftTime` registram a última tentativa concluída e a última detecção de drift. [Condições e diagnóstico](../docs/flows/terraformstack.md#6-status-e-diagnóstico)
- `status.result` guarda resultado, diagnósticos e detalhes como a decisão e o hash do plano. `status.claim` relaciona intenção, Workflow e autorização. O log de auditoria do Hub confirma a identidade que retomou o gate. As evidências são o status, o ConfigMap e a ausência de mudança em um segundo plan.
- Um Workflow pode terminar em `Succeeded` após registrar `Deny` ou uma rejeição. Nesses casos, o CR exibe `ReconcileBlocked` e o ConfigMap preserva o último valor aplicado.
- Consulte `hub -n platform-system describe terraformstack simulator-network` para ver os Events ainda retidos pela API. Correlacione logs do controller com o Workflow e o hash da intenção, sem publicar tokens ou payloads. [Métricas e diagnóstico](../control-plane/terraformstack/controller/README.md#observabilidade-operacional).
- Guarde revisão, ID ou digest de imagem e resultados do exercício. O PVC contém os planos e states locais.

## 5. Executar o mesmo fluxo com OpenTofu

O runner da Etapa 05 inclui o OpenTofu. Use a stack dependente e um PVC vazio para manter os ensaios separados. Conclua o destroy Terraform da dependente, exclua seu CR e remova seus Workflows concluídos antes de excluir o PVC. Os Pods desses Workflows podem manter a proteção do PVC ativa. Sincronize **terraformstack-controller** para recriar o PVC vazio.

Em uma revisão de teste, defina `iac` como `tofu` no perfil do Simulator. Suspenda a Network, remova o `dependsOn` da dependente e ative a dependente. Sincronize **terraformstack-workflows** e depois **terraformstack-simulator** com essa revisão. Confira no Workflow o primeiro apply, a mudança de CIDR com aprovação e o destroy explícito aprovado da dependente. Use `status.resolvedSource.binding` para correlacionar as tentativas no mesmo state. [Perfil de execução](../workflow-plane/argo-workflows/terraformstack/simulator.json) · [Runner e isolamento](../workflow-plane/README.md#binário-de-iac-no-simulator)

Após o destroy verificado, remova os recursos do ensaio conforme a seção anterior e sincronize o WorkflowTemplate da revisão padrão com `iac: terraform`.

## Evidências de aceite

### Validação automatizada do runtime consolidado

Os PRs #60–#62 acrescentaram logging, seleção de integrações, autorização opcional e observabilidade por CR. A CI do [PR #62](https://github.com/DavidFerreira21/platform-engineering-poc/pull/62) passou nas validações do controller e do runner. Os cenários de múltiplos aliases e `authorization: false` têm cobertura automatizada; a integração de referência permanece com handshake habilitado. As evidências de execução real abaixo correspondem às revisões registradas, anteriores a essa consolidação. Não representam uma nova execução de ponta a ponta do runtime atualizado.

### Aceite do contrato na revisão do PR #58

O aceite usou o Hub `kind` `platform-management` com as revisões da refatoração. Depois do merge, repetimos as Etapas 01 a 06 com a `main` em `ce89b5f`. Os cenários versionaram somente dados do Simulator e as mudanças aplicadas foram ConfigMaps locais. [PR #58](https://github.com/DavidFerreira21/platform-engineering-poc/pull/58).

| Cenário | Resultado observado |
| --- | --- |
| GitOps e drift | O primeiro apply criou o baseline. O plan periódico retornou `NoChanges` sem claim. A alteração externa para `10.83.0.0/16` foi detectada e corrigida no commit já fixado. |
| Política e UI | O Resume autorizado aplicou `10.82.0.0/16` e apareceu na auditoria. Rejeição, hash incorreto, negativa do OPA e erro de Rego preservaram esse valor. A identidade somente leitura não tinha permissão de atualizar Workflows. |
| Outputs e integridade | Outputs ausentes, inválidos e contraditórios terminaram em `InvalidExecutionContract`. A adulteração do binário falhou antes do claim. O retry usou a mesma referência e o mesmo vínculo e exigiu outra aprovação. |
| Idempotência e cancelamento | O sync repetido e a troca de líder conservaram o gate. Um novo commit terminou a espera obsoleta antes de criar o próximo Workflow. Uma revisão que chegou após o claim aguardou o apply e a verificação anteriores. A admissão recusou claim sem autorização. |
| Origem e destino | O Workflow recusou repositório e contexto fora do perfil antes de preparar o volume. Uma mudança de vínculo foi recusada com `BindingChanged` e o ConfigMap foi preservado. |
| Dependências e destroy explícito | O controller detectou o ciclo, aguardou a Network ficar pronta e provisionou a dependente com state próprio. O destroy da Network só alcançou o gate depois do destroy verificado da dependente. A rejeição preservou o ConfigMap. A aprovação subsequente removeu o ConfigMap e manteve o CR em `Destroyed`. |
| Retain, vínculo do PVC e finalizer | A exclusão padrão preservou o ConfigMap e o PVC. Recriar o CR com um perfil incompatível retornou `StateBindingMismatch`. Restaurar o perfil retomou o state com `NoChanges`. O destroy opt-in na exclusão removeu o ConfigMap e liberou o finalizer após aprovação e verificação. |
| OpenTofu 1.13.9 | Com PVC novo, o Simulator passou por plan JSON, OPA, apply automático, atualização de CIDR aprovada, revalidação, verify e destroy aprovado. O backend local recusou um plan concorrente durante apply em testes separados com Terraform e OpenTofu. |

### Revisão anterior

Em 26/11/2026, recriamos o Hub `platform-management` e seguimos as Etapas 01 a 06. O WorkflowTemplate foi sincronizado a partir de `e7245bd`. Para os testes de dependência e teardown, reconstruímos o controller com a correção `77e4b77`. O Kubernetes registrou o ImageID `sha256:17fb6897534d9569ab8333c09581468eed856306eca44dc860a9fa1a5ec5d6af` no controller e `sha256:8e9155c8134166b516c7d27bc2eab74c40aec79f961a19fb06d185f57eb46a67` no runner.

| Cenário | Resultado observado |
| --- | --- |
| GitOps e OPA | `AutoApply` criou o ConfigMap com `10.81.0.0/16`. O Resume autorizado aplicou `10.82.0.0/16` e registrou o aprovador na auditoria. Token somente leitura, rejeição, hash incorreto, `Deny` e `PolicyError` preservaram o valor já aplicado. |
| Integridade, retry e drift | Plano adulterado e deadline do Pod de plan dispararam novas tentativas. Após a volta ao baseline, a checagem periódica detectou `10.83.0.0/16` e restaurou o CIDR na mesma revisão. |
| Concorrência | Sync repetido e troca de líder conservaram o gate. Uma revisão nova cancelou um gate obsoleto. Quando o apply já tinha iniciado, a nova revisão aguardou a conclusão do verify anterior. |
| Dependência e teardown | A dependente chegou a `Ready` com ConfigMap e state próprios. O destroy da Network foi bloqueado até a dependente ser destruída. O destroy explícito aprovado removeu o ConfigMap e manteve o CR. |
| Exclusão | A exclusão padrão preservou ConfigMap e PVC. O sync seguinte retomou o state com `NoChanges`. A exclusão opt-in removeu ConfigMap e CR após aprovação e verificação. O Hub permaneceu disponível após a limpeza. |

[← Etapa 05](05-terraformstack-runtime.md) · [Roteiro](../docs/learning-roadmap.md) · [Etapa 07 →](07-crossplane-core.md)
