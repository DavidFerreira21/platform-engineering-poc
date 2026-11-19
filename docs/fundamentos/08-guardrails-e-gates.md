# Aula 08 — Guardrails, políticas e aprovações

**Trilha:** Produto e experiência · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 07

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Guardrail é controle integrado ao fluxo que previne ou detecta violações de maneira previsível: schema, RBAC, quotas, políticas e testes. Gate é uma decisão ou pausa explícita; pode ser importante para operações caras, sensíveis ou dependentes de equipes externas, mas não deve ser aplicado indiscriminadamente. Bons guardrails permitem avançar de forma segura na maior parte dos casos.

## Como aparece na nossa POC

Um XRD Database só aceita classes aprovadas e exige tags. Já a criação de conta por AWSEnvironment pode precisar de um gate humano auditado no Argo Workflows.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Separe três situações resolvidas por políticas automáticas e duas que justificam aprovação humana.

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Governança por padrão e aprovação por exceção.

## Referência e leitura complementar

Inspirado no tema **Capítulo 1 — guardrails versus gates.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./07-golden-paths.md) · [Índice da trilha](README.md) · [Próxima aula](./09-construir-comprar-adotar.md)
