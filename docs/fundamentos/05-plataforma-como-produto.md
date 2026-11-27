# Plataforma como produto interno

Um produto de plataforma tem clientes internos, proposta de valor, responsabilidades, documentação, suporte, indicadores e ciclo de evolução. Isso é diferente de uma coleção de ferramentas instaladas. O time deve pesquisar problemas dos consumidores, selecionar capacidades, acompanhar adoção e melhorar a experiência. A adoção genuína é evidência de valor; obrigar o uso de um portal não prova que ele é útil.

## Contrato, operação e evolução

Produto interno não equivale a um item no catálogo. A plataforma precisa definir quem são seus clientes, o que a capacidade entrega, quais requisitos aceita, como reporta erro, como evolui sem quebrar consumidores e qual suporte oferece. A plataforma deve tratar mudanças de API, deprecations, observabilidade, documentação e incidentes como responsabilidades contínuas.

## Valor e adoção

ObjectStorage será um produto quando possuir contrato estável, segurança padrão, tempos esperados, documentação, acompanhamento e lifecycle, não apenas quando existir um bucket AWS. O TerraformStack já exemplifica um produto interno para o Platform Team: API, documentação, testes e sinais operacionais. Ter um contrato implementado não comprova adoção ou satisfação.

A adoção compulsória pode inflar métricas e esconder frustração. Avalie tempo de provisionamento, taxa de sucesso, satisfação e número de exceções. Uma Capability que elimina uma dependência concreta tem mais valor do que instalar quatro ferramentas sem nenhum consumidor. O backlog deve priorizar problemas observados, não entusiasmo com tecnologia.

## Leitura complementar

- [DORA — Platform Engineering](https://dora.dev/capabilities/platform-engineering/) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — platform as internal product e product thinking.

---

[Índice](README.md) · [Aula anterior](04-friccao-e-sintomas.md) · [Próxima aula](06-plataforma-como-servico.md)
