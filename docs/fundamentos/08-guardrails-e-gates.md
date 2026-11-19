# Guardrails, políticas e aprovações

Guardrail é controle integrado ao fluxo que previne ou detecta violações de maneira previsível: schema, RBAC, quotas, políticas e testes. Gate é uma decisão ou pausa explícita; pode ser importante para operações caras, sensíveis ou dependentes de equipes externas, mas não deve ser aplicado indiscriminadamente. Bons guardrails permitem avançar de forma segura na maior parte dos casos.

## Controle automático e decisão humana

Guardrails podem operar antes da criação (validação de schema, políticas de admissão), durante a reconciliação (regras de IAM, tags e criptografia) e depois (detecção de desvio). Gates interrompem um fluxo até que uma pessoa ou sistema autorizado decida. Confundir os dois produz uma plataforma lenta: se um padrão pode ser validado automaticamente, não há motivo para abrir uma fila humana para cada solicitação.

## Critério de desenho

Um XRD Database só aceita classes aprovadas e exige tags. Já um workflow de criação de conta AWS, caso seja adotado futuramente, pode exigir gate humano auditado no Argo Workflows.

Os fluxos de ObjectStorage com configurações aprovadas devem seguir automaticamente. Uma criação de conta com rota dependente de outra equipe pode demandar uma etapa de aprovação e validação de resultado, com identidade e trilha de auditoria.

## Leitura complementar

- [CNCF — Platforms White Paper](https://tag-app-delivery.cncf.io/whitepapers/platforms/) — material de referência para aprofundamento.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — guardrails versus gates.. Referência temática; explicação e exemplos autorais.

---

[Aula anterior](07-golden-paths.md) · [Índice](README.md) · [Próxima aula](09-construir-comprar-adotar.md)
