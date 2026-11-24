# TerraformStack Flow — Platform Engineering POC

[← Arquitetura](../architecture.md) · [GitOps Flow](gitops.md) · [ADR-0006](../adr/0006-terraformrun-and-application-ci.md) · [Roadmap](../learning-roadmap.md)

> **Estado:** desenho arquitetural aprovado para documentação; implementação incremental, conforme o roadmap.  
> **Implementado/documentado até aqui:** contrato `TerraformStack` v1alpha1 e seu guia da Etapa 05.  
> **Previsto:** controller e execução simulada (Etapa 06); execução real com aprovação (Etapa 09); detecção de drift (Etapa 18); reconciliação por políticas (Etapa 19).
>
> **Nota sobre status:** os nomes de estados apresentados neste documento descrevem o **lifecycle lógico esperado**, não campos ou enums já implementados no CRD. O schema atual expõe `status.conditions`; a estrutura definitiva de `phase`, condições e referências será consolidada durante a implementação do controller.

## 1. Visão geral

O `TerraformStack` representa **infraestrutura desejada**, e não uma execução isolada de Terraform. Seu propósito é integrar provisionamento, atualização, verificação de drift e reconciliação governada ao modelo declarativo da plataforma.

O Git mantém a intenção versionada; Argo CD sincroniza o objeto Kubernetes; o TerraformStack Controller governa seu lifecycle; Argo Workflows orquestra plan, gates e apply; Terraform modifica recursos na **conta AWS de destino**.

![Fluxo TerraformStack — mudança desejada, aprovação, drift e lifecycle](./terraformstack-flow.png)

[Visualizar o diagrama em tamanho completo](./terraformstack-flow.png)

> O arquivo de imagem `terraformstack-flow.png` será adicionado separadamente. O desenho ilustra uma **topologia lógica**: conta da plataforma e contas AWS dos Environments são fronteiras diferentes. Na POC inicial, o Control Plane executa em `kind`; isso não implica que um EKS de plataforma ou uma conexão entre VPCs já existam.

### Resumo do caminho principal

```text
Git / PR revisado
  → Argo CD central
  → Kubernetes API / TerraformStack (CR)
  → TerraformStack Controller
  → Argo Workflows (plan, policy, gate, apply)
  → [Auto Apply | Approval Required | Deny/Block]
  → Terraform Apply (quando autorizado)
  → Verify
  → TerraformStack.status
  → AWS na conta da Entity / Environment autorizado
```

Há **dois gatilhos operacionais**: (a) criar/alterar a intenção por Git; (b) verificar periodicamente uma stack existente ou solicitar nova verificação controlada. Ambos passam por validações; qualquer alteração efetiva exige uma decisão de política e autorização adequadas.

## 2. Responsabilidades e fronteiras

| Componente | Responsabilidade |
| --- | --- |
| **Git / PR** | Definir versão do módulo, configuração e destino lógico aprovados; manter auditoria de alterações desejadas |
| **Argo CD central** | Entregar o **CR** `TerraformStack` do Git à API Kubernetes do Hub; não executar Terraform |
| **Kubernetes API** | Validar o objeto segundo o CRD e disponibilizar `spec` e `status` |
| **TerraformStack Controller** | Observar mudanças, resolver o destino autorizado, evitar concorrência, coordenar execuções e atualizar status |
| **Argo Workflows** | Orquestrar plan, avaliação de política, pausa para aprovação, apply e verificação |
| **Terraform / executor do Workflow** | Calcular e aplicar mudanças usando identidade temporária, backend e módulo autorizados |
| **Policy Engine (previsto)** | Classificar cada plano em **Auto Apply**, **Require Approval** ou **Deny/Block** |
| **Backend de state e artefatos** | Armazenar state/lock e planos imutáveis em local privado, fora do Git |
| **AWS de destino** | Hospedar VPC, IAM, EKS e demais recursos provisionados para a Entity / Environment |

**Fronteiras AWS:** a conta da plataforma hospeda os componentes de controle/execução quando migrarmos do `kind`. Cada Environment (`uat`, `prod`) corresponde a uma **conta AWS de destino distinta** da Entity. Uma stack resolve a conta/região/role por um cadastro confiável; o usuário **não escolhe livremente credenciais ou account IDs por manifesto**.

**Separação CRD e CR:** a Etapa 05 registra o tipo `TerraformStack` na API por meio do CRD. Nos fluxos operacionais, o Argo CD sincroniza **instâncias** desse tipo. O controller da Etapa 06 passa a observá-las.

## 3. Contrato declarativo

A definição inicial está em `control-plane/terraformstack/crd/terraformstack.yaml`, com API `platform.example.io/v1alpha1` e escopo `Namespaced`.

