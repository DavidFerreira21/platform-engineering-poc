# Drift, convergência e remediação

Drift é uma divergência relevante entre o estado desejado e o estado externo observado. Nem toda mudança detectada em um plan é drift: uma alteração intencional no Git também pode produzir diferenças. Detectar, decidir e remediar são responsabilidades distintas.

## Observar antes de corrigir

Uma plataforma pode reportar drift, corrigir automaticamente diferenças de baixo risco, solicitar aprovação ou bloquear alterações. A política depende do produto, do impacto e do destino. Remediação indiscriminada pode apagar uma intervenção emergencial legítima. Alterações manuais devem ter fluxo de exceção e retorno ao estado declarado.

A frequência das verificações equilibra tempo de detecção, consumo de APIs, custo e carga do executor. Retry operacional não é sinônimo de verificação de drift: um tenta recuperar falha; o outro examina se uma intenção anteriormente reconciliada ainda converge.

## Aplicação na POC

O Simulator agenda verificações periódicas com a referência Git já resolvida. A pipeline reporta divergência externa observada no plan; o OPA determina a política de remediação. O controller não precisa executar Terraform diretamente nem conhecer o significado interno do diff.

A condição DriftDetected pode ser transitória. O status driftPending permanece quando há drift reportado ainda não confirmado como convergente; um resultado Ready limpa esse sinal, e a ausência inicial do campo não significa infraestrutura sem drift. lastCheckTime registra uma tentativa concluída, inclusive falha; não prova que uma inspeção foi bem-sucedida. Na AWS, as Etapas 28–29 validarão detecção sem correção e remediação governada.

## Leitura complementar

- [Fluxo e gatilhos do TerraformStack](../flows/terraformstack.md)
- [Simulator — exercício de drift](../../labs/06-terraformstack-simulator.md#3-observar-drift-e-reconciliação)

---

[Aula anterior](28-idempotencia-e-concorrencia.md) · [Índice](README.md) · [Próxima aula](30-configuracao-state-e-binding.md)
