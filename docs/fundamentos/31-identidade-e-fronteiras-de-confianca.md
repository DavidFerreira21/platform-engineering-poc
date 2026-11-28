# Identidade, autorização e fronteiras de confiança

Uma plataforma opera em nome de consumidores que não devem possuir necessariamente as mesmas permissões do executor. Identificar quem solicita, quem aprova e quem altera um recurso é essencial para prevenir escalonamento de privilégios. Autenticação comprova identidade; autorização determina o que essa identidade pode fazer.

## Controles em camadas

| Fronteira | Exemplo de controle |
| --- | --- |
| Git | Revisão por owner, proteção de branch e regras por escopo |
| API Kubernetes | RBAC, validação de schema e admissão |
| Workflow | Template e identidade aprovados, autorização de gates e integridade da execução |
| AWS | Role temporária, trust policy, permissões mínimas e validação do destino |

Esses controles se complementam. Uma ServiceAccount autorizada a atualizar um Workflow não deveria poder adulterar o resultado ou alterar livremente o template. Uma conta AWS registrada em Git não concede acesso à conta. Uma allowlist de integrações também não autoriza todos os solicitantes a escolher qualquer alias.

## Aplicação na POC

O Simulator protege a criação do Workflow de referência e limita o executor a registrar o claim vigente. O Platform Team controla templates, RBAC e admissão. A aprovação do PR, o gate de negócio e o handshake operacional são decisões separadas. Desabilitar o handshake requer uma integração compatível; não elimina as demais fronteiras.

A arquitetura AWS prevê identidade temporária de Pod no Hub e AssumeRole para o destino autorizado. Na AWS, serão verificados conta, região, permissões, isolamento e custo antes de executar alterações.

## Leitura complementar

- [ADR-0004 — Segurança AWS](../adr/0004-aws-security-and-roadmap.md)
- [Identidade do executor](../../foundation/iam/executor/README.md)
- [Kubernetes — RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)

---

[Aula anterior](30-configuracao-state-e-binding.md) · [Índice](README.md) · [Próxima aula](32-observabilidade-de-recursos.md)
