# ADR-0008 — Argo CD central gerenciado e conectividade entre VPCs

[← Decisões](README.md) · [Arquitetura](../architecture.md) · [Roadmap](../learning-roadmap.md)

**Status:** aceito — implementação e validação AWS pendentes.

## Contexto

A POC possui três contas AWS (Platform, UAT e PROD), cada uma com sua VPC isolada e saída via NAT Gateway. Não haverá VPC Peering nem Transit Gateway entre elas. As APIs dos clusters EKS permanecem privadas. O Argo CD autogerenciado no Hub não consegue alcançar essas APIs somente por meio de NAT ou do registro via CLI.

## Decisão

- No Hub `kind`, manter Argo CD autogerenciado, instalado por Helm, e os laboratórios existentes.
- No Hub EKS Auto Mode, adotar **Amazon EKS Capability for Argo CD** como Argo CD central e mecanismo de acesso GitOps às APIs privadas dos EKS em Platform, UAT e PROD. Trata-se de uma capability gerenciada, não de um EKS Managed Add-on nem de uma função exclusiva do Auto Mode.
- Registrar explicitamente o Hub e os clusters remotos por ARN. Criar Access Entries nos destinos para a IAM Capability Role da conta Platform e conceder RBAC adequado ao escopo. O serviço AWS fornece conectividade gerenciada; a CLI ou o Secret de cluster apenas registra o destino.
- Usar ApplicationSets com labels de Entity, Environment e perfil para distribuir add-ons e configurações de plataforma. Usar AppProjects e permissões de destino para limitar o alcance.
- Manter Argo CD local no Management EKS de cada Environment para Domain Applications, com acesso aos EKS de sua própria VPC conforme autorização. O Argo central instala e mantém sua configuração-base; nenhum recurso terá dois reconciliadores GitOps.
- Crossplane, Argo Workflows e TerraformStack Controller permanecem autogerenciados no Hub. A conectividade da capability não é um túnel genérico para Pods, kubectl, bancos de dados ou telemetria. Fluxos de observabilidade e acesso operacional privado exigem soluções próprias, sem presumir roteamento entre VPCs.

## Implementação e evidências

| Etapa | Entrega e aceite |
| --- | --- |
| 09–10 | Confirmar disponibilidade regional, custo, IAM Capability Role, AWS IAM Identity Center e acesso operacional ao Hub privado para bootstrap. Criar capability; registrar o próprio Hub por ARN, adaptar destinos dos manifests (sem `kubernetes.default.svc`), configurar Git/AppProjects e repetir os aceites TerraformStack e Crossplane. Não instalar um segundo Argo CD central. |
| 14–16 | Criar Access Entries nos EKS de UAT, registrar ARNs e labels e sincronizar um recurso de teste por ApplicationSet. Comprovar acesso com endpoints privados e VPCs isoladas, além de negar acesso fora do escopo. Instalar Argo CD local e validar seus destinos na VPC UAT. |
| 30–31 | Repetir os controles para PCI e PROD antes de liberar os destinos; verificar separação de ownership, identidade, autorização e ausência de rotas entre VPCs. |

O registro declarativo usa Secret em `argocd` com label `argocd.argoproj.io/secret-type: cluster` e `stringData.server` igual ao ARN do EKS, não à URL da API. A escolha de RBAC detalhado, bootstrap administrativo, parâmetros de IaC e custo será validada nas etapas correspondentes. Não versionar identificadores reais fora das exceções previstas em AGENTS.md. Esta decisão não autoriza criação ou alteração de recursos AWS.

## Alternativas

VPC Peering/TGW não atendem à premissa de isolamento. Argo CD Agent introduziria outro componente e endpoint de comunicação; não será necessário para o fluxo escolhido. Argo CD autogerenciado continua adequado ao laboratório local e aos Management EKS, não ao acesso central entre VPCs isoladas.

## Referências

- [AWS — Registro de clusters EKS, privados e cross-account](https://docs.aws.amazon.com/eks/latest/userguide/argocd-register-clusters.html)
- [AWS — Criação da capability Argo CD](https://docs.aws.amazon.com/eks/latest/userguide/create-argocd-capability.html)
- [AWS — Comparação com Argo CD autogerenciado](https://docs.aws.amazon.com/eks/latest/userguide/argocd-comparison.html)
