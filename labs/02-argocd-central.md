# Lab 02 — Argo CD central no Hub kind

[← Lab 01](01-platform-management-cluster.md) · [Índice dos labs](README.md)

Neste lab, o Helm faz o bootstrap do Argo CD. Depois, o Argo CD lê uma revisão do Git e sincroniza um ConfigMap no Hub. Você vai observar a divergência entre o Git e o cluster e restaurar o valor declarado. O exercício usa apenas recursos locais e tem custo AWS zero.

## Pré-requisitos

- Lab 01 concluído, com `kind`, `kubectl` e `helm` disponíveis.
- Uma revisão publicada neste monorepo que contenha `workflow-plane/argocd/demo/configmap.yaml`.
- Conectividade do Hub ao GitHub e aos registries das imagens.

Execute os comandos da raiz do monorepo.

## 1. Instalar e acessar o Argo CD

```bash
bash platform/hub/scripts/verify.sh
bash platform/hub/scripts/install-argocd.sh
kubectl --context kind-platform-management -n argocd get deployments,statefulsets,pods,svc
```

O [chart fixado](../workflow-plane/argocd/chart-version.txt) instala os CRDs de Application e AppProject. O Service `argocd-server` deve aparecer como `ClusterIP`. Você pode executar o script novamente para reaplicar a configuração.

Deixe o port-forward aberto em outro terminal.

```bash
kubectl --context kind-platform-management -n argocd port-forward \
  --address 127.0.0.1 svc/argocd-server 8081:443
```

Abra `https://localhost:8081`. O servidor usa um certificado local autoassinado. Para o primeiro acesso, use o usuário `admin` e consulte a senha inicial no seu terminal.

```bash
kubectl --context kind-platform-management -n argocd \
  get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 --decode
printf '\n'
```

Após o login, troque a senha em **User Info → Update Password**. Mantenha as credenciais fora do Git e de registros compartilhados.

## 2. Sincronizar uma revisão publicada

O monorepo deste lab é público. O Argo CD lê a URL HTTPS declarada na Application sem credenciais. Comece pelo comando abaixo.

```bash
read -r -p 'SHA completo do commit publicado: ' lab02_sha
bash workflow-plane/argocd/bootstrap/apply-demo.sh "$lab02_sha"
```

O AppProject `lab02` restringe a origem ao monorepo e permite apenas ConfigMaps em `platform-system` no Hub. Na UI, abra a aplicação `lab02-gitops-demo`, selecione **Sync** e confirme em **Synchronize**. Confira `Synced`, `Healthy`, a revisão sincronizada e o valor do ConfigMap.

```bash
kubectl --context kind-platform-management -n argocd get application lab02-gitops-demo \
  -o 'custom-columns=SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
kubectl --context kind-platform-management -n platform-system get configmap lab02-gitops-demo \
  -o 'jsonpath={.data.mensagem}{"\n"}'
```

O valor esperado é `definido-no-git`. Se a aplicação mostrar `Unknown` ou `ComparisonError`, inspecione `kubectl --context kind-platform-management -n argocd describe application lab02-gitops-demo`. Confira a conectividade ao GitHub, o SHA publicado e o caminho. Consulte os logs de `deployment/argocd-repo-server` no namespace `argocd` se a comparação continuar falhando.

## 3. Provocar drift e recuperar

```bash
kubectl --context kind-platform-management -n platform-system patch configmap lab02-gitops-demo \
  --type merge -p '{"data":{"mensagem":"alterado-no-cluster"}}'
```

Use **Refresh** na aplicação e observe `OutOfSync`. Abra **Diff** e compare o valor no cluster com o valor do Git. Execute **Sync** novamente e repita os comandos de verificação. A revisão deve continuar a mesma e o valor deve voltar a `definido-no-git`.

## 4. Limpar o exercício

```bash
kubectl --context kind-platform-management -n argocd delete application lab02-gitops-demo --ignore-not-found
kubectl --context kind-platform-management -n platform-system delete configmap lab02-gitops-demo --ignore-not-found
kubectl --context kind-platform-management -n argocd delete appproject lab02 --ignore-not-found
bash platform/hub/scripts/verify.sh
```

Encerre o port-forward com Ctrl+C. Mantenha o Argo CD instalado para o Lab 02B.

Para desinstalar o Argo CD opcionalmente, primeiro confirme que nenhuma outra aplicação depende dele. Execute `helm --kube-context kind-platform-management -n argocd uninstall argocd`. O chart mantém os CRDs. A exclusão dos CRDs é uma operação separada e remove todas as instâncias desses tipos no cluster.
