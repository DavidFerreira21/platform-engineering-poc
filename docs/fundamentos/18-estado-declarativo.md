# Estado desejado, estado observado e reconciliação

O estado desejado descreve resultado, como um bucket com política ou um Domain com quota; o estado observado informa o que existe de fato. Um reconciliador compara os dois e tenta convergir de forma idempotente, atualizando condições e expondo erros. Git contém o spec revisado; Kubernetes API contém o recurso e o status. Drift é divergência relevante, não apenas diferença textual de YAML.

## A reconciliação é contínua

Um controller recebe eventos e também reavalia periodicamente recursos sob sua responsabilidade, compara estado atual e declarado e busca convergência. Isso não garante sucesso imediato: dependências podem estar indisponíveis, credenciais podem expirar e políticas podem rejeitar configurações. O contrato precisa expor condições claras de progresso, falha e prontidão. A implementação deve ser idempotente, evitando recriar recursos a cada tentativa.

## Dois caminhos dentro da mesma API

Ao criar ObjectStorage, Crossplane observa o XR e seu recurso AWS até informar Ready. Os registros de Entity, Environment e Domain são declarações lógicas no Git: não possuem loop de reconciliação próprio na primeira versão. Terraform realiza Foundation/EKS por execução autorizada.

Em ObjectStorage, o Kubernetes API Server persiste o recurso, e Crossplane Compositions/Providers criam ou ajustam o bucket AWS. Um registro `environment.yaml` aponta um destino AWS, mas não é CRD nem provisiona conta. O Workflow Plane poderá orquestrar Terraform com gates e auditoria.

## Leitura complementar

- [Kubernetes — Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 3 — declared/actual state e loops de reconciliação.. Referência temática; texto autoral adaptado à POC.

---

[Aula anterior](17-planos-e-responsabilidades.md) · [Índice](README.md) · [Próxima aula](19-abstracoes-e-escape-hatches.md)
