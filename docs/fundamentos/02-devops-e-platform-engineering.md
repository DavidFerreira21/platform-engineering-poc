# DevOps não é a mesma coisa que Platform Engineering

DevOps é um conjunto de práticas e uma cultura que aproxima desenvolvimento e operações, enfatizando feedback, automação e responsabilidade compartilhada. Platform Engineering estrutura produtos internos que tornam essas práticas sustentáveis em escala. Não é uma etapa que substitui DevOps: é uma forma de habilitar times de produto a trabalhar com autonomia sem que cada equipe recrie a mesma fundação.

## Fronteiras de responsabilidade

O princípio DevOps de responsabilidade por entrega e operação continua válido. A plataforma remove tarefas transversais repetidas e oferece um caminho sustentado para as equipes executarem seu trabalho. O time de aplicação ainda é responsável pela lógica de negócio, SLOs do produto e configuração dentro do contrato; o time de plataforma responde pelo funcionamento, segurança e evolução das capacidades compartilhadas. A interface entre ambos precisa explicitar esse acordo.

## Um teste útil

GitOps e automação pertencem à nossa implementação, mas não são suficientes: o diferencial é o contrato de Domain, Database ou ObjectStorage com experiência previsível.

Se cada time precisar ajustar manifests, redes e credenciais em dezenas de passos exclusivos, apenas mudamos o endereço da complexidade. Uma plataforma bem desenhada permite consumir o produto sem depender do operador da ferramenta a cada mudança.

## Leitura complementar

- [CNCF — Platforms White Paper](https://tag-app-delivery.cncf.io/whitepapers/platforms/) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — evolução DevOps e equipes de plataforma.. Referência temática; esta explicação é autoral, não uma tradução do livro.

---

[Índice](README.md) · [Aula anterior](01-por-que-platform-engineering.md) · [Próxima aula](03-carga-cognitiva.md)
