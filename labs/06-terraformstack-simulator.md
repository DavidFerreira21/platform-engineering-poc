# Etapa 06 — Terraform Stack Simulator de ponta a ponta

## Objetivo

Reconciliar uma stack Terraform real no Hub EKS e comprovar pelo status do `TerraformStack` as decisões `AutoApply`, `RequireApproval` e `Deny`, o drift do ambiente e o teardown governado.

## O que será feito

O Argo CD sincroniza a intenção em Git. O controller abre um Workflow e a pipeline fixa o commit de código e configuração. O Terraform grava o mapa efetivo de Network no campo `network.json` de um ConfigMap no namespace `terraformstack-simulator`. O exercício usa a hierarquia de Configuration de `entity-a/uat` e identidade Kubernetes do executor.

## Por que fizemos assim

Este lab exercita a integração de referência `terraformstack`, com `contract.authorization: true`. Não altere esse campo para `false` no template atual: ele depende do handshake. O mesmo caminho de reconciliação atende o primeiro plan, as mudanças declaradas e a observação periódica do ambiente. Mantemos o state em um PVC por stack, com locking do backend local, e o plano binário em um diretório por execução. O OPA avalia o JSON derivado do plano. Antes do apply, a pipeline confere o hash do binário e compara o efeito previsto com um novo plan. A função `claim()` confere a intenção e a autorização vigentes no CR e registra `status.claim` por atualização condicional. O apply usa o binário salvo. [Autorização final do apply](../docs/flows/terraformstack.md#autorização-final-do-apply)

O perfil local da pipeline autoriza os dois nomes do exercício, a raiz Terraform do repositório e o destino lógico `entity-a/simulator`. O Simulator lê Configuration com o perfil de dados `entity-a/uat`. O código Terraform das revisões aceitas é mantido pelo Platform Team. O Workflow usa o perfil `simulator`, composto de `common` e `simulator`, em um pacote imutável; o status publica a decisão, o hash do plano e o SHA-256 do pacote. [Contrato](../docs/flows/terraformstack.md) · [Política OPA](../workflow-plane/policy-engine/README.md)

> O primeiro apply do Simulator foi validado no Hub EKS. Os critérios de aprovação, negativa, retry e destroy permanecem para a rodada completa no EKS.

## Pré-requisitos

Concluir o aceite do runtime na Etapa 05. Usar o perfil do Simulator, armazenamento de teste e uma revisão acessível ao Hub EKS. Conferir o destino e o diff antes de sincronizar solicitações.

## 2. Exercitar as decisões OPA pelo GitOps

Validar `AutoApply`, `RequireApproval` e `Deny`. No caso manual, conferir identidade do aprovador, auditoria da decisão e vínculo com o hash do plano aprovado. Confirmar que a política de admissão limita o executor ao claim da tentativa vigente.

## 3. Observar drift e reconciliação

Alterar apenas o ConfigMap de teste e observar a checagem periódica. Correlacionar detecção, política aplicada, referência Git fixada e status da tentativa. Validar retry após uma falha operacional controlada.

## 4. Exercitar destroy e limpar

Autorizar o destroy do recurso de teste, conferir o plano e validar a remoção do ConfigMap. Limpar os recursos exclusivos do exercício e preservar as dependências compartilhadas.

## O que observar

Correlacionar revisão Git, Workflow, hash do plano, claim e condições do TerraformStack. Conferir logs, Events e métricas sem publicar planos, tokens ou state.

## Evidências de aceite

### Primeiro apply no Hub EKS

Em 28/11/2026, aplicamos os CRs de teste pela API. O runner foi baixado do ECR privado por digest e o PVC de `simulator-network` ficou `Bound` com a StorageClass do EKS Auto Mode. O Workflow concluiu em `Succeeded`, o TerraformStack informou `Verified` e o ConfigMap de teste foi criado.

### Uso do PVC em checagens periódicas

A sequência de checagens periódicas mostrou que checkouts e arquivos de plano acumulados no PVC podiam esgotar seus 1 GiB. A pipeline passou a limpar o diretório da tentativa ao concluir os passos. O state e o marcador de binding ficam na raiz do PVC. Em volumes criados antes dessa mudança, identificar Workflows terminais antes de remover seletivamente seus diretórios. Preservar os arquivos das falhas enquanto o diagnóstico estiver em andamento.

Após remover somente os diretórios das tentativas bem-sucedidas do teste, a ocupação do PVC caiu para 3%. Repetimos a reconciliação com a mesma revisão Git. O Workflow retornou `NoChanges` e o step `cleanup` ficou `Succeeded`. Em seguida, alteramos o CIDR fictício do ConfigMap para simular drift. O OPA retornou `AutoApply` e o Simulator restaurou o valor de referência.

### Validação automatizada do runtime consolidado

Os PRs #60–#62 acrescentaram logging, seleção de integrações, autorização opcional e observabilidade por CR. A CI do [PR #62](https://github.com/DavidFerreira21/platform-engineering-poc/pull/62) passou nas validações do controller e do runner. Os cenários de múltiplos aliases e `authorization: false` têm cobertura automatizada. A integração de referência mantém o handshake habilitado. As subseções dos ensaios iniciais abaixo identificam as revisões e os resultados de cada execução.

### Aceite do contrato na revisão do PR #58

O ensaio inicial usou as revisões da refatoração. Depois do merge, repetimos o procedimento daquela revisão com a `main` em `ce89b5f`. Os cenários versionaram somente dados do Simulator e as mudanças aplicadas foram ConfigMaps locais. [PR #58](https://github.com/DavidFerreira21/platform-engineering-poc/pull/58).

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

Em 26/11/2026, recriamos o cluster do ensaio inicial e seguimos o procedimento daquela revisão. O WorkflowTemplate foi sincronizado a partir de `e7245bd`. Para os testes de dependência e teardown, reconstruímos o controller com a correção `77e4b77`. O Kubernetes registrou o ImageID `sha256:17fb6897534d9569ab8333c09581468eed856306eca44dc860a9fa1a5ec5d6af` no controller e `sha256:8e9155c8134166b516c7d27bc2eab74c40aec79f961a19fb06d185f57eb46a67` no runner.

| Cenário | Resultado observado |
| --- | --- |
| GitOps e OPA | `AutoApply` criou o ConfigMap com `10.81.0.0/16`. O Resume autorizado aplicou `10.82.0.0/16` e registrou o aprovador na auditoria. Token somente leitura, rejeição, hash incorreto, `Deny` e `PolicyError` preservaram o valor já aplicado. |
| Integridade, retry e drift | Plano adulterado e deadline do Pod de plan dispararam novas tentativas. Após a volta ao baseline, a checagem periódica detectou `10.83.0.0/16` e restaurou o CIDR na mesma revisão. |
| Concorrência | Sync repetido e troca de líder conservaram o gate. Uma revisão nova cancelou um gate obsoleto. Quando o apply já tinha iniciado, a nova revisão aguardou a conclusão do verify anterior. |
| Dependência e teardown | A dependente chegou a `Ready` com ConfigMap e state próprios. O destroy da Network foi bloqueado até a dependente ser destruída. O destroy explícito aprovado removeu o ConfigMap e manteve o CR. |
| Exclusão | A exclusão padrão preservou ConfigMap e PVC. O sync seguinte retomou o state com `NoChanges`. A exclusão opt-in removeu ConfigMap e CR após aprovação e verificação. O Hub permaneceu disponível após a limpeza. |

[← Etapa 05](05-terraformstack-runtime.md) · [Roteiro](../docs/learning-roadmap.md) · [Etapa 07 →](07-crossplane-core.md)
