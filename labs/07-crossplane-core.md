# Etapa 07 — Crossplane Core via GitOps

[← Etapa 06](06-terraformstack-simulator.md) · [Guias das etapas](README.md)

## Objetivo

Validar a reconciliação de um recurso composto no Hub e observar como o schema de um XRD controla o que a API aceita.

## O que será feito

- Sincronizar a Application `crossplane-core`, definida no [catálogo do Hub](../platform/hub/addons/catalog/crossplane/crossplane-core.yaml), para instalar o Crossplane Core e seus CRDs.
- Instalar a função de composição do exercício, aplicar o XRD de `LabMessage` e sua Composition e criar um `LabMessage` em `platform-system`. Conferir as condições do recurso e o ConfigMap criado pelo Crossplane.
- Consultar o schema publicado pelo XRD e conferir a mensagem do `LabMessage` no ConfigMap.
- Remover os recursos do exercício e manter o Crossplane Core instalado.

O exercício usa recursos Kubernetes de teste no Hub EKS. Crossplane Core é instalado via GitOps; os manifests da demonstração são aplicados diretamente à API para observar cada etapa. Isso é um smoke test, não o aceite GitOps completo, reservado à Etapa 08.

## Por que fizemos assim

O Crossplane Core é compartilhado pelas etapas. O `LabMessage` usa um XRD namespaced da API v2 e uma Composition de demonstração para tornar a reconciliação visível em um ConfigMap do Hub. Assim é possível rever a validação do contrato e a reconciliação local mesmo em um Hub que já tenha outras Capabilities.

> O Core e a API de demonstração passaram no smoke test no EKS. A instalação pelo catálogo GitOps segue como aceite.

## Aceite no Hub EKS

Depois do Simulator, conferir o destino por ARN, o AppProject e o diff da Application `crossplane-core`. Confirmar acesso ao chart, às imagens e à Function. Sincronizar a API de demonstração e confirmar que o XRD e a Composition local estão disponíveis.

## O que observar

O Core e a Function ficam saudáveis, o CRD torna-se `Established` e a Composition fica observável na API. Usar recursos Kubernetes de teste. A Etapa 08 valida a reconciliação pelo GitOps.

## Evidência no Hub EKS

Em 28/11/2026, instalamos o Crossplane Core 2.4.3 com Helm e aplicamos Function, XRD e Composition de demonstração pela API. A Function ficou `Healthy` e o CRD de `LabMessage` ficou `Established`. A Etapa 08 registrou o sync do primeiro XR.
