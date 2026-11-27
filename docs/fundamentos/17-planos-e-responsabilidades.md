# Camadas e planos de uma plataforma

Arquiteturas de plataforma podem agrupar responsabilidades em experiências, controle, entrega/workflows, recursos, segurança e observabilidade. Não existe uma taxonomia única universal: são modelos para tornar limites claros, não regras rígidas. Na nossa POC usamos Experience Layer, Control Plane, Workflow Plane, Observability Plane e Data Plane. Planes lógicos podem coexistir num cluster; topologia física e responsabilidades não são sinônimos.

## Plano lógico não é cluster físico

Um plano organiza responsabilidade de arquitetura. Control Plane representa contratos e reconciliação; Workflow Plane representa processos, integração GitOps e executores; Observability Plane dá visibilidade transversal; Data Plane reúne contas, infraestrutura e workloads; Experience Layer oferece os canais de consumo. Sistemas distintos podem rodar no mesmo cluster, e um componente pode servir a mais de um fluxo sem perder uma responsabilidade principal.

## A posição do Crossplane

Platform Management Cluster hospeda Control/Workflow/Observability. UAT e PROD são Data Planes; Hub-and-Spoke explica gerenciamento, não necessariamente rede.

Crossplane fica no Control Plane da POC porque reconcilia XRs definidos por XRDs. O TerraformStack Controller também fica nesse plano, coordenando o lifecycle de suas solicitações por um contrato de execução. Argo Workflows fica no Workflow Plane porque coordena passos e gates. Ambos podem rodar fisicamente no Platform Management Cluster local, sem criar clusters separados só para respeitar o diagrama.

## Leitura complementar

- [CNCF — Platforms White Paper](https://tag-app-delivery.cncf.io/whitepapers/platforms/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 3 — arquitetura em planos.

---

[Aula anterior](16-charter-e-backlog.md) · [Índice](README.md) · [Próxima aula](18-estado-declarativo.md)
