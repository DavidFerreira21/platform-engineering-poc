# Carga cognitiva e troca de contexto

Carga cognitiva é o esforço mental para lidar com informações e tarefas. É útil distinguir dificuldade intrínseca ao problema, esforço extrínseco produzido por ferramentas/processos ruins e esforço dedicado a construir modelos mentais. O objetivo de uma plataforma não é esconder tudo, mas reduzir esforço extrínseco: credenciais manuais, YAML repetido, múltiplos fluxos para o mesmo resultado e interrupções por dependências. Cada mudança de contexto consome atenção e amplia risco de erro.

## Três tipos de esforço

A complexidade intrínseca vem da tarefa em si: uma aplicação crítica realmente precisa de escolhas de resiliência e consistência. A extrínseca vem de detalhes periféricos, como decorar inúmeros parâmetros IAM ou compreender a topologia de VPC para criar um bucket. A carga construtiva é o esforço investido para criar entendimento duradouro. A plataforma reduz principalmente o esforço extrínseco e ensina o necessário, mas não deve esconder decisões que alteram risco, custo ou disponibilidade.

## Como reconhecer a melhoria

Em uma futura Capability Database, o consumidor descreve tamanho, disponibilidade e ambiente em vez de aprender configuração de VPC, IAM e Provider. Para o Platform Team, o TerraformStack reduz a necessidade de lembrar a sequência de execução, ao expor intenção, status e dependências. Cada persona recebe uma abstração diferente; esconder um risco operacional relevante não é reduzir carga cognitiva.

Não basta diminuir a quantidade de campos em um formulário. Uma boa abstração diminui dúvidas, caminhos alternativos, dependência de especialistas e retrabalho. O indicador é conseguir executar uma tarefa frequente com menos interrupções, mantendo qualidade e capacidade de diagnóstico.

## Leitura complementar

- [DORA — Platform Engineering](https://dora.dev/capabilities/platform-engineering/) — referência pública para aprofundar o conceito.
- *Platform Engineering in Practice*, Zebb Wright (2026), Capítulo 1 — carga cognitiva e context switching.

---

[Índice](README.md) · [Aula anterior](02-devops-e-platform-engineering.md) · [Próxima aula](04-friccao-e-sintomas.md)
