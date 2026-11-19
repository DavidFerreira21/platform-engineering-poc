# Roteiro de aprendizado — Platform Engineering POC

**Método:** construir progressivamente, entender cada conceito, observar os recursos em execução e reproduzir falhas e recuperações. A topologia-alvo de sete clusters não significa que devemos provisioná-los no início.

| Fase | Implementação e conceito | Evidência verificável | Falha e recuperação |
| --- | --- | --- | --- |
| 00 | Documentação, ADRs e estrutura do Git | Links e nomenclaturas consistentes | Encontrar referência inválida |
| 01 | Platform Management Cluster em `kind`, API e RBAC | Observar recursos, eventos e permissões | Manifest inválido e acesso negado |
| 02 | `AWSEnvironment` CRD e controller com AWS mockada | `uat` evolui condições de status | Erro de validação e retry |
| 03 | Argo Workflows e Terraform `plan` | Workflow vinculado ao Environment | Falha de execução e retomada |
| 04 | Gates manuais, espera de IPs, rotas e certificados | Aprovação autorizada + validação técnica | Aprovação indevida bloqueada |
| 05 | Argo CD e GitOps | Commit → CR → controller → status | Drift do spec corrigido |
| 06 | UAT/PROD com Management EKS, PCI e Non-PCI simulados | Configurações isoladas por ambiente | Bloqueio de destino incorreto |
| 07 | Domain, namespaces e relacionamentos | Múltiplos domínios no mesmo ambiente | Acesso cruzado negado |
| 08 | Capabilities `Database` e `ObjectStorage` | Instâncias e status de recursos tipados | Parâmetros inválidos |
| 09 | Observability Plane ponta a ponta | Diagnóstico com logs, métricas e traces | Localizar workflow com falha |
| 10 | Backstage, CLI e AI/MCP | Mesmos contratos consumidos por canais diferentes | Usuário sem permissão rejeitado |
| 11 | Crossplane e AWS real opcional | Comparar motores sem dupla gestão | Simular conflito de ownership |
| 12 | Segurança, PCI e recuperação | Procedimento de recuperação testado | Falha do cluster de gerenciamento |

## Contrato de qualidade de cada lab
Explicar problema e conceito, arquitetura antes/depois, pré-requisitos, passos manuais, comandos, resultados esperados, observação de eventos e logs, falha proposital segura, diagnóstico, recuperação, critérios de conclusão, automação e limpeza. Referenciar fontes canônicas do monorepo; não duplicar implementação.

## Restrições
- Começar com `kind` e mocks, evoluindo para papéis de clusters adicionais.
- Não simular uma integração com chamados como se estivesse implementada.
- `terraform plan` não significa autorização de `apply`.
- Qualquer criação de conta AWS, recursos reais, custos e destruição exige aprovação explícita.
- Backstage, AI/MCP, Crossplane e observabilidade completa são componentes da arquitetura-alvo e podem ser introduzidos depois.
