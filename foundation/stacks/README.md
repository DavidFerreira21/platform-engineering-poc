# Preparação do Hub e das spokes

As raízes `network` e `eks` usam o módulo `configuration` para compor os YAMLs de `foundation/config` (global, entidade e ambiente). Executar a partir do checkout completo da revisão fixada. O profile `platform/hub` habilita apenas o Hub. Em `entity-a/uat`, Management e Non-PCI são independentes e PCI começa desligado. O override inicial de UAT habilita Management e mantém Non-PCI desligado até a Etapa 15. Desligar a flag de um cluster já gerenciado propõe sua remoção no próximo plano.

## Bootstrap manual do Hub

A raiz [`platform-hub`](platform-hub/main.tf) compõe Network, EKS e o bucket de state do TerraformStack em um único plano. A mesma raiz ativa o Argo CD gerenciado quando recebe a configuração do IAM Identity Center. O Platform Team executa essa raiz manualmente e mantém seu state sempre local, fora do checkout, com backup protegido. As raízes das spokes carregam YAML via Configuration e usam seus backends autorizados.

### Preparar o acesso e as entradas

1. Configurar o perfil `platform-hub` com `aws configure sso --profile platform-hub`, selecionando a conta de plataforma, a role autorizada e `us-east-1` como região padrão. O provider do Hub seleciona esse perfil. UAT e PROD usam perfis próprios.
2. Para personalizar as entradas, copiar o [modelo](platform-hub/hub.tfvars.example) para `hub.auto.tfvars` no diretório da stack. O Terraform carrega esse arquivo automaticamente, e o Git o ignora. Usar `account_id` para fixar a conta esperada. Por padrão, o Hub usa as AZs `us-east-1a/b/c`, o CIDR `10.80.0.0/16` e a versão Kubernetes mais recente em suporte padrão. Para ativar o Argo CD, configurar a instância e o grupo autorizado em `identity_center`.
3. O [`backend.tf`](platform-hub/backend.tf) declara o backend local. Em `backend_override.tf`, ignorado pelo Git, informar o caminho absoluto do state protegido fora do checkout. Para uma instalação existente, informar o caminho do state corrente e conferir os recursos com `terraform state list`. Se houver recursos gerenciados por outro state, reconciliar a propriedade antes de aplicar.

```hcl
terraform {
  backend "local" {
    path = "/caminho/protegido/terraform.tfstate"
  }
}
```

Executar `terraform init` dentro da stack no primeiro uso e quando uma dependência mudar.

### Planejar e aplicar

Com o SSO ativo, conferir a conta e rodar da raiz do repositório:

```bash
aws sts get-caller-identity --profile platform-hub --no-cli-pager

cd foundation/stacks/platform-hub
terraform validate
terraform plan
terraform apply
```

Conferir conta, região, alterações e custo. O output `aws_target` identifica o destino. O `apply` gera seu próprio plano e pede confirmação. Conferir esse plano antes de autorizar a operação. Em uma retomada, verificar se a VPC e o NAT existentes são preservados.

