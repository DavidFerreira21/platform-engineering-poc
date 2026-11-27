package controller

import (
	"context"
	"encoding/json"
	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution/argo"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const name = "simulator-network"

var fixed = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func fixture(t *testing.T) (*Reconciler, *unstructured.Unstructured) {
	t.Helper()
	o := stackObject()
	o.SetName(name)
	o.SetNamespace("platform-system")
	o.SetUID(types.UID("test-uid"))
	o.SetGeneration(1)
	o.SetAnnotations(map[string]string{revision: strings.Repeat("a", 40)})
	o.Object["spec"] = M{"sourceRef": M{"repository": "https://example.org/stacks.git", "path": "foundation/stacks/simulator"}, "context": M{"arbitrary-key": "arbitrary-value"}}
	cl := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithStatusSubresource(stackObject(), argo.WorkflowObject()).WithObjects(o).Build()
	config, e := LoadConfig("../../controller-config.yaml")
	if e != nil {
		t.Fatal(e)
	}
	r := &Reconciler{Engine: &argo.Adapter{Client: cl, Reader: cl, Config: config.Workflow}, Client: cl, Reader: cl, StackNamespace: "platform-system", Now: func() time.Time { return fixed }}
	return r, o
}
func getStack(t *testing.T, r *Reconciler) *unstructured.Unstructured {
	t.Helper()
	o := stackObject()
	if e := r.Get(context.Background(), client.ObjectKey{Namespace: r.StackNamespace, Name: name}, o); e != nil {
		t.Fatal(e)
	}
	return o
}
func getRun(t *testing.T, r *Reconciler) M {
	return obj(obj(getStack(t, r).Object["status"])["execution"])
}
func tick(t *testing.T, r *Reconciler, n int) ctrl.Result {
	t.Helper()
	var result ctrl.Result
	for i := 0; i < n; i++ {
		var e error
		result, e = r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKey{Namespace: r.StackNamespace, Name: name}})
		if e != nil {
			t.Fatal(e)
		}
	}
	return result
}
func current(t *testing.T, r *Reconciler) *unstructured.Unstructured {
	t.Helper()
	w, e := r.Engine.(*argo.Adapter).Raw(context.Background(), str(getRun(t, r)["workflow"]))
	if e != nil || w == nil {
		t.Fatalf("workflow: %v", e)
	}
	return w
}
func workflowStatus(t *testing.T, r *Reconciler, phase, progress, outcome, token string) {
	t.Helper()
	w := current(t, r)
	run := getRun(t, r)
	names := r.Engine.(*argo.Adapter).Config.Parameters.Output
	p := []any{}
	add := func(name string, fields M) {
		fields["workflow"], fields["intentHash"] = run["workflow"], run["intentHash"]
		data, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		p = append(p, M{"name": name, "value": string(data)})
	}
	add(names.ResolvedSource, M{"reference": "immutable-rev", "binding": "destination-01"})
	if progress != "" {
		add(names.State, M{"phase": progress})
	}
	if outcome != "" {
		add(names.Result, M{"outcome": outcome, "reason": "FixtureDiagnostic"})
	}
	if token != "" {
		add(names.Authorization, M{"token": token})
	}
	w.Object["status"] = M{"phase": phase, "outputs": M{"parameters": p}}
	if err := r.Status().Update(context.Background(), w); err != nil {
		t.Fatal(err)
	}
}

func grantAndClaim(t *testing.T, r *Reconciler) {
	t.Helper()
	workflowStatus(t, r, "Running", "", "", "hash")
	tick(t, r, 1)
	o := getStack(t, r)
	status := obj(o.Object["status"])
	run := obj(status["execution"])
	if run["authorization"] != "hash" {
		t.Fatal(run)
	}
	status["claim"] = M{"workflow": run["workflow"], "intentHash": run["intentHash"], "authorization": "hash"}
	if err := r.Status().Update(context.Background(), o); err != nil {
		t.Fatal(err)
	}
}
func change(t *testing.T, r *Reconciler, k string, v any) {
	t.Helper()
	o := getStack(t, r)
	obj(o.Object["spec"])[k] = v
	o.SetGeneration(o.GetGeneration() + 1)
	if e := r.Update(context.Background(), o); e != nil {
		t.Fatal(e)
	}
}
func reason(t *testing.T, r *Reconciler) string {
	a := arr(obj(getStack(t, r).Object["status"])["conditions"])
	if len(a) == 0 {
		return ""
	}
	return str(obj(a[0])["reason"])
}
func TestIntent(t *testing.T) {
	_, o := fixture(t)
	d, err := Intent(o)
	if err != nil || obj(d["context"])["arbitrary-key"] != "arbitrary-value" ||
		obj(d["sourceRef"])["version"] != "latest" || d["revisionAnnotation"] != strings.Repeat("a", 40) {
		t.Fatalf("%v %v", d, err)
	}
	obj(o.Object["spec"])["context"] = M{"region-free": "updated"}
	updated, err := Intent(o)
	if err != nil || execution(d, nil, "GitOps", 0)["intentHash"] == execution(updated, nil, "GitOps", 0)["intentHash"] {
		t.Fatal("mudança de contexto não alterou a identidade")
	}
	for _, path := range []string{"../x", "a/../x", "a/./x", "a//x", "a/", "/a", "./a"} {
		obj(obj(o.Object["spec"])["sourceRef"])["path"] = path
		if _, err := Intent(o); err == nil {
			t.Fatalf("caminho inválido aceito: %s", path)
		}
	}
}

