# Multi-tenancy, isolamento e células

Multi-tenancy define como diferentes consumidores compartilham recursos e controle. Isolamento lógico pode usar namespaces e RBAC; isolamento mais forte pode exigir contas e clusters separados. Cells são unidades com limites próprios para conter falhas e reduzir blast radius. A escolha considera risco, custo, autonomia e operação; namespace não equivale a fronteira de segurança completa.

## Existem vários limites de isolamento

Namespace é um mecanismo de organização e aplicação de controles Kubernetes, mas não equivale ao isolamento físico de uma conta AWS. Um cluster compartilhado tem um blast radius operacional e uma superfície de administração; separar clusters e contas modifica ambos, com custos e complexidade maiores. Cells são unidades de isolamento/entrega capazes de limitar impacto de falhas. A topologia deve derivar de requisitos, não de nomenclatura.

## PCI e gerenciamento remoto

UAT e PROD em contas AWS separadas; EKS PCI e Non-PCI por ambiente. A POC não precisa alegar multi-tenant enterprise só porque há múltiplos Domains.

Na POC, UAT e PROD são contas distintas, e cada uma poderá ter EKS PCI e EKS Non-PCI. Um Argo CD do Management EKS com credenciais administrativas para ambos cruza os dois escopos; a conformidade e os controles de segurança precisam ser examinados antes da implantação.

## Leitura complementar

- [CNCF — Platform Engineering Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 3 — multi-tenancy e cell models.

---

[Aula anterior](19-abstracoes-e-escape-hatches.md) · [Índice](README.md) · [Próxima aula](21-api-first-operadores-adrs.md)
