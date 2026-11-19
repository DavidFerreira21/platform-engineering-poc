# Aula 09 — Build, Buy ou Adopt

**Trilha:** Produto e experiência · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 08

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Ao escolher uma capacidade, avalie construção própria, contratação de produto ou adoção/configuração de open source. Compare tempo para valor, custo total, equipe disponível, integração, extensibilidade, lock-in e manutenção. O problema antecede a ferramenta: construir um controller de Database só faz sentido se Crossplane não cumprir contratos necessários. A mesma escolha pode variar por produto.

## Como aparece na nossa POC

Na POC, adotamos Crossplane para Capabilities e construiremos um Environment Controller apenas para lifecycle complexo; Backstage é alternativa de experiência, não a API canônica.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Faça uma matriz simples com custo operacional e flexibilidade para controller próprio vs Crossplane.

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Avaliar ferramentas como estratégia de produto, não coleção de tecnologias.

## Referência e leitura complementar

Inspirado no tema **Capítulo 1 — build, buy or adopt.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./08-guardrails-e-gates.md) · [Índice da trilha](README.md) · [Próxima aula](./10-team-topologies.md)
