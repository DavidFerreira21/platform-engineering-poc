# Foundation — configuração e módulos centrais

A Foundation é de **ownership da plataforma central**. O monorepo reutilizará módulos Terraform e configurações por Entity/Environment.

- [Registro UAT](config/entity-a/uat/environment.yaml) e [modelo](config/entity-a/uat/environment.example.yaml) mantêm a referência lógica da conta, a região e os perfis. O [Lab 03B](../labs/03b-uat-e-autenticacao-aws.md) valida o destino com IAM Identity Center.
- A autenticação do Provider AWS no Hub será validada antes de provisionar `ObjectStorage` ou Aurora PostgreSQL. O acesso do operador usa credenciais temporárias do IAM Identity Center.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Roadmap](../docs/learning-roadmap.md)
