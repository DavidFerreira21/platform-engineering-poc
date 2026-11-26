# Foundation — configuração e módulos centrais

A Foundation é de **ownership da plataforma central**. O monorepo reutiliza stacks Terraform, o módulo compartilhado [Configuration](modules/configuration/main.tf) e valores [Network](config/network.yaml)/[EKS](config/eks.yaml) por Entity/Environment.

- [Preparação das raízes Network e EKS](stacks/README.md) com os módulos compartilhados, perfis Hub/spoke e outputs do Bridge.
- Os registros [UAT](config/entity-a/uat/environment.yaml) e [PROD](config/entity-a/prod/environment.yaml) declaram os IDs das contas, a região, o baseline de rede e os clusters EKS previstos. Use o [modelo](environment.example.yaml) para novos registros. O operador valida UAT via IAM Identity Center na [Etapa 11](../docs/learning-roadmap.md#etapa-11--registro-de-uat). PROD é usado nos testes de isolamento da [exercício de isolamento](../docs/learning-roadmap.md#etapa-31--isolamento-por-environment).

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Roadmap](../docs/learning-roadmap.md)

A Etapa 06 usa a raiz Simulator em Kubernetes. As Etapas 09–10 preparam a conta de plataforma e o Hub EKS. A [Etapa 12](../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state) valida a identidade temporária e o backend do executor no Hub EKS para o primeiro apply da spoke na Etapa 13.
