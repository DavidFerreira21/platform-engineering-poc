# Por que Platform Engineering surgiu?

A especialização crescente de infraestrutura fez equipes de produto dependerem de múltiplos especialistas para publicar mudanças. A passagem de operações centralizadas para DevOps ampliou o ownership das equipes, mas não tornou simples a compreensão simultânea de rede, IAM, Kubernetes, CI/CD, observabilidade e políticas. Platform Engineering surge para oferecer capacidades internas repetíveis, com contratos e suporte ao fluxo de trabalho, sem retirar das equipes a responsabilidade sobre suas aplicações.

## Da infraestrutura compartilhada à plataforma como produto

Uma equipe central pode automatizar tarefas sem oferecer uma plataforma. A diferença aparece quando seus consumidores encontram capacidades padronizadas, documentação confiável, contratos versionados e mecanismos para acompanhar a própria solicitação. A evolução não é uma linha obrigatória Ops → DevOps → Plataforma: as organizações combinam práticas conforme suas necessidades. O problema central é que conhecimento de infraestrutura cresce mais rápido do que a capacidade de cada equipe aprender tudo.

## Onde a POC se encaixa

No laboratório, o resultado desejado é permitir que um consumidor solicite storage ou banco e acompanhe a reconciliação sem aprender como um Provider AWS é configurado. Kubernetes, Crossplane e Terraform são meios para oferecer essa experiência, não a definição de Platform Engineering.

## Leitura complementar

- [CNCF — O que é Platform Engineering](https://www.cncf.io/blog/2025/11/19/what-is-platform-engineering/) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — evolução de DevOps para plataformas. Referência temática; esta explicação é autoral, não uma tradução do livro.

---

[Índice](README.md) · [Próxima aula](02-devops-e-platform-engineering.md)
