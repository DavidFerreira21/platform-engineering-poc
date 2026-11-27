# Testes de contratos e evolução segura

Uma plataforma é uma dependência compartilhada. Alterar um default, uma política de exclusão ou um significado de status pode afetar muitos consumidores. Testes precisam verificar comportamento e garantias de segurança, não apenas sintaxe ou compilação.

## Camadas de validação

| Validação | Finalidade |
| --- | --- |
| Estática | Schema, manifests, configuração e documentação |
| Unitária | Regras, transições, dependências e resultados |
| Integração | Comportamento com APIs e componentes reais |
| Ponta a ponta | Jornada completa e efeitos no recurso gerenciado |

Testar caminhos de erro é tão importante quanto o caminho de sucesso. Exemplos: resposta incompatível, dependência bloqueada, reinício durante espera, retry, concorrência e deleção. CI aprovada não comprova implantação ou sucesso em uma conta AWS; o ambiente e a revisão das evidências devem ser identificados.

## Aplicação na POC

A suíte Go do TerraformStack usa fake clients, relógio determinístico e envtest com API Kubernetes e etcd. O runner e o Policy Engine possuem testes específicos. O Simulator exercita o fluxo real com Argo, IaC, OPA e ConfigMap no Hub local. A identidade e os efeitos em AWS precisam de aceites separados.

Upgrades compatíveis preservam state e contratos: drenar execuções, sincronizar CRD, configuração e controller na mesma revisão e conferir o rollout. Não apagar PVCs indiscriminadamente nem forçar finalizers para resolver falhas. Mudanças incompatíveis exigem estratégia de migração. PR, CI e ADR registram decisões e validações; os labs documentam evidências de execução.

## Leitura complementar

- [Testes da POC](../../tests/README.md)
- [Etapa 05 — Instalação e atualização](../../labs/05-terraformstack-runtime.md)
- [Etapa 06 — Simulator](../../labs/06-terraformstack-simulator.md)

---

[Aula anterior](32-observabilidade-de-recursos.md) · [Índice](README.md) · [Próxima aula](34-gitops-bridge-e-outputs.md)
