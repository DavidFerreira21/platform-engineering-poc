# Preparação do Hub e das spokes

As raízes `network` e `eks` usam o módulo `configuration` para compor os YAMLs de `foundation/config` (global, entidade e ambiente). Executar a partir do checkout completo da revisão fixada. O profile `platform/hub` habilita apenas o Hub. Em `entity-a/uat`, Management e Non-PCI são independentes e PCI começa desligado. O override inicial de UAT habilita Management e mantém Non-PCI desligado até a Etapa 15. Desligar a flag de um cluster já gerenciado propõe sua remoção no próximo plano.

## Ordem de execução

1. Usar a rede da conta de plataforma para o Hub e uma rede própria em cada spoke. Selecionar CIDRs sem sobreposição quando houver conectividade entre VPCs. Informar `entity`, `environment`, `region` e `account_id` conforme o registro confiável, e três AZs em ordem estável. A identidade temporária do executor precisa estar assumida na conta de destino. O provider confere o account ID.
2. Inicializar cada raiz com seu próprio backend S3 (`bucket`, `region`, `key` por entidade/ambiente/stack e `use_lockfile=true`). Executar a raiz Network pelo fluxo de plan, revisão e apply aprovado. Ela divide a VPC IPv4 em blocos iguais para web PCI, web Non-PCI, database PCI e database Non-PCI. Cada grupo recebe três subnets privadas, uma por AZ.
3. Com `network.nat.enabled: true`, uma 13ª subnet, pública, na primeira AZ hospeda o NAT zonal com EIP e rota para o Internet Gateway. Apenas as duas tabelas web usam rota padrão para o NAT. Um NAT compartilhado reduz o número de gateways e concentra a saída numa AZ. NAT, IPv4 público e tráfego entre AZs podem gerar custos. As tabelas de database mantêm a rota local. As quatro classes de subnet identificam destinos para os controles de acesso e testes de isolamento da Etapa 30.
4. Para a raiz EKS, passar os outputs Network `vpc_id` e `subnets`, uma versão Kubernetes suportada, os CIDRs privados de gestão em `api_client_cidrs` e a role administrativa autorizada em `admin_role_arn`. O plano cria EKS Auto Mode nos grupos web habilitados. Preparar acesso ao endpoint privado desde o ambiente de bootstrap/onboarding. O bootstrap externo usa esta mesma raiz na conta de plataforma. Em UAT o Hub coordena a execução pelo TerraformStack.
5. Após verify, o Workflow pode executar o publicador abaixo com os parâmetros do destino confiável. A flag YAML `publishParameterStore` de cada stack controla o conteúdo de `bridge_parameters`. O padrão é desligado para o Hub e o override UAT habilita a publicação na spoke. Os outputs Terraform continuam disponíveis em ambos. O publicador usa AWS CLI com identidade temporária restrita ao prefixo do destino. Repetir a etapa após uma falha parcial atualiza os mesmos caminhos. Desabilitar a flag interrompe novas escritas e preserva os valores já publicados. A limpeza dos parâmetros obsoletos é explícita.

```bash
set -o pipefail
terraform output -json bridge_parameters | python3 ../../scripts/publish-bridge-outputs.py \
  --entity "$ENTITY" --environment "$ENVIRONMENT" --region "$AWS_REGION" --account-id "$ACCOUNT_ID"
```

O contrato exporta ID da VPC, listas de subnets por classe em CSV, nome e endpoint de cada cluster habilitado. ESO no Management EKS é o consumidor previsto na Etapa 17. Identidades específicas de add-ons, instalação de ESO e regras de acesso de aplicações entram nos respectivos labs. [Fluxo do Bridge](../../docs/flows/gitops-bridge.md)
