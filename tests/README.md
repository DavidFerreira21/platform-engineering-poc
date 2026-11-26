# ✅ Testes

Os testes cobrirão contratos Kubernetes, validação de schemas, controllers, workflows, segurança e integrações.

- **Padrão:** testes locais, mocks e verificações estáticas sem recursos AWS reais.
- **Integração AWS:** habilitada apenas quando houver configuração e autorização específicas.
- **Qualidade:** validar links Markdown, manifests e caminhos do repositório ao evoluir os labs.

## Regressão do TerraformStack Controller

A suíte em `tests/controller/` executa o reconciliador Python real contra uma API Kubernetes simulada com cópias dos objetos, conflitos de resourceVersion e relógio determinístico. Não usa credenciais, cluster, Terraform ou AWS.

```bash
python3 -m pip install pytest==8.3.5 -r control-plane/terraformstack/controller/requirements.txt
python3 -m pytest -q tests/controller
```

Cobre resolução da intenção, catálogo, idempotência, retomada, concorrência de status, aprovação, cancelamento, apply ativo, dependências, drift, retry e finalizers. O CI executa a suíte em cada PR. Os testes não substituem o Simulator do Lab 06: a atomicidade do claim do runner, a persistência real da API e a execução OPA/Terraform precisam de validação de integração.

---

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro prático](../docs/learning-roadmap.md)
