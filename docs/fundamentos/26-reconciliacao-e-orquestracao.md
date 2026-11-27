# Reconciliação e orquestração

Reconciliação é a responsabilidade de acompanhar a intenção e buscar convergência ao longo do tempo. Orquestração é coordenar passos de uma execução finita, incluindo dependências internas, pausas, verificações e resultados. São mecanismos complementares, não sinônimos.

## Quem decide e quem executa

Um reconciliador pode receber eventos repetidos, reiniciar, esperar uma dependência e voltar a avaliar o recurso. Um workflow normalmente representa uma tentativa com começo e fim. A execução de um workflow não torna automaticamente seu motor um reconciliador de recursos externos. Um scheduler recorrente também não basta para garantir convergência: é preciso conhecer intenção, resultado e regras de reexecução.

Separar essas responsabilidades permite mudar o processo interno sem reimplementar o lifecycle. Em contrapartida, surge uma fronteira distribuída: identidade da tentativa, estado persistido, resultado terminal e recuperação precisam de um contrato claro.

## Aplicação na POC

Crossplane reconcilia XRs e recursos compostos. O TerraformStack Controller acompanha cada CR, dependências, suspensão, retry e verificações periódicas; o Argo Workflows executa a pipeline selecionada. Na referência, os passos resolvem fonte, produzem plano, avaliam OPA, aguardam aprovação quando exigida e aplicam/verificam.

O controller não interpreta os steps do Argo para inferir sucesso ou início de apply. Ele usa sinais tipados do contrato. Um Workflow nativamente Succeeded pode representar bloqueio funcional, por exemplo quando a política nega a mudança.

## Leitura complementar

- [Kubernetes — Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [ADR-0009 — Runtime e contrato](../adr/0009-terraformstack-runtime.md)

---

[Aula anterior](25-gitops-e-fontes-de-verdade.md) · [Índice](README.md) · [Próxima aula](27-contratos-e-integracoes.md)