func TestIdempotency(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 2)
	run := getRun(t, r)
	if num(run["sequence"]) != 1 {
		t.Fatal(run)
	}
	w, e := r.Engine.(*argo.Adapter).Raw(context.Background(), str(run["workflow"]))
	if e != nil || w != nil {
		t.Fatal("workflow antes de persistir intenção")
	}
	tick(t, r, 5)
	if getRun(t, r)["workflow"] != run["workflow"] || !yes(getRun(t, r)["created"]) {
		t.Fatal("duplicação")
	}
}
func TestAuthorization(t *testing.T) {
	for _, test := range []struct {
		name, phase, outcome, token string
		allow                       bool
	}{
		{"request", "Running", "", "hash", true},
		{"waiting", "Waiting", "", "", false},
		{"external gate completed", "Waiting", "", "hash", true},
		{"contradictory", "Running", "Blocked", "hash", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, _ := fixture(t)
			tick(t, r, 4)
			workflowStatus(t, r, "Running", test.phase, test.outcome, test.token)
			tick(t, r, 1)
			if (getRun(t, r)["authorization"] == "hash") != test.allow {
				t.Fatal(getRun(t, r))
			}
		})
	}
}
func TestCancelBeforeApply(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	change(t, r, "context", M{"arbitrary-key": "changed-value"})
	tick(t, r, 2)
	if obj(current(t, r).Object["spec"])["shutdown"] != "Terminate" {
		t.Fatal("não cancelou")
	}
	workflowStatus(t, r, "Failed", "", "", "")
	tick(t, r, 3)
	if num(getRun(t, r)["sequence"]) != 2 {
		t.Fatal(getRun(t, r))
	}
}
func TestActiveApplyPreserved(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	grantAndClaim(t, r)
	change(t, r, "context", M{"arbitrary-key": "changed-value"})
	tick(t, r, 1)
	if obj(current(t, r).Object["spec"])["shutdown"] != nil {
		t.Fatal("apply interrompido")
	}
	workflowStatus(t, r, "Succeeded", "Running", "Verified", "hash")
	tick(t, r, 2)
	if num(getRun(t, r)["sequence"]) != 2 {
		t.Fatal(getRun(t, r))
	}
}
func TestMissingWorkflow(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	if e := r.Delete(context.Background(), current(t, r)); e != nil {
		t.Fatal(e)
	}
	tick(t, r, 2)
	if reason(t, r) != "ExecutionMissing" {
		t.Fatal(reason(t, r))
	}
}
func TestDriftTimer(t *testing.T) {
	r, _ := fixture(t)
	change(t, r, "reconciliation", M{"driftCheckInterval": "45s"})
	tick(t, r, 4)
	workflowStatus(t, r, "Succeeded", "Running", "NoChanges", "")
	tick(t, r, 1)
	result := tick(t, r, 1)
	if result.RequeueAfter != 45*time.Second {
		t.Fatal(result)
	}
	r.Now = func() time.Time { return fixed.Add(45 * time.Second) }
	tick(t, r, 2)
	if getRun(t, r)["trigger"] != "DriftCheck" || getRun(t, r)["pinnedReference"] != "immutable-rev" {
		t.Fatal(getRun(t, r))
	}
}
func TestRetry(t *testing.T) {
	r, _ := fixture(t)
	change(t, r, "reconciliation", M{"retryPolicy": M{"strategy": "Fixed", "interval": "10s", "maxRetries": int64(1)}})
	tick(t, r, 4)
	workflowStatus(t, r, "Failed", "Running", "", "")
	tick(t, r, 1)
	if tick(t, r, 1).RequeueAfter != 10*time.Second {
		t.Fatal("timer")
	}
	r.Now = func() time.Time { return fixed.Add(10 * time.Second) }
	tick(t, r, 2)
	if num(getRun(t, r)["attempt"]) != 1 || getRun(t, r)["pinnedReference"] != "immutable-rev" {
		t.Fatal(getRun(t, r))
	}
	workflowStatus(t, r, "Failed", "Running", "", "")
	tick(t, r, 1)
	r.Now = func() time.Time { return fixed.Add(time.Hour) }
	tick(t, r, 2)
	if num(getRun(t, r)["sequence"]) != 2 {
		t.Fatal("retry ilimitado")
	}
}
func TestSuspend(t *testing.T) {
	r, _ := fixture(t)
	change(t, r, "suspend", true)
	tick(t, r, 4)
	if reason(t, r) != "Suspended" || len(getRun(t, r)) != 0 {
		t.Fatal("suspensão")
	}
}
func TestDependencies(t *testing.T) {
	_, o := fixture(t)
	obj(o.Object["spec"])["dependsOn"] = []any{M{"name": "dependent"}}
	spec, err := specOf(o)
	if err != nil {
		t.Fatal(err)
	}
	if dependencyReason(nil, o, spec, "apply") != "DependencyBlocked" {
		t.Fatal("dependência ausente aceita")
	}
	other := o.DeepCopy()
	other.SetName("dependent")
	obj(other.Object["spec"])["dependsOn"] = []any{M{"name": o.GetName()}}
	if dependencyReason([]unstructured.Unstructured{*o, *other}, o, spec, "apply") != "DependencyCycle" {
		t.Fatal("ciclo aceito")
	}
	obj(other.Object["spec"])["dependsOn"] = []any{}
	obj(other.Object["spec"])["context"] = M{"another": "context"}
	desired, err := Intent(other)
	if err != nil {
		t.Fatal(err)
	}
	other.Object["status"] = M{"observedGeneration": int64(1), "execution": M{"intent": desired}, "conditions": []any{M{"type": "Ready", "status": "True", "observedGeneration": int64(1)}}}
	if dependencyReason([]unstructured.Unstructured{*o, *other}, o, spec, "apply") != "" {
		t.Fatal("dependência pronta de outro contexto deveria ser aceita")
	}
	other.SetGeneration(2)
	if dependencyReason([]unstructured.Unstructured{*o, *other}, o, spec, "apply") != "DependencyBlocked" {
		t.Fatal("geração obsoleta aceita")
	}
}
func TestRetainAndDestroy(t *testing.T) {
	for _, destroy := range []bool{false, true} {
		t.Run(map[bool]string{false: "retain", true: "destroy"}[destroy], func(t *testing.T) {
			r, _ := fixture(t)
			if destroy {
				change(t, r, "lifecycle", M{"destroyResourcesOnDeletion": true})
			}
			tick(t, r, 1)
			o := getStack(t, r)
			if e := r.Delete(context.Background(), o); e != nil {
				t.Fatal(e)
			}
			tick(t, r, 4)
			if destroy {
				if obj(getRun(t, r)["intent"])["operation"] != "destroy" {
					t.Fatal("intenção destroy")
				}
				grantAndClaim(t, r)
				workflowStatus(t, r, "Succeeded", "Waiting", "Verified", "hash")
				tick(t, r, 2)
			}
			check := stackObject()
			e := r.Get(context.Background(), client.ObjectKey{Namespace: r.StackNamespace, Name: name}, check)
			if !apierrors.IsNotFound(e) {
				t.Fatalf("finalizer: %v", e)
			}
		})
	}
}
func TestBlockedResult(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	workflowStatus(t, r, "Succeeded", "Running", "Blocked", "")
	tick(t, r, 1)
	result := obj(obj(getStack(t, r).Object["status"])["result"])
	if reason(t, r) != "ExecutionBlocked" || result["reason"] != "FixtureDiagnostic" {
		t.Fatal(reason(t, r), result)
	}
}

