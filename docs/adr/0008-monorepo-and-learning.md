# ADR-0008: Estrutura do monorepo e labs incrementais

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
A organização do Git deve refletir os planos e os ambientes, preservando reutilização de código e transparência pedagógica.

## Decisão
Criar diretórios control-plane, workflow-plane, observability-plane, platform/hub e data-plane/uat e data-plane/prod (cada qual com management-eks, eks-pci, eks-non-pci e aws-runtime), além de docs, labs e tests. Hub-and-Spoke fica na documentação como conceito auxiliar, não como diretório obrigatório. Todo conteúdo explicativo será em pt-BR.

## Consequências e pontos de atenção
A divisão de pastas organiza responsabilidades, mas não equivale a isolamento runtime. Os componentes serão implementados por fases com testes e exercícios observáveis.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.
