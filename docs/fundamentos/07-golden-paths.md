# Golden Paths e Paved Roads

Golden Path é o percurso recomendado e suportado para realizar uma tarefa recorrente, com escolhas seguras e documentação. Deve ser mais simples que reinventar a integração. Não significa caminho único e obrigatório: exceções justificadas precisam de escape hatches. A maturidade vem de reduzir decisões acidentais e oferecer resultados previsíveis, como criar um domínio, solicitar um banco ou publicar um serviço.

## Um Golden Path é uma experiência completa

Um caminho recomendado conecta descoberta, escolha do produto, geração/validação do contrato, provisionamento, feedback e operação. Ele é opinativo: oferece defaults seguros e dispensa decisões que não deveriam recair sobre o consumidor. Mas não deve ser um único formulário inflexível. Casos fora do padrão podem exigir evolução do produto ou uma saída governada (*escape hatch*).

## Um exemplo concreto

Golden Path de storage: selecionar ambiente e Domain, escolher perfil aprovado, gerar CR, abrir PR, observar reconciliação e obter endpoint/configuração permitida.

Na POC, um Golden Path para ObjectStorage começa na necessidade de armazenar objetos, define perfil e ambiente, produz um Custom Resource revisado em Git, acompanha Ready/NotReady e mostra como encontrar evidências de erro e remover com segurança. Backstage pode facilitar o percurso mais tarde; o Golden Path já existe como contrato, documentação e automação antes de um portal.

## Leitura complementar

- [Platform Engineering — Fundamentos](https://university.platformengineering.org/course-1) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — golden paths e paved roads.. Referência temática; esta explicação é autoral, não uma tradução do livro.

---

[Índice](README.md) · [Aula anterior](06-plataforma-como-servico.md) · [Próxima aula](08-guardrails-e-gates.md)
