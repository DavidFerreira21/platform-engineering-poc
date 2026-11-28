# Identidade temporária e backend do Terraform Executor

A Etapa 12 valida a identidade temporária do Pod executor no Hub EKS da conta da Plataforma e a assunção da role autorizada em UAT. O Platform Team provisiona as roles e restringe suas permissões às operações da etapa. O controller do TerraformStack acompanha o Workflow, enquanto o runner usa a identidade de execução aprovada.

A pipeline AWS deverá resolver conta, região, identidade e vínculo do backend a partir do cadastro autorizado de destino. O primeiro apply da spoke depende da verificação dessas permissões e da negativa para destinos fora do escopo. [Contrato de execução](../../../docs/flows/terraformstack-execution-contract.md)

## Backend do Hub EKS

A Etapa 01 provisiona, na conta da Plataforma, o bucket de state do executor e publica sua configuração no output `executor_backend`. A Etapa 12 valida o acesso da identidade autorizada ao backend, com key e lock segregados por destino e stack. A pipeline ainda precisa ligar o registro do destino, a identidade de execução e o vínculo do backend antes do primeiro apply na spoke. [Stack do Hub](../../stacks/README.md) · [Roteiro da Etapa 12](../../../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state)

[Registro UAT](../../../docs/learning-roadmap.md#etapa-11--registro-de-uat) · [Identidade e state](../../../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state)
