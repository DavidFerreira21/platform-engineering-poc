# 🧪 Roteiro prático da POC

[← Início](../README.md) · [Arquitetura](architecture.md) · [Fundamentos](fundamentos/README.md)

Construção incremental e observável: cada lab terá pré-requisitos, comandos, evidências, falha controlada, diagnóstico, recuperação, custo e limpeza. **Primeira implementação: Lab 01 (kind, namespaces, rede local e Traefik)**, seguido por **Lab 02 (Argo CD central)**, **Lab 02B (Argo Workflows instalado via GitOps)** e **Lab 03 (Crossplane instalado via GitOps)**, depois `ObjectStorage` e **`Database` com Aurora PostgreSQL**. A conta inicial será apenas UAT. A decisão de autenticação AWS fica para uma fase posterior; não é necessária para criar o kind. **A ordem pedagógica difere da sequência de provisionamento de uma entidade completa.**

| Fase | Tema | Evidência |
| :--- | :--- | :--- |
| **00** | Documentação e monorepo | Conceitos e decisões consistentes |
| **01** | Hub `kind` + Traefik | API, CoreDNS, namespaces, mapeamento de portas, Ingress HTTP e verificações locais |
| **02** | [Argo CD central (GitOps)](../labs/02-argocd-central.md) | Bootstrap do Argo CD, acesso local, Git como fonte da verdade e aplicações da plataforma |
| **02B** | Argo Workflows no Hub por GitOps | Argo CD instala Argo Workflows e CRDs; executar um Workflow local inofensivo, observar Pods/logs/status e validar reexecução e limpeza, sem AWS ou Terraform |
| **03** | [Crossplane Core por GitOps](../labs/03-crossplane-core.md) | Argo CD instala o Crossplane; observar CRDs, XRDs, Compositions e status sem AWS |
| **03B** | UAT: registro de conta e autenticação AWS | Configurar arquivo de Environment e método de autenticação a definir; validar STS **antes de qualquer operação** |
| **04** | `ObjectStorage` real | XRD/Composition/Provider AWS, bucket de laboratório, `Ready`, drift e teardown |
| **04B** | `Database` com Amazon Aurora PostgreSQL | Contrato de Database e cluster Aurora PostgreSQL em UAT, sujeito a custo, rede, IAM, plano e teardown autorizados |
| **05** | Entity, Environment e Domain (registros lógicos) | Schema e validação de identidade/ownership/destino, **sem depender de Domain Bootstrap** |
| **06** | Argo CD central + primeira Capability de Domain | GitOps de `ObjectStorage` → Crossplane → AWS, com destino autorizado e isolamento |
| **07** | Foundation Terraform reutilizável | Config central por Entity/Environment, plan/state isolado |
| **08** | Integração Argo Workflows + gates + Terraform Executor | Reutilizar o Lab 02B; validar plan/apply, state seguro, aprovação e auditoria |
| **08B** | `TerraformRun` CRD + Controller (**propostos**) | Validação, status, idempotência, disparo único do Argo Workflow e gates (primeiro com executor simulado) |
| **09** | Entity Core e EKS Auto Mode incremental | Criar **um EKS por vez** a partir do Foundation/Core, validar identidade, custo, rede e controles PCI |
| **10** | Argo CD local + configuração do Management EKS | Management EKS gerencia GitOps para EKS PCI/Não PCI autorizados; selecionar Operators progressivamente, validar conectividade e RBAC |
| **10B** | Smoke test técnico do Data Plane com app mínima | **Somente após EKS + Argo CD local funcionarem:** subir app HTTP mínima, preferencialmente Python + Flask, de **repositório separado** e verificar deploy, conectividade, health check e rollback simples; sem portal, Golden Path ou DX |
| **10C** | Validação de ownership e permissões por persona | Simular Platform Team, Federated DevOps e Application Team; PRs autorizados/negados por path e repositório, AppProjects, RBAC e IAM |
| **11** | UAT/PROD segregados | Evoluir para duas contas AWS separadas, com isolamento de identidade, GitOps, acesso e recursos |
| **12** | Observability Plane — OpenTelemetry / Grafana / Loki / Tempo | Implementar coleta/encaminhamento nos spokes e backends centrais; usar a aplicação mínima, workloads sintéticos e os controllers para provar logs/traces/métricas, alertas e auditoria |
| **13** | Segurança, custo e recuperação — consolidação | Validar gates, isolamento PCI/Não PCI, custos, políticas, falhas, recuperação e teardown; controles mínimos são exigidos **desde os primeiros labs** |
| **14** | Extensão opcional: criação de contas | Justificar contrato/controller de lifecycle se houver necessidade real; não bloquear fases finais |
| **14B** | Reavaliar Domain Bootstrap (opcional) | ADR preservado; decidir se esse gerador será necessário, **sem implementá-lo por padrão nem bloquear DX** |
| **15** | Developer Experience — canais e padrões de consumo | Com plataforma funcional, definir/validar experiência para as três personas: catálogo, abstrações, Application Capabilities, Backstage/CLI/Git e padrões de uso. Não exigir Backstage como gateway |
| **16** | Golden Path / integração final da plataforma | **Última etapa:** demonstrar jornada de onboarding e entrega com a aplicação mínima já usada no Lab 10B, repositório separado, padrões/Starter Kit leves, CI/CD, GitOps, Capabilities, políticas e observabilidade; medir carga cognitiva **sem desenvolver aplicação complexa** |