Exemplo da Etapa 05 (dados fictícios):

```yaml
apiVersion: platform.example.io/v1alpha1
kind: TerraformStack
metadata:
  name: network-uat
  namespace: platform-system
spec:
  module:
    name: lab-network
    version: "1.0.0"
  target:
    entity: entity-a
    environment: uat
```

O contrato será estendido nas etapas correspondentes para selecionar a revisão dos YAMLs de configuração (`spec.configuration.revision`), resolver identidade, backend e módulo aprovado, e relatar o lifecycle. **O exemplo acima não implica execução AWS** enquanto o controller e as dependências não estiverem implementados.

Os valores de módulos previstos em `config/` têm precedência **global → Entity → Environment**. O destino AWS real é resolvido a partir de `foundation/config/<entity>/<env>/environment.yaml`; Git e Terraform state não devem conter credenciais.

`TerraformRun` fica reservado como eventual nome de uma execução individual. Por enquanto, um Argo Workflow é suficiente para representar cada execução: **não criar um segundo CRD sem necessidade**.

## 4. Fluxo A — provisionamento e mudança desejada

1. O Platform Team ou um solicitante autorizado altera o manifesto `TerraformStack` e, quando aplicável, a configuração versionada, submetendo PR.
2. Validações de schema, política e ownership antecedem o merge. **A aprovação do PR não substitui a aprovação do plano Terraform.**
3. O Argo CD central sincroniza o CR com a API Kubernetes. `Synced` comprova a entrega do manifesto, **não** que a AWS esteja convergida.
4. O TerraformStack Controller identifica a `generation`/revisão, valida módulo, destino, permissões, estado da stack e concorrência, e inicia o Workflow adequado.
5. Argo Workflows gera um `terraform plan`, registra seu resumo e o submete à avaliação de políticas.
6. A política determina **aplicação permitida**, **aprovação humana obrigatória** ou **bloqueio**.
7. Somente em caminho autorizado, o Workflow aplica **o mesmo plano validado/aprovado**, verifica o resultado e atualiza `TerraformStack.status`.

A execução é **event-driven e governada**; um reconcile do Kubernetes não deve gerar `terraform apply` indiscriminadamente. A implantação inicial do controller poderá simular esse fluxo sem realizar chamadas AWS.

### 4.1 Três caminhos após Policy Evaluation

| Decisão | Quando | Próximo passo | Resultado esperado |
| --- | --- | --- | --- |
| **Auto Apply** | Regra versionada permite explicitamente a ação, o destino e o risco | Apply → Verify | `Ready` quando validado |
| **Require Approval** | Mudança sensível ou política exige análise humana | Pausa no Argo Workflows → aprovação vinculada ao plano → Apply → Verify | `Ready` ou falha documentada |
| **Deny / Block** | Operação proibida, destino inválido ou violação de política | **Não executar apply** | `Blocked` com motivo verificável |

**Limite do roadmap:** até a Etapa 09, todo `apply` real exige **aprovação humana**. A decisão automática passa a ser implementada/validada somente na **Etapa 19**, com regras explícitas e testadas; o desenho mostra a **arquitetura-alvo**, não um Auto Apply já liberado.

Exemplos ilustrativos de decisão:

- **Auto Apply (futuro):** alteração de tag permitida em UAT, após verificação da política e do impacto.
- **Require Approval:** alteração de rede sensível, mudanças de maior risco ou operações definidas como críticas.
- **Deny / Block:** destino não autorizado, recurso fora dos padrões, exposição insegura ou operação proibida.

Ao bloquear, a plataforma deve retornar uma explicação concreta, como `PolicyViolation: alteração não permitida para o Environment solicitado`, com identificação da regra. **Bloqueio por política não é a mesma coisa que rejeição humana**, e os dois motivos devem permanecer distintos na auditoria.

### 4.2 Approval Gate — plano imutável

O canal escolhido é a **Argo Workflows UI**, com `suspend: {}` e autenticação/autorização individual para os aprovadores. Não há prazo fixo de aprovação configurado no `suspend` nesta fase.

A aprovação deve ficar vinculada à **identidade do aprovador**, Workflow, hash do plano, geração do CR, revisão Git/configuração, conta/região, backend/state e identidade do destino. O arquivo binário de plano fica em armazenamento privado e possui validade operacional controlada.

Após o `resume`, um passo separado valida a autorização, os vínculos e o estado atual antes de executar `terraform apply <plano>`. Um simples `resume` não é evidência suficiente. Se a validação falhar, não há apply: gera-se novo plano e exige-se nova decisão.

