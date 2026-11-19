# Plataforma como serviço e self-service

A plataforma deve ser consumida como serviço: capacidades acessíveis por uma interface estável, preferencialmente sem tickets para tarefas aprovadas por política. Self-service não é ausência de governança; é automatizar verificações previsíveis e reservar aprovação humana para exceções ou decisões com risco real. O consumidor pede o resultado, acompanha o status e recebe erros acionáveis.

## O que significa consumir como serviço

Consumir uma capacidade como serviço significa pedir um resultado por contrato estável, receber feedback de estado, ter limites claros de suporte e poder operar dentro de permissões definidas. Portal, CLI e API são interfaces possíveis; não são três plataformas independentes. Um serviço interno pode oferecer opções controladas, como classes de banco, regiões permitidas ou retenção, sem expor as configurações internas do Provider AWS.

## Self-service com governança

O usuário aplica um CR ObjectStorage via GitOps e consulta Ready/NotReady, sem operar diretamente o Provider AWS. Para ações permanentes, Backstage/CLI/AI-MCP podem abrir um PR.

A maior parte das solicitações conformes deveria ser atendida automaticamente por validação de schema, identidade e políticas. Para ações excepcionalmente caras ou sensíveis, um gate autorizado é apropriado. Self-service não significa ausência de regras; significa que regras previsíveis foram incorporadas ao caminho do consumidor.

## Leitura complementar

- [CNCF — Platforms White Paper](https://tag-app-delivery.cncf.io/whitepapers/platforms/) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulos 1 e 3 — produto interno e contratos declarativos; ampliação autoral.. Referência temática; esta explicação é autoral, não uma tradução do livro.

---

[Índice](README.md) · [Aula anterior](05-plataforma-como-produto.md) · [Próxima aula](07-golden-paths.md)
