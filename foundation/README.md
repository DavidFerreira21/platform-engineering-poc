# Foundation — configuração e módulos centrais

A Foundation é de **ownership da plataforma central**. O monorepo reutiliza stacks Terraform, o módulo compartilhado [Configuration](modules/configuration/main.tf) e valores [Network](config/network.yaml)/[EKS](config/eks.yaml) por Entity/Environment.

- [Preparação das raízes Network e EKS](stacks/README.md) com os módulos compartilhados, perfis Hub/spoke e outputs do Bridge.
- Os registros [UAT](config/entity-a/uat/environment.yaml) e [PROD](config/entity-a/prod/environment.yaml) declaram os IDs das contas, a região, o baseline de rede e os clusters EKS previstos. Use o [modelo](environment.example.yaml) para novos registros. O operador valida UAT via IAM Identity Center na [Etapa 11](../docs/learning-roadmap.md#etapa-11--registro-de-uat). PROD entra no [exercício de isolamento](../docs/learning-roadmap.md#etapa-31--isolamento-por-environment).

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Roadmap](../docs/learning-roadmap.md)

A Etapa 06 usa a raiz Simulator em Kubernetes. As Etapas 09–10 preparam a conta de plataforma e o Hub EKS. A [Etapa 12](../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state) valida a identidade temporária e o backend do executor no Hub EKS para o primeiro apply da spoke na Etapa 13.

## Configuration

Cada stack informa a raiz de configuração, o nome base do YAML, a Entity e o Environment. Com `stack = "EC2"`, o módulo procura `EC2.yaml` na raiz, depois sob a Entity e, por fim, sob o Environment. Utilize a mesma capitalização do arquivo. Os níveis ausentes são ignorados.

Uma entrada como esta fica disponível em `module.configuration.values.ec2.network.cidrs`.

```yaml
ec2:
  instance_type: t3.micro
  network:
    cidrs:
      - 10.80.0.0/16
      - 10.81.0.0/16
```

Overrides em objetos preservam as chaves vizinhas. Um override de `cidrs` substitui a lista inteira. A stack EC2 pode adotar esse YAML e tipar os campos nas suas próprias variáveis. [Testes de herança](modules/configuration/tests/configuration.tftest.hcl)