**Aprovação humana pode permanecer pendente sem um timer de expiração do Workflow; isso não implica que um plano antigo ou inválido possa ser aplicado.** A revalidação e a política de validade do artefato continuam obrigatórias.

### 4.3 Resultados e falhas

- **Plan sem diferenças:** registrar `NoChanges` e encerrar a verificação, sem apply desnecessário.
- **Aprovado:** aplicar exatamente o plano autorizado, verificar e atualizar status.
- **Rejeitado por aprovador:** registrar a decisão e encerrar sem apply; não classificar como violação da policy.
- **Bloqueado por política:** registrar regra e motivo, encerrar sem apply.
- **Erro, cancelamento ou perda de validade:** não reutilizar plano nem aprovação em outra execução; registrar condição e evidência.
- **Deleção/destroy:** remover o CR do Git **não dispara destroy automaticamente**; operações destrutivas exigem fluxo e autorização próprios.

## 5. Fluxo B — drift detection

Depois da conclusão do provisionamento, o `TerraformStack` permanece ativo como representação do estado desejado.

1. O scheduler do controller agenda uma verificação de stack existente (ou recebe uma solicitação manual autorizada).
2. O Workflow executa uma verificação Terraform controlada, sem apply.
3. O controller compara e classifica o resultado, diferenciando **mudança declarada ainda pendente** de **divergência externa** entre configuração, state e AWS.
4. **Sem drift:** mantém `Ready` e registra data e evidência da última checagem.
5. **Com drift:** registra `DriftDetected`, o resumo não sensível da diferença e a referência do plan.

A **Etapa 18 é somente detecção e alerta**. Não se corrige automaticamente drift nessa etapa.

## 6. Fluxo C — drift reconciliation

A Etapa 19 acrescenta decisão e remediação governadas por política:

```text
DriftDetected
  → Policy Evaluation
     ├─ Auto Reconcile permitido → Apply → Verify → Ready
     ├─ Require Approval → Argo Workflows UI → Apply → Verify → Ready
     └─ Deny / Block → ReconcileBlocked + motivo
```

O mesmo cuidado com plano aprovado, identidade de destino, proteção do state e revalidação se aplica à correção de drift. **Nem todo drift deve ser remediado automaticamente**: remoções, substituições, mudanças de rede e atributos críticos podem exigir autorização ou ser bloqueados.

O controller não deve interpretar uma ação genérica `create` ou `update` como segura sem analisar recurso, atributos, ambiente e impacto. A detecção deve ser idempotente; apenas uma decisão autorizada pode gerar apply.

## 7. Lifecycle e status do TerraformStack

O CRD atual possui `status.conditions`. Os nomes abaixo são **estados conceituais para o diagrama**, a serem mapeados na implementação para `phase` e/ou condições Kubernetes; não representam enum já criado.

### 7.1 Mudança desejada

| Estado lógico | Significado |
| --- | --- |
| `Pending` | Solicitação observada, aguardando validação/execução |
| `Planning` | Preparação e geração do plano |
| `PolicyEvaluating` | Classificação das mudanças |
| `ApprovalRequired` | Aguardando aprovação de um plano específico |
| `Blocked` | Regra de política negou a operação |
| `Applying` | Apply autorizado em curso |
| `Verifying` | Verificação pós-apply e coleta de evidências |
| `Ready` | Estado desejado confirmado pela última verificação aplicável |
| `Failed` / `Cancelled` | Execução terminou com erro ou foi cancelada, com motivo registrado |

**`Approved`** e **`NoChanges`** são resultados/eventos que podem ser representados em condições ou histórico, sem exigir obrigatoriamente uma fase persistente.

### 7.2 Drift e reconciliação

| Estado lógico | Significado |
| --- | --- |
| `Ready` | Sem divergência conhecida na última verificação |
| `DriftCheckRunning` | Checagem periódica/manual em execução |
| `DriftDetected` | Diferença externa identificada |
| `ReconcilePending` | Aguardando decisão/política ou aprovação |
| `ReconcileBlocked` | Remediação negada pela política |
| `Reconciling` | Correção autorizada em execução |
| `DriftCorrected` | Evento de correção concluída; retorno a `Ready` após verificar |
| `Failed` | Verificação/remediação falhou |

O `status` deverá distinguir **estado atual da stack** e **última execução**: uma falha de drift check não significa automaticamente que a infraestrutura deixou de existir, e um Workflow bem-sucedido não garante que toda divergência externa tenha sido corrigida.

Metadados de status propostos: `observedGeneration`, revisão Git e de configuração, referência da última execução, tipo de gatilho, última checagem, resumo de plan, decisão e justificativa de política, condição de aprovação, referência do hash/artefato e observações de drift. **Sem tokens, segredos, plano integral ou conteúdo de state no CR.**

