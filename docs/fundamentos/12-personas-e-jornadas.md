# Personas e jornada do desenvolvedor

Persona é representação de necessidades, responsabilidades e dificuldades de um grupo de consumidores, validada por evidências. Jornada mapeia cada passo de uma tarefa, esperas, erros, ferramentas e emoções. Um único fluxo pode ter consumidores distintos: desenvolvedor quer criar banco; SRE quer rastrear falhas; segurança quer evidência de política. A plataforma precisa explicitar essas diferenças.

## Persona não é estereótipo

Uma boa persona deriva de atividades e obstáculos observados. Um desenvolvedor de aplicação quer criar uma dependência com baixo tempo de espera; um SRE quer investigar falhas e indicadores; um responsável por segurança quer evidências de política e auditoria. Cada um percebe a mesma plataforma de forma diferente. A jornada descreve etapas reais, ferramentas utilizadas, decisões, atrasos e caminhos de recuperação.

## Mapeamento de uma jornada

Na jornada-alvo de Database, a squad identifica a necessidade; o DevOps Federado escolhe um perfil e propõe a intenção para UAT; o Platform Team mantém contrato, políticas e motor. A squad consome a configuração autorizada do banco. Cada persona precisa de feedback adequado, sem receber credenciais administrativas do motor. [Personas e ownership](../adr/0007-personas-and-developer-experience.md).

Para Database, observe desde a descoberta do contrato e escolha do ambiente até o resultado Ready. Registre também o que acontece quando a Composition falha ou o usuário não possui autorização. O caminho de erro é parte do produto e merece tanta atenção quanto o caminho de sucesso.

## Leitura complementar

- [DORA — Platform Engineering](https://dora.dev/capabilities/platform-engineering/) — material de referência para aprofundamento.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 2 — personas e developer workflows.

---

[Aula anterior](11-modos-de-interacao.md) · [Índice](README.md) · [Próxima aula](13-descoberta-e-entrevistas.md)
