# Build, Buy ou Adopt

Ao escolher uma capacidade, avalie construção própria, contratação de produto ou adoção/configuração de open source. Compare tempo para valor, custo total, equipe disponível, integração, extensibilidade, lock-in e manutenção. O problema antecede a ferramenta: construir um controller de Database só faz sentido se Crossplane não cumprir contratos necessários. A mesma escolha pode variar por produto.

## Decidir com custo total, não com preferência

Construir internamente oferece controle sobre a interface, mas traz manutenção, documentação, testes e atualização contínua. Comprar pode acelerar a entrega e transferir parte da operação, mas cria dependência contratual. Adotar software aberto acelera a experimentação, mas a equipe ainda responde por integração e operação. Compare o problema, requisitos de segurança, competências disponíveis, integração, portabilidade e custo de mudança.

## A escolha da nossa POC

Na POC, adotamos Crossplane para Capabilities. Entity, Environment e Domain são inicialmente registros lógicos no Git; não construiremos um controller próprio antes de identificar necessidade real. Backstage é alternativa de experiência, não a API canônica.

Usamos a API Kubernetes e Crossplane para Capabilities porque oferecem contratos e reconciliação sem exigir um controller customizado por serviço. Para a criação futura de contas AWS, avaliaremos Argo Workflows e, somente se necessário, um contrato e controller de lifecycle dedicado.

## Leitura complementar

- [CNCF — Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/) — material de referência para aprofundamento.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — build, buy or adopt.

---

[Aula anterior](08-guardrails-e-gates.md) · [Índice](README.md) · [Próxima aula](10-team-topologies.md)
