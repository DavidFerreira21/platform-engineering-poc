# API-first, CRDs, XRDs, Operators e ADRs

API-first significa definir o contrato do consumidor antes de selecionar ferramentas. No Kubernetes, CRDs estendem a API; XRDs do Crossplane definem recursos compostos e sua reconciliação; Compositions e Providers materializam serviços. Controllers/Operators observam CRs e executam lógica. ADR registra contexto, decisão e consequências para impedir que escolhas virem regras sem justificativa. Não confunda uma API pública com um controller de execução.

## Quem define o quê

A API Kubernetes expõe Custom Resources, e um CRD descreve a forma de um tipo. O XRD do Crossplane permite definir um recurso composto, exposto pela mesma Kubernetes API; Composition especifica como ele vira recursos concretos e Providers se comunicam com sistemas como a AWS. Controllers próprios são programas que observam recursos e executam lógica. Argo Workflows é diferente: coordena tarefas sequenciais, pausas e aprovações, não substitui a reconciliação de todos os recursos.

## Exemplos do nosso desenho

AWSEnvironment é CRD + Environment Controller próprio; Domain e Capabilities são XRD + Crossplane. Todos são consumidos pelo mesmo Kubernetes API Server, mas executados por reconciliadores diferentes.

AWSEnvironment será CRD nativo com Environment Controller que coordena Argo Workflows/Terraform. Domain e Database serão XRDs especializados reconciliados pelo Crossplane. Um ADR registra essa divisão com contexto e consequências. O mesmo consumidor consegue consultar os recursos pelo Kubernetes API Server, independentemente do controller responsável.

## Leitura complementar

- [Kubernetes — Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) — documentação e material público complementar.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 3 — API-first, operator pattern e ADRs; aplicação adaptada à POC.. Referência temática; texto autoral adaptado à POC.

---

[Aula anterior](20-multitenancy-cells.md) · [Índice](README.md) · [Próxima aula](22-idp-e-developer-portal.md)
