# Lab 03B — Registro de UAT e acesso temporário à AWS

[← Lab 03](03-crossplane-core.md) · [Índice dos labs](README.md)

O Platform Team registra uma conta UAT existente e confirma, via STS, que o perfil do IAM Identity Center aponta para ela. O teste usa apenas consultas à AWS. A autenticação do Provider AWS no Hub tem um critério de aceite próprio antes do primeiro ObjectStorage.

## Pré-requisitos

- Lab 03 concluído e contexto `kind-platform-management` disponível.
- AWS CLI v2, `kubectl` e Python 3 para inspecionar o discovery OIDC do Hub.
- Uma conta membro UAT com acesso pelo IAM Identity Center. Localize o ID da conta e o permission set no portal de acesso AWS. Confirme a região de trabalho com o Platform Team.

## 1. Registrar o destino localmente

A partir da raiz do monorepo, copie o [template de UAT](../foundation/config/entity-a/uat/environment.example.yaml) para fora do Git. Preencha `aws.accountId` e `aws.region` com os valores confirmados. Preserve o ID entre aspas.

```bash
uat_dir="${XDG_CONFIG_HOME:-$HOME/.config}/platform-engineering-poc/entity-a/uat"
mkdir -p "$uat_dir"
if [[ ! -e "$uat_dir/environment.yaml" ]]; then
  install -m 600 foundation/config/entity-a/uat/environment.example.yaml "$uat_dir/environment.yaml"
fi
```

Esse arquivo registra o destino lógico. O perfil SSO fica na configuração da AWS CLI e o cache da sessão fica no diretório do usuário.

## 2. Autenticar e confirmar conta e região

```bash
aws configure sso --profile platform-poc-uat
```

No assistente, reutilize a sessão do IAM Identity Center quando ela já existir. Selecione a conta membro UAT, o permission set atribuído e a região do registro. Quando precisar iniciar uma sessão nova, use `aws sso login --profile platform-poc-uat`. A sessão pode ser reutilizada até expirar. [Configuração do IAM Identity Center na AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)

No mesmo terminal, informe os valores esperados a partir da cópia local do registro.

```bash
read -r -p 'ID da conta UAT no registro: ' uat_account
read -r -p 'Região UAT no registro: ' uat_region
bash foundation/scripts/verify-uat-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

O script deve retornar `OK` após verificar a conta e a região do perfil e conferir a conta e a role SSO retornadas pelo STS. A região é selecionada na CLI. A resposta do STS confirma a identidade, sem atestar que todos os serviços ou controles regionais permitem operações nesse destino. O teste de identidade usa o permission set disponível na UAT. Cada operação de infraestrutura exige permissões próprias e aprovação específica.

## 3. Provocar uma divergência e recuperar

Escolha uma região diferente da registrada e passe-a como destino esperado. O script deve rejeitar a divergência antes de consultar o STS.

```bash
wrong_region=us-west-2
if [[ "$uat_region" == "$wrong_region" ]]; then wrong_region=us-east-1; fi
bash foundation/scripts/verify-uat-identity.sh platform-poc-uat "$uat_account" "$wrong_region"
bash foundation/scripts/verify-uat-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

O primeiro comando de verificação deve informar que a região do perfil difere do registro. O segundo deve retornar `OK`. A mesma verificação rejeita um perfil configurado para outra conta. Se o STS indicar sessão expirada, renove o login SSO e repita a verificação com o perfil da UAT.

## 4. Inspecionar a identidade de workload do Hub

A sessão SSO é do operador. O caminho proposto para o Provider AWS em `kind` usa `WebIdentity` e uma role dedicada na UAT. O Kubernetes projeta um token rotativo para uma ServiceAccount fixa e o provider o troca por credenciais temporárias no STS.

```bash
kubectl --context kind-platform-management get --raw /.well-known/openid-configuration |
  python3 -c 'import json, sys; d = json.load(sys.stdin); print(d["issuer"]); print(d["jwks_uri"])'
```

No Hub padrão, o issuer e o JWKS usam endereços internos. Para federar com o IAM, a Foundation precisa aprovar e operar um issuer HTTPS cujo discovery e JWKS sejam alcançáveis pela AWS. A trust policy deve limitar a audiência a `sts.amazonaws.com` e o subject à ServiceAccount exata. A API Kubernetes e a chave privada de assinatura ficam privadas. O ProviderConfig e a role serão vinculados pela plataforma ao destino UAT. [OIDC no IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html) · [Autenticação do Provider AWS](https://github.com/crossplane-contrib/provider-upjet-aws/blob/main/AUTHENTICATION.md)

**Gate para o Lab 04:** após aprovar os recursos de identidade, confirmar `GetCallerIdentity` a partir do Hub com a ServiceAccount e a role dedicadas, repetir após a expiração da primeira sessão STS e verificar a rejeição de outra ServiceAccount ou audiência. O teste inicial da role não precisa de permissão para criar buckets. A role de ObjectStorage terá política própria e autorização de provisionamento.

## 5. Limpeza

Esta execução consulta a AWS e o Hub, sem criar recursos faturáveis. A cópia local do Environment e o perfil SSO podem ser reutilizados nos próximos labs. O comando `aws sso logout` encerra sessões SSO armazenadas pela AWS CLI e pode afetar outros perfis.
