# Developer Experience, DORA e SPACE

Meça resultado da plataforma, não apenas disponibilidade de ferramentas. DORA ajuda a observar fluxo de entrega e estabilidade; SPACE considera satisfação, desempenho, atividade, comunicação/colaboração e eficiência/fluxo. Para uma plataforma, complemente com tempo para provisionar, taxa de sucesso self-service, incidentes, adoção voluntária e esforço de suporte. Métricas isoladas geram incentivos ruins; combine quantitativo e qualitativo.

## Indicadores que mostram experiência

Um painel da plataforma pode medir latência de reconciliação e disponibilidade, mas esses sinais não substituem o que o consumidor percebe. DORA ajuda a enxergar fluxo de mudanças e estabilidade; SPACE amplia a análise com satisfação, desempenho, atividade, colaboração e eficiência. Esses modelos não devem virar ranking individual de engenheiros. Use medidas de tendências por fluxo e combine dados quantitativos com relatos de dificuldades.

## Métricas de produto na POC

O TerraformStack já publica duração de tentativas, condições e contadores de falhas. Isso mede comportamento operacional, não Developer Experience por si só. Uma medida de jornada pode incluir tempo entre PR, aprovação e Ready, número de intervenções e feedback do consumidor. No laboratório, não atribuir melhoria de DX a métricas sem baseline ou pesquisa com usuários.

Para ObjectStorage, registre o tempo entre submissão do contrato e Ready, quantas falhas são recuperáveis pelo usuário e quantas exigem o operador. A velocidade só é um ganho se controles e observabilidade permanecerem adequados.

## Leitura complementar

- [DORA — Platform Engineering](https://dora.dev/capabilities/platform-engineering/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 2 — DORA, SPACE e experiência do desenvolvedor.

---

[Aula anterior](14-tvp.md) · [Índice](README.md) · [Próxima aula](16-charter-e-backlog.md)
