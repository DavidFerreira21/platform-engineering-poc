# Fundamentos de Platform Engineering

Textos de referência conceitual, com explicações aprofundadas em português, exemplos da POC e fontes externas relevantes. Não é uma apostila de objetivos/exercícios.

**Comece por aqui antes dos labs.** O conteúdo é inteiramente conceitual e não exige AWS, Kubernetes ou instalações. A proposta é compreender **por que** construímos uma plataforma, **para quem**, **qual produto** oferecemos e **como** nossos contratos/planos operam.

## Referência bibliográfica e autoria

A sequência temática é inspirada principalmente nos **capítulos 1–3** de *Platform Engineering in Practice* — **Zebb Wright**, GitforGits (2026): motivos para criar plataformas; pensar como equipe de plataforma; arquitetura de IDP. A ordem, explicações, exemplos e decisões são **autorais e próprios da POC**, e não uma tradução, edição ou reprodução do livro. O estudo aprofunda o exemplar original adquirido separadamente.

### História e problema

Por que a disciplina existe e quais dores ela resolve.

- [Aula 01 — Por que Platform Engineering surgiu?](01-por-que-platform-engineering.md)
- [Aula 02 — DevOps não é a mesma coisa que Platform Engineering](02-devops-e-platform-engineering.md)
- [Aula 03 — Carga cognitiva e troca de contexto](03-carga-cognitiva.md)
- [Aula 04 — Sintomas de uma plataforma ausente ou fraca](04-friccao-e-sintomas.md)

### Produto e experiência

Como transformar infraestrutura recorrente em um produto consumível.

- [Aula 05 — Plataforma como produto interno](05-plataforma-como-produto.md)
- [Aula 06 — Plataforma como serviço e self-service](06-plataforma-como-servico.md)
- [Aula 07 — Golden Paths e Paved Roads](07-golden-paths.md)
- [Aula 08 — Guardrails, políticas e aprovações](08-guardrails-e-gates.md)
- [Aula 09 — Build, Buy ou Adopt](09-construir-comprar-adotar.md)

### Organização e descoberta

Como entender usuários, organizar responsabilidades e medir resultados.

- [Aula 10 — Team Topologies e tipos de equipes](10-team-topologies.md)
- [Aula 11 — Modos de interação entre equipes](11-modos-de-interacao.md)
- [Aula 12 — Personas e jornada do desenvolvedor](12-personas-e-jornadas.md)
- [Aula 13 — Discovery: entrevistas e identificação de fricções](13-descoberta-e-entrevistas.md)
- [Aula 14 — Thinnest Viable Platform (TVP) e priorização](14-tvp.md)
- [Aula 15 — Developer Experience, DORA e SPACE](15-metricas-dx.md)
- [Aula 16 — Platform Charter, escopo e backlog](16-charter-e-backlog.md)

### Arquitetura conceitual

Como contratos, planos, reconciliação e isolamento funcionam.

- [Aula 17 — Camadas e planos de uma plataforma](17-planos-e-responsabilidades.md)
- [Aula 18 — Estado desejado, estado observado e reconciliação](18-estado-declarativo.md)
- [Aula 19 — Abstrações, contratos e escape hatches](19-abstracoes-e-escape-hatches.md)
- [Aula 20 — Multi-tenancy, isolamento e células](20-multitenancy-cells.md)
- [Aula 21 — API-first, CRDs, XRDs, Operators e ADRs](21-api-first-operadores-adrs.md)
- [Internal Developer Platform versus Developer Portal](22-idp-e-developer-portal.md)
- [Catálogo de serviços, produtos e capacidades](23-catalogo-e-capacidades.md)
- [Ciclo de vida de produtos e serviços de plataforma](24-ciclo-de-vida-produtos.md)

## Como estudar

Leia os artigos na ordem sugerida, reflita sobre os exemplos e consulte o [mapa da arquitetura](../architecture.md) para identificar onde cada ideia se manifesta. Ao concluir os 24 artigos, avance para o [roteiro de implementação](../learning-roadmap.md). Instalar ferramentas sem compreender o problema não é critério de conclusão.

## Padrão editorial

Os artigos são textos explicativos autorais, sem seções fixas de objetivos, exercícios ou resumos. Devem desenvolver o conceito com profundidade, explicar implicações e limites, trazer exemplos pertinentes da POC e links confiáveis de aprofundamento, preservando a navegação entre textos. Os labs, em contraste, exigem comandos, evidências, falhas e recuperação.

## O que esta trilha não pretende

Não reconta página por página o livro, não reproduz os diagramas/códigos do autor e não toma os produtos do livro como obrigações técnicas. Na prática utilizaremos nossos contratos Kubernetes-native, Crossplane, Argo Workflows, AWS real e a topologia UAT/PROD da POC.
