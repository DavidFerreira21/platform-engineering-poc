# Aula 21 — API-first, CRDs, XRDs, Operators e ADRs

**Trilha:** Arquitetura conceitual · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 20

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

API-first significa definir o contrato do consumidor antes de selecionar ferramentas. No Kubernetes, CRDs estendem a API; XRDs do Crossplane definem recursos compostos e sua reconciliação; Compositions e Providers materializam serviços. Controllers/Operators observam CRs e executam lógica. ADR registra contexto, decisão e consequências para impedir que escolhas virem regras sem justificativa. Não confunda uma API pública com um controller de execução.

## Como aparece na nossa POC

AWSEnvironment é CRD + Environment Controller próprio; Domain e Capabilities são XRD + Crossplane. Todos são consumidos pelo mesmo Kubernetes API Server, mas executados por reconciliadores diferentes.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Um pedido de Database 'bate no Crossplane' diretamente ou na Kubernetes API? Quem atualiza status e por quê?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Unir os conceitos para ler os ADRs e iniciar os labs com clareza.

## Referência e leitura complementar

Inspirado no tema **Capítulo 3 — API-first, operator pattern e ADRs; aplicação adaptada à POC.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./20-multitenancy-cells.md) · [Índice da trilha](README.md) · [Seguir para implementação](../learning-roadmap.md)