func TestEnvtestStatusCAS(t *testing.T) {
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		t.Skip("configure KUBEBUILDER_ASSETS")
	}
	env := &envtest.Environment{CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "crd")}, ErrorIfCRDPathMissing: true}
	cfg, e := env.Start()
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := env.Stop(); e != nil {
			t.Error(e)
		}
	}()
	c, e := client.New(cfg, client.Options{})
	if e != nil {
		t.Fatal(e)
	}
	ns := &unstructured.Unstructured{Object: M{"apiVersion": "v1", "kind": "Namespace", "metadata": M{"name": "platform-system"}}}
	if e = c.Create(context.Background(), ns); e != nil {
		t.Fatal(e)
	}
	_, o := fixture(t)
	o.SetUID("")
	o.SetResourceVersion("")
	o.SetGeneration(0)
	if e = c.Create(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	stale := o.DeepCopy()
	o.Object["status"] = M{"claim": M{"workflow": "test-wf", "intentHash": strings.Repeat("a", 64), "authorization": "test-token"}}
	if e = c.Status().Update(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	stale.Object["status"] = M{"execution": M{"cancelling": true}}
	if e = c.Status().Update(context.Background(), stale); !apierrors.IsConflict(e) {
		t.Fatalf("esperava conflito, obtido %v", e)
	}

	// A partir da mesma versão, claim e revogação disputam o status.
	ctx := context.Background()
	o.Object["status"] = M{"execution": M{"authorization": "token"}}
	if err := c.Status().Update(ctx, o); err != nil {
		t.Fatal(err)
	}
	cancel, applying := o.DeepCopy(), o.DeepCopy()
	cancel.Object["status"] = M{"execution": M{"cancelling": true}}
	patch, err := json.Marshal([]M{
		{"op": "test", "path": "/metadata/resourceVersion", "value": o.GetResourceVersion()},
		{"op": "add", "path": "/status/claim", "value": M{
			"workflow": "test-wf", "intentHash": strings.Repeat("a", 64), "authorization": "token",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	start, results := make(chan struct{}), make(chan error, 2)
	go func() {
		<-start
		results <- c.Status().Update(ctx, cancel)
	}()
	go func() {
		<-start
		results <- c.Status().Patch(ctx, applying, client.RawPatch(types.JSONPatchType, patch))
	}()
	close(start)
	first, second := <-results, <-results
	if (first == nil) == (second == nil) {
		t.Fatalf("esperava apenas um vencedor, obtidos %v e %v", first, second)
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(o), o); err != nil {
		t.Fatal(err)
	}
	status := obj(o.Object["status"])
	cancelled, claimed := yes(obj(status["execution"])["cancelling"]), len(obj(status["claim"])) != 0
	if cancelled == claimed {
		t.Fatalf("estado concorrente incompatível: %v", status)
	}
}

func TestRenamedIntegration(t *testing.T) {
	r, _ := fixture(t)
	r.Engine.(*argo.Adapter).Config.TemplateRef.Name = "custom-workflow"
	r.Engine.(*argo.Adapter).Config.Parameters.Input.Intent = "request-json"
	r.Engine.(*argo.Adapter).Config.Parameters.Output.State = "policy-output"
	r.Engine.(*argo.Adapter).Config.Parameters.Output.Authorization = "authorized-hash"
	r.Engine.(*argo.Adapter).Config.Parameters.Output.Result = "verification"
	tick(t, r, 4)
	w := current(t, r)
	spec := obj(w.Object["spec"])
	if obj(spec["workflowTemplateRef"])["name"] != "custom-workflow" {
		t.Fatal(spec)
	}
	inputs := arr(obj(spec["arguments"])["parameters"])
	if len(inputs) != 1 || obj(inputs[0])["name"] != "request-json" {
		t.Fatal(inputs)
	}
	workflowStatus(t, r, "Running", "Running", "", "hash")
	tick(t, r, 1)
	if getRun(t, r)["authorization"] != "hash" {
		t.Fatal("contrato renomeado não reconhecido")
	}
}
func TestIntegrationChangeBlocksActiveRun(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	old := getRun(t, r)["workflow"]
	r.Engine.(*argo.Adapter).Config.Parameters.Output.State = "different-decision"
	tick(t, r, 1)
	if reason(t, r) != "IntegrationChanged" || getRun(t, r)["workflow"] != old {
		t.Fatal("execução não preservada")
	}
}
func TestWorkflowOperationalFailure(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	w := current(t, r)
	w.Object["status"] = M{"phase": "Failed", "nodes": M{"arbitrary": M{"templateName": "arbitrary-step", "phase": "Error"}}}
	if err := r.Status().Update(context.Background(), w); err != nil {
		t.Fatal(err)
	}
	tick(t, r, 1)
	if reason(t, r) != "ExecutionFailed" {
		t.Fatal(reason(t, r))
	}
}

func TestUnclaimedVerifiedResult(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	workflowStatus(t, r, "Succeeded", "Running", "Verified", "hash")
	tick(t, r, 1)
	if reason(t, r) != "InvalidExecutionContract" {
		t.Fatal(reason(t, r))
	}
}

func TestAuthorizationCannotChangeAfterGrant(t *testing.T) {
	for _, changed := range []string{"", "changed-token"} {
		t.Run(changed, func(t *testing.T) {
			r, _ := fixture(t)
			tick(t, r, 4)
			workflowStatus(t, r, "Running", "Running", "", "hash")
			tick(t, r, 1)
			if getRun(t, r)["authorization"] != "hash" {
				t.Fatal(getRun(t, r))
			}
			workflowStatus(t, r, "Running", "Running", "", changed)
			tick(t, r, 2)
			if getRun(t, r)["authorization"] != nil ||
				obj(current(t, r).Object["spec"])["shutdown"] != "Terminate" {
				t.Fatal(getRun(t, r))
			}
		})
	}
}
