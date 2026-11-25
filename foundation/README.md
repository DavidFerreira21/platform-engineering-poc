# Foundation — configuração e módulos centrais

A Foundation é de **ownership da plataforma central**. O monorepo reutilizará módulos Terraform e configurações por Entity/Environment.

- Os registros [UAT](config/entity-a/uat/environment.yaml) e [PROD](config/entity-a/prod/environment.yaml) declaram os IDs das contas, a região, o baseline de rede e os clusters EKS previstos. Use o [modelo](environment.example.yaml) para novos registros. O operador valida UAT via IAM Identity Center na [Etapa 07](../labs/07-registro-ambientes-aws.md). PROD é usado nos testes de isolamento da [exercício de isolamento](../labs/07-registro-ambientes-aws.md#uso-do-registro-prod-na-etapa-27).
- A [Etapa 08](../labs/08-identidade-executor-aws.md) documenta um laboratório reproduzível com SSO, STS, duas roles autorizadas, Secret efêmero em `kind` e Job de verificação; **o backend de state e o Executor de Terraform real ainda precisam ser implementados** antes da Etapa 09. Após o provisionamento dos clusters, a Etapa 13 valida a identidade do Provider AWS no Hub, antes de provisionar `ObjectStorage` ou Aurora PostgreSQL.
- O bootstrap SSO/STS do Terraform Executor é um mecanismo supervisionado para o Lab 08, **não a escolha definitiva para o Provider Crossplane**. A opção em avaliação para o Provider AWS no `kind` é `WebIdentity`, com issuer HTTPS verificável pela AWS, audiência `sts.amazonaws.com` e confiança limitada à ServiceAccount prevista. [OIDC no IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html) · [Autenticação do Provider AWS](https://github.com/crossplane-contrib/provider-upjet-aws/blob/main/AUTHENTICATION.md)

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Roadmap](../docs/learning-roadmap.md)
