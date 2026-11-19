# Aula 14 — Thinnest Viable Platform (TVP) e priorização

**Trilha:** Organização e descoberta · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 13

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

TVP é o menor conjunto de capacidades que reduz uma dor relevante e pode evoluir com feedback. Não é a arquitetura completa em miniatura, nem um MVP vazio. Priorize dor, frequência, impacto, risco e esforço. Explicite o que está fora do escopo e só amplie após provar o fluxo ponta a ponta. O objetivo é um produto utilizável, mesmo estreito.

## Como aparece na nossa POC

Nossa primeira fatia entrega ObjectStorage real numa conta AWS preexistente, com XRD, Composition, status e cleanup. Não exige Backstage nem sete clusters EKS.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Qual é a menor demonstração que provaria valor do nosso Control Plane antes do Domain e do Environment completos?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Evitar implantação de ferramentas sem comprovação de capacidade.

## Referência e leitura complementar

Inspirado no tema **Capítulo 2 — thinnest viable platform.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./13-descoberta-e-entrevistas.md) · [Índice da trilha](README.md) · [Próxima aula](./15-metricas-dx.md)
