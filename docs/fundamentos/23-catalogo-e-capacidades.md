# Catálogo de serviços, produtos e capacidades da plataforma

Uma plataforma geralmente oferece um **catálogo de capacidades**: opções que consumidores podem descobrir, entender e solicitar, como bancos de dados, namespaces, object storage, mensageria, observabilidade ou pipelines. Catálogo é o mecanismo de descoberta e organização; capacidade é aquilo que a plataforma entrega. Um serviço de plataforma é a oferta operada e suportada, com contrato e lifecycle próprios. Os termos são próximos, mas não são intercambiáveis.

Um catálogo útil precisa apresentar o resultado de cada opção, público atendido, pré-requisitos, limites, custos/risco e quem mantém o produto. Sem esses elementos, a interface vira apenas um índice de ferramentas. A curadoria também importa: muitas variantes irrelevantes aumentam carga cognitiva, enquanto uma lista pequena demais força exceções constantes.

## Famílias versus produtos especializados

Na POC, **Capability é uma família arquitetural**, não um CRD genérico. Os produtos concretos terão XRDs especializados, como `Database`, `ObjectStorage`, `ComputeInstance` e `Messaging`. Isso oferece contratos tipados e permite políticas adequadas para cada serviço. `Domain` é a unidade lógica de negócio, ownership e consumo de Capabilities e aplicações; `Environment` é o registro declarativo do destino UAT/PROD associado a uma conta AWS. Ambos são conceitos representados inicialmente por configuração Git, **não XRDs**.

Um catálogo poderá apontar para essas opções e suas versões. O portal não define a semântica do recurso nem executa a reconciliação; o API Server e seus controllers permanecem responsáveis por isso. Um produto pode ter múltiplos perfis suportados sem exigir um XRD novo para cada tamanho de banco.

## Aplicação na POC

Antes de implementar Backstage, escrevemos a documentação de `ObjectStorage` com perfil aceito, permissões, comportamento de exclusão, resultado esperado e estados de erro. Posteriormente o catálogo pode mostrar essa informação e criar o manifesto por meio de um Golden Path.

## Leitura complementar

- [CNCF — Platform Engineering Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/)
- [Backstage — Software Catalog](https://backstage.io/docs/features/software-catalog/)

---

[Aula anterior](22-idp-e-developer-portal.md) · [Índice](README.md) · [Próxima aula](24-ciclo-de-vida-produtos.md)
