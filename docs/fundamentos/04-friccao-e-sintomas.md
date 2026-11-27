# Sintomas de uma plataforma ausente ou fraca

Filas de tickets, ambientes especiais que ninguém reproduz, pipelines copiados, permissões solicitadas por mensagens e divergência entre ambientes são sintomas de fricção. Não é necessário construir um portal para começar: primeiro mensure tempos de espera, taxa de retrabalho, tarefas repetidas e incidentes ligados ao provisionamento. Procure a capacidade comum subjacente e valide se padronização seria útil.

## Sintomas não são soluções

Um backlog de solicitações de banco pode ter causas distintas: capacidade operacional insuficiente, política insegura, interface confusa ou ausência de catálogo. Comprar um portal não resolve automaticamente nenhuma delas. Primeiro descreva a jornada, identifique os pontos de espera, diferencie trabalho manual inevitável de repetição automatizável e escolha uma capacidade estreita para melhorar.

## Padronização versus exceção

Uma solicitação de bucket que exige várias aprovações manuais não é resolvida apenas instalando Backstage. Precisamos de contrato, política de acesso, criação e status auditável. Na POC, o Simulator demonstra como uma solicitação GitOps pode indicar espera, bloqueio, falha e sucesso sem consultar manualmente cada Pod. Isso é evidência técnica de uma solução; medir a redução real de fricção exigiria observar usuários.

Ambientes snowflake dependem de ajustes não declarados e conhecimentos individuais. Padronizar reduz surpresa e habilita recuperação, mas uma plataforma saudável prevê exceções explícitas com contexto e responsabilidade. O objetivo não é proibir diferença: é tornar diferenças intencionais, rastreáveis e operáveis.

## Leitura complementar

- [CNCF — Maturity Model](https://tag-app-delivery.cncf.io/whitepapers/platform-eng-maturity-model/) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — filas, snowflakes e proliferação de ferramentas.

---

[Índice](README.md) · [Aula anterior](03-carga-cognitiva.md) · [Próxima aula](05-plataforma-como-produto.md)
