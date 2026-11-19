# Platform Charter, escopo e backlog

O Platform Charter descreve missão, público, problemas atendidos, limites, princípios, responsabilidades e métricas. O backlog conecta dores às capacidades propostas; não é uma lista de ferramentas para instalar. Defina como tratar novas solicitações, exceções e mudanças de contrato para evitar que o produto perca coerência.

## Um charter transforma intenção em responsabilidade

O charter formaliza para quem a plataforma existe, o que promete entregar, o que fica de fora, que decisões são suas e como será acompanhada. O backlog liga necessidades comprovadas a capacidades, com justificativa e forma de validação. Sem isso, o roadmap vira uma sequência de instalações motivadas pela curiosidade com ferramentas.

## Escopo da nossa POC

Na POC, o charter pode afirmar: entregar AWS self-service via Kubernetes API; início em storage; exclusão de criação automática de conta na fase 1; sucesso medido por reconciliação e recuperação.

No início, entregaremos uma Capability ObjectStorage em AWS real, a partir de uma API Kubernetes e com reconciliação observável. A automação de criação de contas e a implementação integral de UAT/PROD ficam para fases posteriores. Essas escolhas deixam a experiência ensinável e o progresso mensurável.

## Leitura complementar

- [CNCF — Platform Engineering Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 2 — platform charter e backlog.. Referência temática; texto autoral adaptado à POC.

---

[Aula anterior](15-metricas-dx.md) · [Índice](README.md) · [Próxima aula](17-planos-e-responsabilidades.md)
