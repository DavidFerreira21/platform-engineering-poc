# Control Plane

O **Kubernetes API Server** é a única API da plataforma. Dois reconciliadores principais:

- **Environment Controller próprio:** observa o CRD nativo `AWSEnvironment`, adota inicialmente contas AWS existentes, atualiza conditions e aciona Argo Workflows/Terraform quando necessário. Criação de novas contas somente na terceira etapa.
- **Crossplane Core:** XRDs independentes de `Domain`, `Database`, `ObjectStorage`, `ComputeInstance` e `Messaging`, com Compositions e Providers. O XRD não cria uma API Server paralela.

O gerenciamento remoto de namespaces por Domain precisa de validação técnica e autorização. Crossplane não pertence ao Workflow Plane. Usar AWS real nos labs autorizados. Ver [ADR-0009](../docs/adr/0009-crossplane-e-reconciliacao.md) e [ADR-0010](../docs/adr/0010-aws-real-e-ambientes.md).
