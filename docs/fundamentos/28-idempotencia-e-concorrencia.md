# Idempotência, concorrência e recuperação

Idempotência significa repetir uma operação para a mesma intenção sem produzir efeitos adicionais indevidos. Não é a promessa de que uma operação distribuída ocorre exatamente uma vez. Timeouts, reinícios, mensagens repetidas e perda de confirmação são situações normais que precisam de tratamento explícito.

## Três mecanismos complementares

| Mecanismo | Proteção |
| --- | --- |
| Identidade determinística da tentativa | Evita criar execuções diferentes por repetição do mesmo agendamento |
| Exclusividade por recurso | Protege todo o lifecycle, inclusive esperas antes do apply |
| Lock do state | Protege operações concorrentes sobre um backend de IaC |

Um lock de state normalmente não permanece adquirido durante uma aprovação humana. Portanto, não substitui a exclusividade no reconciliador. E a exclusividade por CR não protege dois CRs distintos apontando para o mesmo state. Compare-and-swap (CAS) impede que duas atualizações concorrentes usem a mesma versão como se ainda fosse vigente.

## Aplicação na POC

O controller persiste a tentativa e sua identidade antes de criar o Workflow. Leader Election limita a reconciliação ativa, mas não substitui o status durável. Com autorização habilitada, controller e executor disputam a versão do recurso ao revogar autorização ou registrar claim. Após o claim, uma intenção nova aguarda a conclusão.

Sem autorização, não existe limite atômico conhecido para o início da alteração. O controller mantém a execução até terminar. Antes da submissão registra launchAttempted; se a criação ficar ambígua ou o Workflow desaparecer, bloqueia replay automático. Recuperação exige verificar a execução, os Pods e o state, em vez de assumir que nada aconteceu.

## Leitura complementar

- [Fluxo do TerraformStack](../flows/terraformstack.md)
- [Kubernetes — conceitos da API e resourceVersion](https://kubernetes.io/docs/reference/using-api/api-concepts/)

---

[Aula anterior](27-contratos-e-integracoes.md) · [Índice](README.md) · [Próxima aula](29-drift-e-convergencia.md)
