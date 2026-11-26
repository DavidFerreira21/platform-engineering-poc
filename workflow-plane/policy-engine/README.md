# Policy Engine — módulos e perfis

[← Workflow Plane](../README.md) · [TerraformStack](../../docs/flows/terraformstack.md)

O OPA avalia o JSON do Terraform plan e o contexto confiável do catálogo. O WorkflowTemplate é genérico: não contém regras de Network, EKS ou de outros produtos.

## Organização

- [engine/decision.rego](engine/decision.rego): contrato de saída, agregação, cobertura e precedência.
- [profiles/profiles.yaml](profiles/profiles.yaml): composição declarativa de módulos por perfil; exemplos de Network e EKS comentados.
- [modules/common/policy.rego](modules/common/policy.rego): regras transversais obrigatórias.
- [modules/simulator/policy.rego](modules/simulator/policy.rego): regras do exercício local, com CIDR, identidade do ConfigMap e drift.
- [kustomization.yaml](kustomization.yaml): pacote imutável de políticas para o Workflow.

## Criar um módulo

Cada módulo utiliza um package `policy_modules.<nome>` e publica `result` com `decision`, `reason` e `covered` (lista de endereços Terraform validados). As decisões aceitas são AutoApply, RequireApproval, Deny, NotApplicable e Error. Um módulo transversal pode impor uma restrição sem cobrir recursos. NotApplicable não autoriza mudanças. O engine exige cobertura explícita de todos os recursos alterados. Módulos futuros devem tratar somente seu escopo, inclusive recursos mistos e endereços não reconhecidos.

## Criar um perfil

O Platform Team adiciona os módulos ao pacote Kustomize e ao perfil ativo em YAML. Exemplo futuro:

```yaml
profiles:
  network:
    - common
    - network
  eks:
    - common
    - eks
  network-xpto:
    - common
    - network
    - xpto
```

Não há regras reais AWS nesses exemplos. Somente `simulator` está ativo. O catálogo confiável de cada stack informa `policyProfile`; o consumidor não escolhe o perfil no CRD. Perfil ausente, módulo inválido, erro OPA e mudanças sem cobertura bloqueiam a execução. Error > Deny > RequireApproval > AutoApply; NoChanges é tratado pelo engine. Destroy continua exigindo aprovação.

## Rastreabilidade

O contexto com o perfil é vinculado ao hash do plan JSON e o resultado inclui perfil, decisões por módulo e digest SHA-256 de todos os arquivos de produção montados. ConfigMaps imutáveis com hash no nome preservam as regras usadas pelo Workflow; não remover versões ainda utilizadas por execuções ativas. O digest identifica o pacote, não substitui a autorização do plano. A troca de política não revoga automaticamente um Workflow já iniciado; mudanças de perfil no catálogo integram a intenção reconciliada.

## Testes

```bash
docker run --rm -v "$PWD/workflow-plane/policy-engine:/policy:ro" \
  openpolicyagent/opa:1.11.0-static test /policy/engine /policy/modules /policy/profiles -v
kubectl kustomize workflow-plane/argo-workflows/terraformstack
```

Os testes do Simulator preservam os cenários originais e verificam composição, precedência, perfil inválido e ausência de cobertura. O Lab 06 continua sendo o aceite de integração.
