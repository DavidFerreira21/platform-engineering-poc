# Etapa 08 — Crossplane de ponta a ponta

[← Etapa 07](07-crossplane-core.md) · [Guias das etapas](README.md)

## Objetivo

Comprovar o ciclo Git → Argo CD → Kubernetes API → Crossplane no Hub EKS, da criação de um `LabMessage` à remoção do ConfigMap composto.

## O que será feito

- Sincronizar a API de demonstração e sua primeira solicitação em Applications com sync manual.
- Confirmar o schema, o vínculo com a Composition, as condições do XR e o conteúdo do ConfigMap.
- Atualizar a mensagem via Git, repetir o sync e observar o Crossplane corrigir um desvio no ConfigMap.
- Verificar os limites dos AppProjects e concluir o ciclo com prune do XR.

## Por que fizemos assim

`crossplane-core` instala o motor compartilhado. `crossplane-demo-api` instala Function, XRD e Composition. `crossplane-demo` entrega apenas o `LabMessage`, depois de a API estar disponível. Os dois últimos reutilizam os manifests da Etapa 07. O Argo CD entrega a intenção e o Crossplane gerencia o ConfigMap composto. O ensaio usa recursos Kubernetes de teste e é independente do Simulator. [Catálogo do Crossplane](../platform/hub/addons/catalog/crossplane-core.yaml) · [ADR-0003](../docs/adr/0003-reconciliation-and-gitops.md)

> O sync do XR e a correção de drift foram exercitados no Hub EKS. A entrega da API pelo catálogo e o ciclo de commits de atualização têm aceite próprio.

## Aceite no Hub EKS

Confirmar que o Core e a API de demonstração estão prontos. Conferir revisão, destino por ARN, AppProjects e diff de cada Application. Entregar o XR pelo Argo CD e observar sua primeira reconciliação.

## O que observar

Conferir `Synced` e `Ready`, a rejeição de um XR inválido e os limites do AppProject. Atualizar o XR, repetir o sync e verificar que o Crossplane corrige a alteração externa do ConfigMap. Remover a solicitação via GitOps e verificar a limpeza dos recursos do exercício.

## Evidências de aceite

### Aceite do XR no Hub EKS

Em 28/11/2026, o Argo CD gerenciado sincronizou a Application `crossplane-demo` a partir de `main`, com operação `Succeeded`. O `LabMessage` ficou `Ready`, e o Crossplane criou o ConfigMap e restaurou seu conteúdo após uma alteração externa. O dry-run rejeitou mensagem vazia. A exclusão do XR pela API removeu o ConfigMap composto, e um novo sync recriou o XR em `Ready`. A remoção por prune no Argo CD segue pendente.

### Ensaio inicial

Em 27/11/2026, sincronizamos o Core 2.4.3 no cluster do ensaio inicial. A Function `v0.8.2` ficou saudável e o CRD foi estabelecido. Com os AppProjects da etapa e manifests locais fornecidos ao Argo CD, a criação e a atualização do XR passaram. O ConfigMap manteve o UID no re-sync e o Crossplane corrigiu o desvio sem outro sync. O dry-run recusou a mensagem vazia e o AppProject recusou o ConfigMap proposto como recurso da Application.

O Argo CD também sincronizou o XR a partir do commit `1b48874` de `main`. Para exercitar o prune com essa mesma revisão, excluímos temporariamente `xr.yaml` pelo filtro da Application. O XR e o ConfigMap desapareceram e o Core permaneceu `Synced` e `Healthy`. O roteiro usa commits publicados para a atualização e a remoção. A descoberta automática das duas entradas do catálogo e a sequência completa de commits ficam para a execução com a revisão publicada da Etapa 08.
