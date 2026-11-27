# API-first, CRDs, XRDs, Operators e ADRs

API-first significa definir o contrato do consumidor antes de selecionar ferramentas. No Kubernetes, CRDs estendem a API; XRDs do Crossplane definem recursos compostos e sua reconciliação; Compositions e Providers materializam serviços. Controllers/Operators observam CRs e executam lógica. ADR registra contexto, decisão e consequências para impedir que escolhas virem regras sem justificativa. Não confunda uma API pública com um controller de execução.

## Quem define o quê

A API Kubernetes expõe Custom Resources, e um CRD descreve a forma de um tipo. O XRD do Crossplane permite definir um recurso composto, exposto pela mesma Kubernetes API; Composition especifica como ele vira recursos concretos e Providers se comunicam com sistemas como a AWS. Controllers próprios são programas que observam recursos e executam lógica. Argo Workflows é diferente: coordena tarefas sequenciais, pausas e aprovações, não substitui a reconciliação de todos os recursos.

## Exemplos do nosso desenho

Entity, Environment e Domain são registros lógicos do Git. As Infrastructure Capabilities são XRDs especializados do Crossplane; TerraformStack usa um CRD próprio, independente dos XRDs.

`environment.yaml` registra conta/região; `domain.yaml` registra ownership. A futura API ObjectStorage será definida por XRD e Composition; o consumidor cria um XR, não o XRD. Na POC, o TerraformStack é um CRD com controller próprio que coordena Workflows via adapter Argo. Nem todo conceito precisa virar CRD. ADRs registram decisões, e testes e evidências registram o que de fato foi validado.

## Leitura complementar

- [Kubernetes — Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 3 — API-first, operator pattern e ADRs.

---

[Aula anterior](20-multitenancy-cells.md) · [Índice](README.md) · [Próxima aula](22-idp-e-developer-portal.md)
