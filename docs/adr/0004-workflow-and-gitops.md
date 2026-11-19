# ADR-0004: GitOps, workflows e execução com responsabilidades distintas

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
A Experience Layer terá Backstage, CLI, GitOps e AI/MCP, sem criar fontes conflitantes de estado desejado.

## Decisão
Git é a fonte da verdade declarativa; Argo CD sincroniza manifests com o Kubernetes. Backstage/CLI/AI-MCP devem preferencialmente produzir PRs para alterações permanentes. Platform Controllers gerenciam lifecycle; Argo Workflows coordena passos e aprovações; Terraform executa infraestrutura; Crossplane é um reconciliador adicional futuro. Argo CDs locais podem gerenciar PCI e Non-PCI conforme segurança.

## Consequências e pontos de atenção
Separar Argo CD de Argo Workflows. Evitar concorrência no spec e dupla propriedade Terraform/Crossplane. CI será avaliado separadamente.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.

## Definição posterior (ADR-0009)

A classificação de Crossplane como reconciliador *futuro* fica **substituída**: a adoção do Crossplane foi decidida para Domain/Capabilities e sua localização lógica é o **Control Plane**. Argo Workflows, Terraform Executor, gates e Argo CD permanecem no Workflow Plane. O Environment Controller próprio aciona workflows; o Crossplane reconcilia seus recursos diretamente. [Detalhamento](0009-crossplane-e-reconciliacao.md).
