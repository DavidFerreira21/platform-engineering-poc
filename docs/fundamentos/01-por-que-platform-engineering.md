# Aula 01 — Por que Platform Engineering surgiu?

**Trilha:** História e problema · **Tipo:** fundamento conceitual · **Pré-requisito:** nenhum

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

A especialização crescente de infraestrutura fez equipes de produto dependerem de múltiplos especialistas para publicar mudanças. A passagem de operações centralizadas para DevOps ampliou o ownership das equipes, mas não tornou simples a compreensão simultânea de rede, IAM, Kubernetes, CI/CD, observabilidade e políticas. Platform Engineering surge para oferecer capacidades internas repetíveis, com contratos e suporte ao fluxo de trabalho, sem retirar das equipes a responsabilidade sobre suas aplicações.

## Como aparece na nossa POC

Na POC, o consumidor não precisa conhecer o módulo Terraform ou cada Provider AWS para pedir ObjectStorage. A plataforma encapsula essa complexidade em um contrato Kubernetes.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Quando o time precisar pedir uma fila, qual trabalho deveria continuar sendo dele e qual deveria ser absorvido pela plataforma?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Reduzir dependências operacionais e trabalho repetitivo, sem transformar a plataforma em novo gargalo.

## Referência e leitura complementar

Inspirado no tema **Capítulo 1 — evolução de DevOps para plataformas.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Índice](README.md) · [Índice da trilha](README.md) · [Próxima aula](./02-devops-e-platform-engineering.md)
