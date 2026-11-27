# 📖 Fundamentos de Platform Engineering

Uma base de conhecimento em **português brasileiro** para compreender os princípios da disciplina antes de começar os laboratórios.

**34 artigos · 5 módulos · Conteúdo conceitual · Exemplos da POC**

> **Como usar:** siga a sequência sugerida ou consulte um tema diretamente. Os artigos explicam os conceitos e suas implicações; a [trilha prática](../learning-roadmap.md) mostra como aplicá-los em Kubernetes e AWS.

## 🗺️ Mapa de leitura

| Módulo | Tema | Escopo |
| :--- | :--- | :--- |
| 🧭 **Contexto e origem** | Artigos 01–04 | Por que Platform Engineering surgiu, DevOps, carga cognitiva e fricções. |
| 🛠️ **Produto e experiência** | Artigos 05–09 | Produto interno, self-service, Golden Paths, guardrails e Build/Buy/Adopt. |
| 👥 **Times e descoberta** | Artigos 10–16 | Team Topologies, personas, discovery, TVP, métricas e charter. |
| 🏗️ **Arquitetura e contratos** | Artigos 17–24 | Planes, reconciliação, isolamento, API-first, portal, catálogo e lifecycle. |
| ⚙️ **Padrões de implementação e operação** | Artigos 25–34 | GitOps, orquestração, contratos de execução, idempotência, drift, configuração, identidade, observabilidade, testes e integração por outputs. |

## 🧭 Contexto e origem

Por que Platform Engineering surgiu, DevOps, carga cognitiva e fricções.

- **01** · [Por que Platform Engineering surgiu?](01-por-que-platform-engineering.md)
- **02** · [DevOps e Platform Engineering](02-devops-e-platform-engineering.md)
- **03** · [Carga cognitiva e troca de contexto](03-carga-cognitiva.md)
- **04** · [Sintomas de uma plataforma ausente](04-friccao-e-sintomas.md)

## 🛠️ Produto e experiência

Produto interno, self-service, Golden Paths, guardrails e Build/Buy/Adopt.

- **05** · [Plataforma como produto interno](05-plataforma-como-produto.md)
- **06** · [Plataforma como serviço e self-service](06-plataforma-como-servico.md)
- **07** · [Golden Paths e Paved Roads](07-golden-paths.md)
- **08** · [Guardrails, políticas e aprovações](08-guardrails-e-gates.md)
- **09** · [Build, Buy ou Adopt](09-construir-comprar-adotar.md)

## 👥 Times e descoberta

Team Topologies, personas, discovery, TVP, métricas e charter.

- **10** · [Team Topologies](10-team-topologies.md)
- **11** · [Modos de interação](11-modos-de-interacao.md)
- **12** · [Personas e jornada do desenvolvedor](12-personas-e-jornadas.md)
- **13** · [Discovery e entrevistas](13-descoberta-e-entrevistas.md)
- **14** · [Thinnest Viable Platform (TVP)](14-tvp.md)
- **15** · [Developer Experience, DORA e SPACE](15-metricas-dx.md)
- **16** · [Platform Charter e backlog](16-charter-e-backlog.md)

## 🏗️ Arquitetura e contratos

Planes, reconciliação, isolamento, API-first, portal, catálogo e lifecycle.

- **17** · [Camadas e planos da plataforma](17-planos-e-responsabilidades.md)
- **18** · [Reconciliação e estado desejado](18-estado-declarativo.md)
- **19** · [Abstrações e escape hatches](19-abstracoes-e-escape-hatches.md)
- **20** · [Multi-tenancy, isolamento e cells](20-multitenancy-cells.md)
- **21** · [API-first, CRDs, XRDs, Operators e ADRs](21-api-first-operadores-adrs.md)
- **22** · [IDP versus Developer Portal](22-idp-e-developer-portal.md)
- **23** · [Catálogo de serviços e Capabilities](23-catalogo-e-capacidades.md)
- **24** · [Lifecycle de produtos de plataforma](24-ciclo-de-vida-produtos.md)

## ⚙️ Padrões de implementação e operação

Princípios de arquitetura e operação aplicados na construção da POC. Os artigos diferenciam o que já foi exercitado no Hub local das integrações AWS planejadas.

- **25** · [GitOps e fontes de verdade](25-gitops-e-fontes-de-verdade.md)
- **26** · [Reconciliação e orquestração](26-reconciliacao-e-orquestracao.md)
- **27** · [Contratos de execução e integrações desacopladas](27-contratos-e-integracoes.md)
- **28** · [Idempotência, concorrência e recuperação](28-idempotencia-e-concorrencia.md)
- **29** · [Drift, convergência e remediação](29-drift-e-convergencia.md)
- **30** · [Configuração, state e vínculo de destino](30-configuracao-state-e-binding.md)
- **31** · [Identidade, autorização e fronteiras de confiança](31-identidade-e-fronteiras-de-confianca.md)
- **32** · [Observabilidade operacional e saúde dos recursos](32-observabilidade-de-recursos.md)
- **33** · [Testes de contratos e evolução segura](33-testes-e-evolucao-segura.md)
- **34** · [Integração por outputs e GitOps Bridge](34-gitops-bridge-e-outputs.md)

---

## 📚 Sobre o conteúdo

Os temas iniciais utilizam como referência os capítulos 1–3 de *Platform Engineering in Practice*, de Zebb Wright (GitforGits, 2026). Os conceitos são apresentados e contextualizados para esta POC; os artigos não são traduções do livro. O quinto módulo amplia o material com padrões de engenharia observados nas decisões, no código e nos laboratórios da POC, sem atribuí-los ao livro.

O formato prioriza explicações bem desenvolvidas, comparações quando úteis, exemplos concretos e referências externas. Os artigos não incluem exercícios, objetivos formais ou resumos repetitivos.

**Continue:** [Arquitetura da POC](../architecture.md) · [Roteiro de implementação](../learning-roadmap.md) · [Página inicial](../../README.md)
