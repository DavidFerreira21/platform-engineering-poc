# Abstrações, contratos e escape hatches

Uma boa abstração permite ao consumidor expressar intenção sem expor parâmetros operacionais de baixo nível. Uma abstração rasa exige conhecimento demais; uma abstração rígida bloqueia casos legítimos. Escolha o nível que preserve decisões relevantes (classe, região, retenção, disponibilidade) e esconda detalhes de implementação (roles, recursos intermediários, módulos). Escape hatch precisa de governança para exceções.

## Contratos devem representar intenções duráveis

Uma abstração útil inclui decisões relevantes para o usuário e deixa de fora detalhes de infraestrutura que só existem por causa da implementação atual. O design precisa considerar compatibilidade entre versões: remover um campo, mudar valor default ou alterar comportamento de exclusão pode quebrar consumidores. Os escape hatches atendem cenários legítimos sem transformar toda exceção em parâmetro de baixo nível aberto para qualquer pessoa.

## Delimitação do produto

Database oferece perfil small/medium e modo de disponibilidade, mas não exige definição de cada parâmetro Terraform. Um caso não previsto pode solicitar evolução do contrato.

Para Database, campos como classe de serviço, ambiente, retenção e disponibilidade fazem sentido. A lista exata de subnets, provider-specific IDs e scripts Terraform normalmente não deveria aparecer no contrato público. A escolha final depende do que os consumidores realmente precisam.

## Leitura complementar

- [CNCF — Platforms White Paper](https://tag-app-delivery.cncf.io/whitepapers/platforms/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 3 — developer-facing abstractions e escape hatches.

---

[Aula anterior](18-estado-declarativo.md) · [Índice](README.md) · [Próxima aula](20-multitenancy-cells.md)
