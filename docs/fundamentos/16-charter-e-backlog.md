# Platform Charter, escopo e backlog

O Platform Charter descreve missão, público, problemas atendidos, limites, princípios, responsabilidades e métricas. O backlog conecta dores às capacidades propostas; não é uma lista de ferramentas para instalar. Defina como tratar novas solicitações, exceções e mudanças de contrato para evitar que o produto perca coerência.

## Um charter transforma intenção em responsabilidade

O charter formaliza para quem a plataforma existe, o que promete entregar, o que fica de fora, que decisões são suas e como será acompanhada. O backlog liga necessidades comprovadas a capacidades, com justificativa e forma de validação. Sem isso, o roadmap vira uma sequência de instalações motivadas pela curiosidade com ferramentas.

## Escopo da nossa POC

Na POC, o charter pode afirmar: entregar AWS self-service via Kubernetes API; início em storage; exclusão de criação automática de conta na fase 1; sucesso medido por reconciliação e recuperação.

A sequência real começa pelo Hub local, TerraformStack e Crossplane local; depois segue para Foundation e Hub EKS, UAT, Capabilities AWS e aplicação. ObjectStorage é a primeira Capability AWS, não a primeira entrega técnica. A automação de criação de contas não é requisito inicial. A Experience Layer e o Golden Path integrado entram no fim da trilha, sem alterar as interfaces já testadas. [Roadmap](../learning-roadmap.md). Essas escolhas deixam a experiência ensinável e o progresso mensurável.

## Leitura complementar

- [CNCF — Platform Engineering Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 2 — platform charter e backlog.

---

[Aula anterior](15-metricas-dx.md) · [Índice](README.md) · [Próxima aula](17-planos-e-responsabilidades.md)
