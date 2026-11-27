# Contratos de execução e integrações desacopladas

Uma interface entre componentes deve especificar significado, não apenas formato. Um contrato de execução descreve a intenção recebida, os identificadores que correlacionam a tentativa, os resultados aceitos e as garantias de segurança. O adapter traduz o transporte externo para o modelo interno. Dessa forma, mudar uma ferramenta não exige modificar todas as decisões do núcleo.

## Flexibilidade com limites

Um contrato que exige nomes de steps, scanners ou regras de uma ferramenta específica transforma a implementação de referência em dependência estrutural. Um contrato que aceita qualquer resposta sem validação perde previsibilidade. O equilíbrio é definir sinais mínimos e validar combinações coerentes: identidade, progresso, resolução da fonte e do destino, resultado e autorização quando habilitada.

Flexibilidade operacional também não significa liberdade para escolher qualquer identidade privilegiada. Integrações precisam de configuração confiável, allowlist e políticas de acesso. Snapshots preservam a interpretação de tentativas ativas quando a instalação evolui.

## Aplicação na POC

O TerraformStack usa a interface interna de execução e um adapter Argo. O Platform Team cadastra aliases para WorkflowTemplates, com default, namespace, ServiceAccount e mapeamento de parâmetros. O CR escolhe um alias autorizado. O handshake de autorização tem default true, mas pode ser desligado por integração compatível. OPA, aprovação humana e ferramenta de IaC continuam decisões da pipeline.

Não há categorias obrigatórias de workflows nem um segundo tipo de adapter implementado. Com autorização desligada, o resultado Verified ainda exige identidade, resolução válida e término nativo bem-sucedido; o controller não infere a fase de apply. O template de referência conserva o handshake.

## Leitura complementar

- [Execution Contract](../flows/terraformstack-execution-contract.md)
- [Configuração e integração](../../control-plane/terraformstack/controller/README.md)

---

[Aula anterior](26-reconciliacao-e-orquestracao.md) · [Índice](README.md) · [Próxima aula](28-idempotencia-e-concorrencia.md)
