# Ciclo de vida de produtos e serviços de plataforma

Uma capacidade da plataforma não termina quando seu primeiro recurso é criado com sucesso. Como qualquer produto, ela possui **descoberta, desenho de contrato, lançamento, operação, evolução e eventualmente depreciação**. Cada etapa precisa considerar consumidores existentes e novas necessidades. A ausência de gestão do ciclo de vida faz uma plataforma acumular versões não suportadas, integrações frágeis e procedimentos que só um especialista conhece.

O produto precisa declarar quem mantém sua API e suas automações, quais mudanças são compatíveis e como os usuários serão informados. Em CRDs e XRDs, alterar schema, valor default ou política de exclusão pode transformar uma atualização aparentemente pequena num evento de migração. Observabilidade, documentação, suporte e resposta a incidentes são parte da entrega, não extras opcionais.

## Evolução controlada

Versionar o contrato possibilita evoluir interfaces com previsibilidade. Uma atualização de Composition pode mudar como recursos são provisionados sem alterar o formato público; isso exige testes que comprovem manutenção do comportamento esperado. Quando uma mudança incompatível for necessária, a plataforma deve definir migração, prazo e estratégia de descontinuação. Mudanças de política de segurança ou custos podem exigir reavaliação de perfis oferecidos.

Adoção e satisfação são sinais importantes, mas métricas de operação também importam: tempo até Ready, falhas de reconciliação, número de solicitações manuais, incidentes e frequência de exceções mostram se o produto permanece sustentável.

## Na nossa POC

A Capability ObjectStorage começa com um perfil simples e uma conta AWS. À medida que surgirem UAT e PROD, políticas adicionais e novas opções, registraremos decisões em ADRs, atualizaremos schemas com cuidado e manteremos um percurso de migração. Não devemos apagar recursos AWS reais apenas porque um exemplo Git foi removido — o comportamento de exclusão precisa ser deliberado.

## Leitura complementar

- [DORA — Platform Engineering](https://dora.dev/capabilities/platform-engineering/)
- [CNCF — Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/)

---

[Aula anterior](23-catalogo-e-capacidades.md) · [Índice](README.md) · [Roteiro prático](../learning-roadmap.md)
