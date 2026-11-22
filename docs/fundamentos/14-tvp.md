# Thinnest Viable Platform (TVP) e priorização

TVP é o menor conjunto de capacidades que reduz uma dor relevante e pode evoluir com feedback. Não é a arquitetura completa em miniatura, nem um MVP vazio. Priorize dor, frequência, impacto, risco e esforço. Explicite o que está fora do escopo e só amplie após provar o fluxo ponta a ponta. O objetivo é um produto utilizável, mesmo estreito.

## Menor plataforma viável não é menor diagrama

Thinnest Viable Platform (TVP) é a menor oferta integrada que resolve um problema concreto de ponta a ponta. Um catálogo vazio, dez operators instalados ou um cluster disponível são peças de tecnologia; só viram um produto quando alguém consegue realizar uma tarefa importante com confiança. Priorizar uma fatia vertical estreita torna visíveis as lacunas de identidade, reconciliação, documentação e suporte.

## Nossa primeira fatia

Nossa primeira fatia entrega ObjectStorage real numa conta AWS preexistente, com XRD, Composition, status e cleanup. Não exige Backstage nem sete clusters EKS.

Um XRD ObjectStorage com Composition que provisiona um bucket real numa conta autenticada, mostra condições de status e permite limpeza segura já demonstra um produto incipiente. Não precisamos aguardar a implementação simultânea de Backstage, dois ambientes completos e seis EKS.

## Leitura complementar

- [CNCF — Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/) — material de referência para aprofundamento.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 2 — thinnest viable platform.

---

[Aula anterior](13-descoberta-e-entrevistas.md) · [Índice](README.md) · [Próxima aula](15-metricas-dx.md)
