# Aula 07 — Golden Paths e Paved Roads

**Trilha:** Produto e experiência · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 06

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Golden Path é o percurso recomendado e suportado para realizar uma tarefa recorrente, com escolhas seguras e documentação. Deve ser mais simples que reinventar a integração. Não significa caminho único e obrigatório: exceções justificadas precisam de escape hatches. A maturidade vem de reduzir decisões acidentais e oferecer resultados previsíveis, como criar um domínio, solicitar um banco ou publicar um serviço.

## Como aparece na nossa POC

Golden Path de storage: selecionar ambiente e Domain, escolher perfil aprovado, gerar CR, abrir PR, observar reconciliação e obter endpoint/configuração permitida.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Quando um Golden Path vira burocracia em vez de ajudar? Que escape hatch seria aceitável?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Diferenciar recomendação suportada de imposição inflexível.

## Referência e leitura complementar

Inspirado no tema **Capítulo 1 — golden paths e paved roads.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./06-plataforma-como-servico.md) · [Índice da trilha](README.md) · [Próxima aula](./08-guardrails-e-gates.md)
