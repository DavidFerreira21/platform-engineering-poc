# Aula 20 — Multi-tenancy, isolamento e células

**Trilha:** Arquitetura conceitual · **Tipo:** fundamento conceitual · **Pré-requisito:** aula 19

## Objetivos de aprendizagem

Ao terminar, você deve conseguir explicar o conceito com suas palavras, reconhecer o problema que ele resolve e relacioná-lo às decisões da POC.

## Conceito e por que existe

Multi-tenancy define como diferentes consumidores compartilham recursos e controle. Isolamento lógico pode usar namespaces e RBAC; isolamento mais forte pode exigir contas e clusters separados. Cells são unidades com limites próprios para conter falhas e reduzir blast radius. A escolha considera risco, custo, autonomia e operação; namespace não equivale a fronteira de segurança completa.

## Como aparece na nossa POC

UAT e PROD em contas AWS separadas; EKS PCI e Non-PCI por ambiente. A POC não precisa alegar multi-tenant enterprise só porque há múltiplos Domains.

## Erro comum

Não confundir o conceito com a simples instalação de uma ferramenta. A evidência de uma capacidade de plataforma está na experiência de consumo, no contrato e na possibilidade de observar e corrigir o resultado.

## Exercício conceitual

Por que compartilhar o mesmo Argo CD entre PCI e Non-PCI exige análise de privilégios?

**Critério de conclusão:** responder com um exemplo, um trade-off e uma forma de verificar se o conceito foi aplicado.

## Para fixar

Entender isolamento por namespace, cluster, conta e escopo administrativo.

## Referência e leitura complementar

Inspirado no tema **Capítulo 3 — multi-tenancy e cell models.** do livro *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026), com explicações e exemplos **autorais**, adaptados à nossa POC. Consulte o exemplar licenciado para aprofundamento; este texto não o substitui nem reproduz trechos.

---

[Aula anterior](./19-abstracoes-e-escape-hatches.md) · [Índice da trilha](README.md) · [Próxima aula](./21-api-first-operadores-adrs.md)
