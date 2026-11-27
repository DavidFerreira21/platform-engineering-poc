# Configuração, state e vínculo de destino

Separar configuração de implementação permite reutilizar uma stack em diversos ambientes. O código define os recursos e invariantes; os valores definem opções autorizadas; o state da ferramenta registra a associação entre objetos declarados e recursos externos. Misturar essas responsabilidades torna alterações e recuperação perigosas.

## Herança e destino

Uma hierarquia de configuração pode aplicar defaults globais, valores da Entity e overrides do Environment. A regra de merge precisa ser explícita: objetos podem preservar chaves não sobrescritas, enquanto listas podem ser substituídas integralmente. O resultado precisa ser validado antes de qualquer alteração.

O destino é outra dimensão. Identidade AWS, conta, região, backend e chave de state não devem ser escolhas arbitrárias de quem preenche o formulário. Separar states por escopo reduz impacto de mudanças e permite locking por unidade. State e planos podem conter dados sensíveis; não são artefatos para publicar no Git ou logs.

## Aplicação na POC

O módulo Configuration compõe YAMLs globais, por Entity e Environment. Os registros environment.yaml descrevem o destino; as flags efetivas das stacks Network e EKS vêm de seus próprios YAMLs. O Simulator fixa um commit e um vínculo opaco de destino (binding). O controller conserva esse vínculo; a pipeline valida seu significado, inclusive quando outra integração é selecionada.

A troca de ferramenta, backend ou destino pode exigir migração explícita. Não se deve descartar state para resolver um erro de binding. O Simulator usa PVC por stack; o backend S3 com chaves segregadas será validado na AWS. A identidade do destino e a integridade dos planos continuam responsabilidades do executor.

## Leitura complementar

- [Foundation e Configuration](../../foundation/README.md)
- [Execution Contract](../flows/terraformstack-execution-contract.md)

---

[Aula anterior](29-drift-e-convergencia.md) · [Índice](README.md) · [Próxima aula](31-identidade-e-fronteiras-de-confianca.md)
