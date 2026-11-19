# Aula 06 — Plataforma como serviço e self-service

**Trilha:** Produto e experiência · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 05

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

A plataforma deve ser consumida como serviço: capacidades acessíveis por uma interface estável, preferencialmente sem tickets para tarefas aprovadas por política. Self-service não é ausência de governança; é automatizar verificações previsíveis e reservar aprovação humana para exceções ou decisões com risco real. O consumidor pede o resultado, acompanha o status e recebe erros acionáveis.

## Como aparece na nossa POC

O usuário aplica um CR ObjectStorage via GitOps e consulta Ready/NotReady, sem operar diretamente o Provider AWS. Para ações permanentes, Backstage/CLI/AI-MCP podem abrir um PR.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Como dar autonomia para criar storage sem permitir buckets públicos por acidente?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Entender contrato, automação e limites do autosserviço.

## Referência e leitura complementar

Inspirado no tema **Capítulos 1 e 3 — produto interno e contratos declarativos; ampliação autoral.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./05-plataforma-como-produto.md) · [Índice da trilha](README.md) · [Próxima aula](./07-golden-paths.md)
