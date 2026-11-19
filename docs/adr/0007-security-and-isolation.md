# ADR-0007: Isolamento de ambientes e proteção de recursos

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
Os ambientes UAT e PROD são contas AWS independentes e PCI/Non-PCI podem ter requisitos diferentes de gestão administrativa.

## Decisão
Aplicar menor privilégio, identidades separadas, RBAC/admission, revisões de PR, gates auditados e proteção de dados. Um Argo CD gerenciando PCI e Non-PCI só será permitido se a avaliação de segurança autorizar. Excluir AWSEnvironment nunca deve excluir automaticamente a conta AWS. Não versionar segredos, estados Terraform ou identificadores reais.

## Consequências e pontos de atenção
Uma instância local de Argo CD pode precisar ser separada por fronteira PCI. Ações AWS reais, custos e operações destrutivas exigem autorização explícita.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.
