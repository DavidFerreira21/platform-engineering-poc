# Foundation — configuração e módulos centrais

A Foundation é de **ownership da plataforma central**. O monorepo reutilizará módulos Terraform e configurações por Entity/Environment.

- Os registros [UAT](config/entity-a/uat/environment.yaml) e [PROD](config/entity-a/prod/environment.yaml) declaram os IDs das contas, a região, o baseline de rede e os clusters EKS previstos. Use o [modelo](environment.example.yaml) para novos registros. A [Etapa 05](../labs/05-registro-ambientes-aws.md) valida os dois destinos com IAM Identity Center.
- A autenticação do Provider AWS no Hub será validada antes de provisionar `ObjectStorage` ou Aurora PostgreSQL. O acesso do operador usa credenciais temporárias do IAM Identity Center.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Roadmap](../docs/learning-roadmap.md)
