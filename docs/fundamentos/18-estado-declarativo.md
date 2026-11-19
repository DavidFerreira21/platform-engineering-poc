# Aula 18 — Estado desejado, estado observado e reconciliação

**Trilha:** Arquitetura conceitual · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 17

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

O estado desejado descreve resultado, como um bucket com política ou um Domain com quota; o estado observado informa o que existe de fato. Um reconciliador compara os dois e tenta convergir de forma idempotente, atualizando condições e expondo erros. Git contém o spec revisado; Kubernetes API contém o recurso e o status. Drift é divergência relevante, não apenas diferença textual de YAML.

## Como aparece na nossa POC

Ao criar ObjectStorage, Crossplane observa o XR e seu recurso AWS até informar Ready. O Environment Controller observa AWSEnvironment e avança condições conforme workflows.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Qual a diferença entre 'o usuário pediu Database' e 'Database está Ready'?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Compreender por que controllers e status são essenciais ao self-service.

## Referência e leitura complementar

Inspirado no tema **Capítulo 3 — declared/actual state e loops de reconciliação.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./17-planos-e-responsabilidades.md) · [Índice da trilha](README.md) · [Próxima aula](./19-abstracoes-e-escape-hatches.md)