## 8. Mudanças durante execuções, retry e cancelamento

As seguintes regras fazem parte do comportamento desejado:

| Situação | Comportamento |
| --- | --- |
| **Spec muda enquanto a aprovação está pendente** | Cancelar Workflow antigo, invalidar seu plano/aprovação e gerar novo plan para a nova geração |
| **Spec muda durante apply já iniciado** | Deixar o apply em andamento concluir; depois avaliar a nova geração em outro ciclo, sem interrupção parcial |
| **Workflow falha ou é cancelado** | Registrar estado/causa; nova execução requer solicitação controlada e gera novo plano |
| **Solicitação manual de nova execução** | Usar uma **annotation no TerraformStack** como sinal de retry/reconcile, com autorização e idempotência |
| **Reconcile repetido sem mudança** | Não disparar duplicatas nem executar apply indevido |
| **Execuções concorrentes sobre o mesmo state** | Bloquear/serializar por stack e respeitar o lock efetivo do backend |

A forma exata e o nome da annotation serão definidos na implementação do controller. Uma nova solicitação manual **não** reaproveita autorização antiga.

## 9. Segurança, state, auditoria e isolamento

- **Destino confiável:** `Entity + Environment` resolve conta, região, role e backend a partir da Foundation, com validação de IAM/STS antes da execução.
- **Isolamento:** UAT e PROD são contas AWS distintas; a plataforma não hospeda os recursos provisionados para esses ambientes.
- **State:** segregado por stack/destino, armazenado fora do Git, com lock, criptografia e permissões mínimas.
- **Plano:** artefato privado e imutável, hash e referências auditáveis; não publicar seu conteúdo em `status` ou logs compartilhados.
- **Aprovação:** identidade verificável e autorização por operação/destino; aprovação de PR, `resume` e autorização do apply são controles distintos.
- **Políticas:** versionadas, testáveis e aplicadas antes de qualquer ação; decisões e razões devem permanecer auditáveis.
- **Observabilidade:** correlacionar CR, generation, Workflow, policy decision, executor, plan, alvo AWS, aprovador e resultado.
- **Custos e remoção:** provisionamento AWS exige autorização explícita, validação de custos e plano de cleanup. Aprovar este documento **não** autoriza executar `apply` ou `destroy`.

## 10. Relação com GitOps e GitOps Bridge

O [GitOps Flow](gitops.md) explica a entrega do manifesto declarativo pela plataforma. Este documento concentra o **lifecycle do TerraformStack** — da solicitação ao resultado e às verificações subsequentes.

A publicação e distribuição dos **outputs selecionados** do Terraform, por exemplo para AWS Systems Manager Parameter Store, pertencem ao documento planejado `docs/flows/gitops-bridge.md`. O desenho do TerraformStack poderá indicar uma etapa posterior de publicação de outputs, mas **não pressupõe** que Parameter Store registre automaticamente clusters no Argo CD.

## 11. Estratégia incremental no roadmap

| Etapa | Entrega observável |
| --- | --- |
| **05** | CRD `TerraformStack`, schema e `status.conditions` disponíveis, com dry-run local |
| **06** | Controller, WorkflowTemplate e executor **simulados** no `kind` |
| **07–08** | Registro de UAT, validação de identidade, backend, state e autorização |
| **09** | Primeiro plan/apply real de rede UAT, com aprovação humana de plano imutável |
| **10–11** | Reutilização do fluxo para EKS Management e Non-PCI, um cluster por vez |
| **18** | Drift detection, sem remediação |
| **19** | Política de Auto Apply / Require Approval / Deny e correção governada |
| **20** | Testes de isolamento UAT/PROD e PCI/Non-PCI |

A execução inicial permanece em `kind`. As etapas AWS serão realizadas somente com destino, credenciais, custos e autorização verificados. O roadmap só será revisado em conjunto depois de documentarmos os três fluxos arquiteturais.

## 12. Referências

- [GitOps Flow](gitops.md)
- [ADR-0006 — TerraformStack, aprovação e reconciliação](../adr/0006-terraformrun-and-application-ci.md)
- [ADR-0003 — Ownership e reconciliação](../adr/0003-reconciliation-and-gitops.md)
- [Contrato TerraformStack v1alpha1](../../control-plane/terraformstack/crd/terraformstack.yaml)
- [Solicitação de exemplo](../../control-plane/terraformstack/demo/terraformstack.yaml)
- [Laboratório da Etapa 05](../../labs/05-terraformstack-crd.md)
- [Roadmap da POC](../learning-roadmap.md)

**Próximo documento:** `docs/flows/gitops-bridge.md`.
