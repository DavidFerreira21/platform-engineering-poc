# Aula 17 — Camadas e planos de uma plataforma

**Trilha:** Arquitetura conceitual · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 16

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Arquiteturas de plataforma podem agrupar responsabilidades em experiências, controle, entrega/workflows, recursos, segurança e observabilidade. Não existe uma taxonomia única universal: são modelos para tornar limites claros, não regras rígidas. Na nossa POC usamos Experience Layer, Control Plane, Workflow Plane, Observability Plane e Data Plane. Planes lógicos podem coexistir num cluster; topologia física e responsabilidades não são sinônimos.

## Como aparece na nossa POC

Platform Management Cluster hospeda Control/Workflow/Observability. UAT e PROD são Data Planes; Hub-and-Spoke explica gerenciamento, não necessariamente rede.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Se Crossplane e Argo Workflows rodam no mesmo cluster, por que pertencem a responsabilidades diferentes?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Ler arquitetura por contrato, reconciliação e execução, não por posição em um diagrama.

## Referência e leitura complementar

Inspirado no tema **Capítulo 3 — arquitetura em planos; classificação da POC é autoral.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./16-charter-e-backlog.md) · [Índice da trilha](README.md) · [Próxima aula](./18-estado-declarativo.md)
