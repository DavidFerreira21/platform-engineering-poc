# Observabilidade operacional e saúde dos recursos

Um produto de plataforma precisa informar tanto a saúde do seu motor quanto a saúde dos recursos que administra. Um controller pode estar saudável enquanto várias solicitações estão bloqueadas. Da mesma forma, uma falha interna pode impedir a atualização do status, fazendo um CR parecer saudável com informações antigas.

## Sinais complementares

| Sinal | Pergunta operacional |
| --- | --- |
| Status e conditions | Qual é o estado conhecido desta solicitação? |
| Logs estruturados | Qual decisão ocorreu e em que contexto? |
| Kubernetes Events | Qual acontecimento relevante foi associado ao recurso? |
| Métricas | Qual o estado atual e a tendência das operações? |

Gauges representam valores atuais; counters acumulam eventos observados pelo processo; histogramas representam distribuições de duração. Sinais com labels de alta cardinalidade, como hash de cada tentativa, podem tornar a coleta cara. Identidade por CR é útil para saúde individual, mas deve ter cardinalidade planejada. Logs e métricas não devem expor credenciais ou planos.

## Aplicação na POC

O TerraformStack publica JSON em stdout, Events Kubernetes e métricas nativas e por CR. Os gauges vêm do status no cache; os contadores observam transições persistidas e reiniciam com o processo. Uma tentativa registrada sem término não comprova que seu Pod está ativo. Um drift desconhecido não equivale a zero drift.

O controller não instala coletores, backends, dashboards ou alertas. O operador escolhe como consumir esses sinais; a integração central permanece no roadmap. A observabilidade não deve alterar as decisões de reconciliação. Para avaliar experiência do consumidor, combinar os sinais técnicos com medidas de jornada e feedback, conforme o artigo 15.

## Leitura complementar

- [Operação do TerraformStack](../../control-plane/terraformstack/controller/README.md#observabilidade-operacional)
- [Observability Plane](../../observability-plane/README.md)
- [Prometheus — tipos de métricas](https://prometheus.io/docs/concepts/metric_types/)

---

[Aula anterior](31-identidade-e-fronteiras-de-confianca.md) · [Índice](README.md) · [Próxima aula](33-testes-e-evolucao-segura.md)
