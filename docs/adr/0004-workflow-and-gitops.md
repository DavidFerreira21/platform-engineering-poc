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
