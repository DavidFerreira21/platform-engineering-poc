# Etapa 04 — API do TerraformStack via GitOps

## Objetivo

Instalar o contrato `TerraformStack` no Hub e conferir o que a Kubernetes API aceita.

## O que será feito

Sincronizar a Application do CRD e usar as solicitações do Simulator em validação no API server.

## Por que fizemos assim

O CRD `v1alpha1` é o contrato instalado para as próximas etapas. `sourceRef` declara o repositório, o caminho e a versão. O mapa opcional `context` acompanha a intenção. `workflowRef.name` seleciona um alias cadastrado no controller; a API valida o formato, não a existência da integração. O subrecurso `status` recebe o estado da execução. O controller da próxima etapa usa o mesmo contrato para orquestrar a pipeline.

> A instalação e o aceite desta etapa no Hub EKS estão pendentes. Os critérios abaixo descrevem a entrega prevista.

## 1. Sincronizar e validar

Revisar o AppProject, o destino por ARN e o diff da Application do CRD antes do sync. Confirmar que a API reconhece o recurso namespaced `TerraformStack` em `platform.example.io/v1alpha1` e que o subrecurso de status está disponível.

## O que observar

O CRD fica `Established`. A API aceita um exemplo válido em dry-run e rejeita campos incompatíveis com o schema. Conferir o destino antes de testar o manifesto.
