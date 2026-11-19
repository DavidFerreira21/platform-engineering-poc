# Aula 19 — Abstrações, contratos e escape hatches

**Trilha:** Arquitetura conceitual · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 18

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Uma boa abstração permite ao consumidor expressar intenção sem expor parâmetros operacionais de baixo nível. Uma abstração rasa exige conhecimento demais; uma abstração rígida bloqueia casos legítimos. Escolha o nível que preserve decisões relevantes (classe, região, retenção, disponibilidade) e esconda detalhes de implementação (roles, recursos intermediários, módulos). Escape hatch precisa de governança para exceções.

## Como aparece na nossa POC

Database oferece perfil small/medium e modo de disponibilidade, mas não exige definição de cada parâmetro Terraform. Um caso não previsto pode solicitar evolução do contrato.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Quais três campos devem estar no spec público de ObjectStorage e quais três não deveriam estar?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Fazer APIs pequenas, estáveis e úteis sem perder extensibilidade.

## Referência e leitura complementar

Inspirado no tema **Capítulo 3 — developer-facing abstractions e escape hatches.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./18-estado-declarativo.md) · [Índice da trilha](README.md) · [Próxima aula](./20-multitenancy-cells.md)