**Macroetapas / dependências:** (1) Hub/Control/Workflow; (2) AWS/Capabilities/Foundation/TerraformStack e Data Plane/EKS; (3) GitOps local e **smoke test mínimo**; (4) UAT/PROD e observabilidade; (5) segurança e extensões opcionais; (6) **Developer Experience**; (7) **Golden Path / demonstração integrada final**. **Não** antecipar DX/Starter Kit para habilitar testes de EKS, nem esperar DX para observar workloads. As fases numeradas 14/14B são opcionais e não bloqueiam 15/16.

**Ordem dos componentes no Hub:** `kind` e Traefik → Argo CD central → **Argo Workflows via Argo CD** → Crossplane Core via Argo CD. O Lab 02B instala e observa o motor com um fluxo de teste local; integração com Terraform, gates humanos e permissões AWS fica para o Lab 08. Não confundir instalar o workflow controller com habilitar execução de infraestrutura real.

**Não implementar `AWSEnvironment` CRD ou Environment Controller próprio como pré-requisito inicial.** Environment e Domain são registros lógicos; somente Capabilities precisam de XRD no primeiro ciclo. **Domain Bootstrap está em reavaliação e adiado**, não é pré-requisito dos labs. **O Argo CD local** aplica namespaces/policies e workloads Kubernetes. `Database` seguirá `ObjectStorage`, com Aurora PostgreSQL, custos e teardown aprovados. `TerraformRun` e CI/CD externo são propostas, **não componentes implantados**.

**Duas utilizações distintas da mesma aplicação mínima:** a primeira é **smoke test técnico do Data Plane no Lab 10B**, depois de provisionar EKS e configurar Argo CD local (sem experiência de portal/Golden Path). A segunda é a **jornada final no Lab 16**, somente **depois** de arquitetura de DX do Lab 15 e de observabilidade/governança. O app é preferencialmente **Python + Flask**, de repositório separado, sem foco em desenvolvimento. Não implementar um produto de software nem um gerador sofisticado. **Platform Team** cuida da plataforma/Core; **Federated DevOps Team** mantém Domain Infrastructure no monorepo; **Application Team** mantém código, CI e manifests no repo separado. Lab 10C valida ownership. Ver [ADR-0007](adr/0007-personas-and-developer-experience.md).

**Conta UAT:** registrar o destino em `foundation/config/entity-a/uat/environment.example.yaml` (template sem dados reais). Conta, região, credenciais e método de autenticação ainda precisam ser definidos e validados no início dos labs AWS. Nenhuma chave ou segredo deve ser versionado.

**Segurança:** toda operação AWS precisa de conta/região confirmadas via STS, IAM mínimo, estimativa de custos, aprovação específica e plano de remoção. O Hub permanece em `kind` inicialmente e os EKS reais favorecem Auto Mode, sujeitos a compatibilidade e PCI. Não duplicar ownership Terraform/Crossplane, versionar segredos nem executar `destroy` indiscriminadamente.

[Arquitetura](architecture.md) · [Decisões](adr/README.md)
