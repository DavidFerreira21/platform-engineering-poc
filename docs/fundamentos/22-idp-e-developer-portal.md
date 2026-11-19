# Internal Developer Platform (IDP) não é a mesma coisa que Developer Portal

Uma **Internal Developer Platform (IDP)** é o conjunto integrado de capacidades, contratos, automações e experiências que permite às equipes desenvolver, disponibilizar e operar software com autonomia. O valor não está numa interface específica, mas na capacidade entregue: por exemplo, solicitar um banco, obter permissões, implementar uma aplicação e compreender seu estado sem depender de coordenação manual para cada detalhe.

Um **Developer Portal** é uma das possíveis interfaces dessa plataforma. Ele organiza descoberta, documentação, catálogo, templates e ações. O Backstage é um exemplo conhecido. Mas instalar um portal não equivale a construir uma IDP: um catálogo com botões que apenas abre tickets continua dependente de trabalho manual. Por outro lado, uma plataforma com APIs confiáveis, contratos declarativos, documentação e automações pode ter self-service antes de possuir qualquer portal.

## Contrato e canais de consumo

O desenho da POC separa a **Experience Layer** (Backstage, CLI, GitOps, AI/MCP) da API oficial: o **Kubernetes API Server**. Um consumidor pode iniciar uma solicitação via PR ou cliente autorizado, e a plataforma usa controllers e reconciliação para materializar a intenção. Assim, a experiência pode evoluir sem redefinir todos os contratos ou criar APIs paralelas para cada canal.

Essa distinção também reduz acoplamento: um catálogo que aponta para um `Database` pode mudar sua apresentação enquanto o recurso declarado continua versionado e estável. O portal deve melhorar descoberta e feedback; não se tornar a única forma de consumo.

## Na nossa POC

A primeira Capability de ObjectStorage pode ser entregue com manifest Git, XRD Crossplane, reconciliação e status, sem esperar Backstage. Quando o portal for incluído, ele consumirá os contratos existentes e oferecerá um percurso mais simples, e não um novo motor de provisionamento.

## Leitura complementar

- [CNCF — Platforms White Paper](https://tag-app-delivery.cncf.io/whitepapers/platforms/)
- [Backstage — documentação oficial](https://backstage.io/docs/overview/what-is-backstage/)

---

[Aula anterior](21-api-first-operadores-adrs.md) · [Índice](README.md) · [Próxima aula](23-catalogo-e-capacidades.md)
