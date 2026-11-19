# ADR-0010: Provisionamento AWS real e evolução progressiva dos ambientes

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto

A POC é um laboratório prático e deverá provisionar recursos **reais na AWS**, não apenas simulá-los. O Hub Kubernetes pode executar localmente com `kind`. Não precisamos criar contas AWS automaticamente no início, mas queremos evoluir até esse objetivo.

## Decisão

1. **Primeira etapa: conta real existente.** O usuário fornece/autentica uma conta AWS e sua região. Validar a identidade com `sts get-caller-identity`, roles/permissões, políticas e viabilidade de acesso do Crossplane executando em `kind`. Provisionar Capabilities e infraestrutura reais com limite de custo; usar mocks apenas para testes isolados.
2. **Segunda etapa: UAT e PROD reais.** O usuário fornece duas contas existentes, com autenticação e isolamento próprios. `AWSEnvironment` passa a representar/adotar cada conta, sem acionar sua criação. Definir credenciais, remote state e recursos por ambiente sem compartilhamento indevido.
3. **Terceira etapa: criação automática de contas.** Evoluir Environment Controller + Argo Workflows + Terraform para criar/adotar contas e fundações, lidando com dependências humanas, validações e aprovações. Não incluir esse provisionamento automático na primeira etapa.
4. O alvo segue com UAT/PROD, Environment Management EKS, EKS PCI, EKS Non-PCI e AWS Runtime, mas **não é obrigatório criar os sete clusters no início**. O Hub `kind` é compatível com a AWS real desde que os mecanismos de credenciais e rede sejam validados.
5. O usuário autoriza esta **direção arquitetural**, não um `apply` indiscriminado. Cada execução AWS requer confirmação de conta/região, permissões mínimas, plano, risco/custo, autorização explícita da operação e estratégia de remoção. Não versionar segredos, kubeconfigs ou state; evitar credenciais permanentes em manifests, env vars persistidas ou Git.

## Consequências

- Preferir primeiro recursos de baixo custo e fácil limpeza (por exemplo S3 com proteções adequadas) e só depois RDS/compute, considerando custo contínuo e teardown.
- Especificar autenticação realista do provider AWS a partir de `kind` (credenciais temporárias/role/federação), sem presumir IRSA ou EKS Pod Identity em `kind`.
- `AWSEnvironment` deve suportar **adoption-first** e gestão de recursos numa conta existente antes de implementar o fluxo de abertura de conta.
- Documentar impacto de billing, quotas, políticas de conta, state do Terraform e salvaguardas contra `destroy` de recursos permanentes.

## Complemento posterior — ADR-0011
O provisionamento dos clusters EKS reais em contas existentes adotará inicialmente **EKS Auto Mode**. A decisão não antecipa o provisionamento de contas nem obriga criar todos os clusters no primeiro laboratório. Consultar [ADR-0011](0011-eks-auto-mode.md).
