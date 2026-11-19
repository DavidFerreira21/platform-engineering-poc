# Foundation — configuração e módulos centrais

A Foundation é de **ownership da plataforma central**. O monorepo reutilizará módulos Terraform e configurações por Entity/Environment.

- [Template UAT](config/entity-a/uat/environment.example.yaml): registra **destino e perfis**, não credenciais. É um exemplo, **não um ambiente configurado ou ativo**.
- O método de autenticação da AWS será definido **antes** de qualquer execução de `ObjectStorage` ou Aurora PostgreSQL. Não usar IRSA/Pod Identity como premissa em Hub `kind`, nem gravar chaves/segredos no repositório.
- Este primeiro PR não inclui módulos Terraform nem operações AWS.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Roadmap](../docs/learning-roadmap.md)