Para operar o Hub pelo terminal local durante o bootstrap, definir `api_public_access_cidrs` em `hub.auto.tfvars` com o IP público de saída do operador no formato `/32`. O endpoint privado continua disponível para o tráfego interno, e o `kubectl` usa a Access Entry da role. Restaurar a lista vazia quando houver conectividade privada para a operação. [Acesso aos endpoints do EKS](https://docs.aws.amazon.com/eks/latest/userguide/config-cluster-endpoint.html)

Verificar o cluster e, quando configurada, a capability em estado `ACTIVE`. Usar o output `argocd.server_url` para validar o login. O registro do Hub por ARN, as permissões de deploy e o primeiro sync seguem na [Etapa 02](../../labs/02-argocd-central.md). [Criação da capability Argo CD](https://docs.aws.amazon.com/eks/latest/userguide/create-argocd-capability.html)

### Backend das execuções do TerraformStack

O output `executor_backend` fornece bucket, região e configuração de locking para cadastrar destinos autorizados. O bucket usa versionamento, SSE-S3, bloqueio de acesso público, propriedade de objetos pelo dono do bucket, HTTPS obrigatório e proteção contra destruição pelo Terraform. As execuções usam chaves segregadas por destino e stack.

Na Etapa 12, validar a identidade temporária e o vínculo de backend do executor. Conceder `s3:ListBucket` no prefixo necessário, `s3:GetObject` e `s3:PutObject` na chave de state e essas mesmas permissões mais `s3:DeleteObject` no arquivo `.tflock`. As permissões de infraestrutura têm escopo próprio. [Backend S3 do Terraform](https://developer.hashicorp.com/terraform/language/backend/s3) · [Contrato de execução](../../docs/flows/terraformstack-execution-contract.md)

## Ordem de execução nas spokes

1. Usar a raiz `network` em cada spoke com os valores da Entity e do Environment. Selecionar CIDRs sem sobreposição quando houver conectividade entre VPCs. Informar `entity`, `environment`, `region` e `account_id` conforme o registro confiável, e três AZs em ordem estável. A identidade temporária do executor precisa estar assumida na conta de destino. O provider confere o account ID.
2. Inicializar cada raiz com seu próprio backend S3 (`bucket`, `region`, `key` por entidade/ambiente/stack e `use_lockfile=true`). Executar a raiz da rede da spoke pelo fluxo de plan, revisão e apply aprovado. Ela divide a VPC IPv4 em blocos iguais para web PCI, web Non-PCI, database PCI e database Non-PCI. Cada grupo recebe três subnets privadas, uma por AZ.
3. Com `network.nat.enabled: true`, uma 13ª subnet, pública, na primeira AZ hospeda o NAT zonal com EIP e rota para o Internet Gateway. Apenas as duas tabelas web usam rota padrão para o NAT. Um NAT compartilhado reduz o número de gateways e concentra a saída numa AZ. NAT, IPv4 público e tráfego entre AZs podem gerar custos. As tabelas de database mantêm a rota local. As quatro classes de subnet identificam destinos para os controles de acesso e testes de isolamento da Etapa 30.
4. Para a raiz EKS, passar os outputs Network `vpc_id` e `subnets`, uma versão Kubernetes suportada, os CIDRs privados de gestão em `api_client_cidrs` e a role administrativa autorizada em `admin_role_arn`. O plano cria EKS Auto Mode no grupo de subnets do perfil habilitado. Preparar acesso ao endpoint privado desde o ambiente de bootstrap/onboarding. Em UAT o Hub coordena a execução pelo TerraformStack, reutilizando as raízes e os valores versionados.
5. Na Etapa 17, a pipeline AWS integra o publicador abaixo após `verify`, com os parâmetros do destino confiável. A flag YAML `publishParameterStore` de cada stack controla o conteúdo de `bridge_parameters`. O padrão é desligado para o Hub e o override UAT habilita a publicação na spoke. Os outputs Terraform continuam disponíveis em ambos. O publicador usa AWS CLI com identidade temporária restrita ao prefixo do destino. Repetir a etapa após uma falha parcial atualiza os mesmos caminhos. Desabilitar a flag interrompe novas escritas e preserva os valores já publicados. A limpeza dos parâmetros obsoletos é explícita.

```bash
set -o pipefail
terraform output -json bridge_parameters | python3 ../../scripts/publish-bridge-outputs.py \
  --entity "$ENTITY" --environment "$ENVIRONMENT" --region "$AWS_REGION" --account-id "$ACCOUNT_ID"
```

O contrato exporta ID da VPC, listas de subnets por classe em CSV, nome e endpoint de cada cluster habilitado. ESO no Management EKS é o consumidor previsto na Etapa 17. Identidades específicas de add-ons, instalação de ESO e regras de acesso de aplicações entram nos respectivos labs. [Fluxo do Bridge](../../docs/flows/gitops-bridge.md)
