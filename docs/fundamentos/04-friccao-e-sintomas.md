# Aula 04 — Sintomas de uma plataforma ausente ou fraca

**Trilha:** História e problema · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 03

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Filas de tickets, ambientes especiais que ninguém reproduz, pipelines copiados, permissões solicitadas por mensagens e divergência entre ambientes são sintomas de fricção. Não é necessário construir um portal para começar: primeiro mensure tempos de espera, taxa de retrabalho, tarefas repetidas e incidentes ligados ao provisionamento. Procure a capacidade comum subjacente e valide se padronização seria útil.

## Como aparece na nossa POC

Uma solicitação de bucket que exige várias aprovações manuais não é resolvida apenas instalando Backstage. Precisamos de contrato, política de acesso, criação e status auditável.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Que evidência você coletaria para demonstrar que um fluxo de provisionamento merece virar produto de plataforma?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Identificar causas recorrentes antes de escolher ferramentas.

## Referência e leitura complementar

Inspirado no tema **Capítulo 1 — filas, snowflakes e proliferação de ferramentas.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./03-carga-cognitiva.md) · [Índice da trilha](README.md) · [Próxima aula](./05-plataforma-como-produto.md)
