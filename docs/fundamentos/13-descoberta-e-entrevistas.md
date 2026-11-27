# Discovery: entrevistas e identificação de fricções

A descoberta investiga trabalho real antes de definir soluções. Entreviste usuários sobre sua última tarefa, observe o fluxo, registre tempo de espera e exemplos concretos, agrupe dores recorrentes e valide hipóteses. Evite perguntas indutoras como 'você usaria Backstage?': pergunte o que a pessoa tentou, por que interrompeu e qual resultado precisava. Os achados devem orientar um backlog de problemas.

## Descobrir antes de prescrever

Entrevistas de descoberta devem partir de ocorrências recentes, não de opiniões abstratas sobre ferramentas. Perguntas como 'conte a última vez que precisou criar um banco' revelam filas, contornos manuais e requisitos que um portal não resolveria sozinho. Agrupar observações recorrentes ajuda a distinguir causas raiz e formular hipóteses de produto.

## Da evidência ao backlog

A POC usa cenários fictícios; não há evidência de entrevistas reais ou de ganho de produtividade medido. Um exemplo de hipótese é reduzir a espera por storage com um contrato ObjectStorage. Para validá-la, registrar dependências, políticas, diagnósticos e tempo de espera antes e depois da entrega.

Se vários consumidores relatam o mesmo atraso por provisionamento de storage, a hipótese não precisa ser 'instalar Backstage'. Pode ser 'disponibilizar ObjectStorage por contrato, com políticas automáticas e estado observável'. A implementação só será um sucesso se reduzir essa fricção.

## Leitura complementar

- [DORA — Platform Engineering](https://dora.dev/capabilities/platform-engineering/) — material de referência para aprofundamento.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 2 — platform discovery interviews.

---

[Aula anterior](12-personas-e-jornadas.md) · [Índice](README.md) · [Próxima aula](14-tvp.md)
