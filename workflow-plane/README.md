# Workflow Plane

Contém **Argo Workflows** (etapas, esperas, aprovações), **Terraform Executor** (plan/apply controlados) e **Argo CD** (GitOps). É uma camada lógica hospedada no Platform Management Cluster, não um cluster separado.

O Environment Controller, localizado no Control Plane, pode acionar Argo Workflows. **Crossplane Core, XRDs, Compositions e Providers pertencem ao Control Plane**, pois reconciliam recursos declarativos sem depender de workflows sequenciais. Recursos AWS reais são provisionados somente em contas autenticadas, mediante plano e autorização.
