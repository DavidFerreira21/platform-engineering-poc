# Integração por outputs e GitOps Bridge

Os produtos de infraestrutura precisam entregar informações consumíveis por outros componentes. Um cluster depende de rede; add-ons podem precisar de identificadores de subnets ou roles. A integração não deve obrigar o consumidor a ler o state inteiro de um motor ou a administrar seus recursos internos.

## Contratos de saída

Um output publicado precisa ter significado, nome, formato, versão, owner, destino e regras de acesso conhecidos. Nem todo output é apropriado para exportação: planos, state, tokens e senhas não devem ser publicados como metadados comuns. Mesmo valores não secretos podem revelar detalhes de topologia e exigem permissões.

O contrato deve separar o produtor, o canal de distribuição e o consumidor. Publicar um valor não configura automaticamente um serviço; consumir um valor não transfere ownership do recurso produzido. Reexecuções devem atualizar os mesmos identificadores e a retirada de valores obsoletos precisa de política própria.

## Aplicação na POC

As raízes AWS Network e EKS expõem outputs selecionados em bridge_parameters. Um publicador versionado prepara a entrega ao Parameter Store da conta de destino, após verificação. A integração real da pipeline AWS e o consumo pelo ESO no Management EKS pertencem à Etapa 17.

O ESO materializa informações em Secrets no cluster em que reconcilia, sem distribuição automática para outros clusters. O Argo CD central não precisa ler Terraform state; onboarding de clusters, IAM e registro por ARN continuam operações próprias. O GitOps Bridge não é uma nova pipeline de IaC nem responsabilidade do TerraformStack Controller.

## Leitura complementar

- [GitOps Bridge Flow](../flows/gitops-bridge.md)
- [Raízes e publicador](../../foundation/stacks/README.md)

---

[Aula anterior](33-testes-e-evolucao-segura.md) · [Índice](README.md) · [Roteiro prático](../learning-roadmap.md)
