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
	c, e := LoadCatalog("../../../overlays/kind/catalog.json")
	if e != nil {
		t.Fatal(e)
	}
	o := stackObject()
	o.SetName(name)
	o.SetNamespace("platform-system")
	o.SetUID(types.UID("test-uid"))
	o.SetGeneration(1)
	o.SetAnnotations(map[string]string{revision: strings.Repeat("a", 40)})
	o.Object["spec"] = M{"sourceRef": M{"name": "foundation-stacks", "path": "foundation/stacks/simulator"}, "target": M{"entity": "entity-a", "environment": "simulator"}}
	cl := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithStatusSubresource(stackObject(), argo.WorkflowObject()).WithObjects(o).Build()
	config, e := LoadConfig("../../controller-config.yaml")
	if e != nil {
		t.Fatal(e)
	}
	r := &Reconciler{Engine: &argo.Adapter{Client: cl, Reader: cl, Config: config.Workflow}, Client: cl, Reader: cl, Catalog: c, StackNamespace: "platform-system", Now: func() time.Time { return fixed }}
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
func workflowStatus(t *testing.T, r *Reconciler, phase, decision, outcome, request string) {
	t.Helper()
	w := current(t, r)
	record, _ := json.Marshal(M{"decision": decision, "hash": "hash"})
	p := []any{M{"name": r.Engine.(*argo.Adapter).Config.Parameters.Output.Decision, "value": string(record)}}
	if outcome != "" {
		p = append(p, M{"name": r.Engine.(*argo.Adapter).Config.Parameters.Output.Result, "value": outcome})
	}
	if request != "" {
		p = append(p, M{"name": r.Engine.(*argo.Adapter).Config.Parameters.Output.Authorization, "value": request})
	}
	w.Object["status"] = M{"phase": phase, "outputs": M{"parameters": p}}
	if e := r.Status().Update(context.Background(), w); e != nil {
		t.Fatal(e)
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
	r, o := fixture(t)
	d, e := Intent(o, r.Catalog)
	if e != nil || d["policyProfile"] != "simulator" || d["revision"] != strings.Repeat("a", 40) {
		t.Fatalf("%v %v", d, e)
	}
	obj(obj(o.Object["spec"])["sourceRef"])["path"] = "../x"
	if _, e = Intent(o, r.Catalog); e == nil {
		t.Fatal("fonte não autorizada")
	}
}
func TestCatalogProfile(t *testing.T) {
	for _, p := range []any{nil, "", "../network"} {
		r, o := fixture(t)
		r.Catalog.Targets["entity-a/simulator"].Stacks[name]["policyProfile"] = p
		if _, e := Intent(o, r.Catalog); e == nil {
			t.Fatal("perfil inválido aceito")
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
	for _, s := range []struct {
		d, q  string
		allow bool
	}{{"AutoApply", "hash", true}, {"RequireApproval", "hash", true}, {"RequireApproval", "bad", false}, {"Deny", "hash", false}, {"Error", "hash", false}} {
		t.Run(s.d+s.q, func(t *testing.T) {
			r, _ := fixture(t)
			tick(t, r, 4)
			workflowStatus(t, r, "Running", s.d, "", s.q)
			tick(t, r, 1)
			if (getRun(t, r)["authorization"] == "hash") != s.allow {
				t.Fatal(getRun(t, r))
			}
		})
	}
}
func TestCancelBeforeApply(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	change(t, r, "configuration", M{"version": strings.Repeat("b", 40)})
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
	o := getStack(t, r)
	obj(obj(o.Object["status"])["execution"])["applying"] = true
	if e := r.Status().Update(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	change(t, r, "configuration", M{"version": strings.Repeat("b", 40)})
	tick(t, r, 1)
	if obj(current(t, r).Object["spec"])["shutdown"] != nil {
		t.Fatal("apply interrompido")
	}
	workflowStatus(t, r, "Succeeded", "AutoApply", "Verified", "")
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
	if reason(t, r) != "WorkflowMissing" {
		t.Fatal(reason(t, r))
	}
}
func TestDriftTimer(t *testing.T) {
	r, _ := fixture(t)
	change(t, r, "reconciliation", M{"driftCheckInterval": "45s"})
	tick(t, r, 4)
	workflowStatus(t, r, "Succeeded", "NoChanges", "", "")
	tick(t, r, 1)
	result := tick(t, r, 1)
	if result.RequeueAfter != 45*time.Second {
		t.Fatal(result)
	}
	r.Now = func() time.Time { return fixed.Add(45 * time.Second) }
	tick(t, r, 2)
	if getRun(t, r)["trigger"] != "DriftCheck" {
		t.Fatal(getRun(t, r))
	}
}
func TestRetry(t *testing.T) {
	r, _ := fixture(t)
	change(t, r, "reconciliation", M{"retryPolicy": M{"strategy": "Fixed", "interval": "10s", "maxRetries": int64(1)}})
	tick(t, r, 4)
	workflowStatus(t, r, "Failed", "AutoApply", "", "")
	tick(t, r, 1)
	if tick(t, r, 1).RequeueAfter != 10*time.Second {
		t.Fatal("timer")
	}
	r.Now = func() time.Time { return fixed.Add(10 * time.Second) }
	tick(t, r, 2)
	if num(getRun(t, r)["attempt"]) != 1 {
		t.Fatal(getRun(t, r))
	}
	workflowStatus(t, r, "Failed", "AutoApply", "", "")
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
	r, o := fixture(t)
	obj(o.Object["spec"])["dependsOn"] = []any{M{"name": "simulator-dependent"}}
	d, e := Intent(o, r.Catalog)
	if e != nil {
		t.Fatal(e)
	}
	if DependenciesReady(nil, o, d, r.Catalog) {
		t.Fatal("dependência ausente")
	}
	other := o.DeepCopy()
	other.SetName("simulator-dependent")
	obj(other.Object["spec"])["dependsOn"] = []any{}
	desired, e := Intent(other, r.Catalog)
	if e != nil {
		t.Fatal(e)
	}
	other.Object["status"] = M{"observedGeneration": int64(1), "execution": M{"intent": desired}, "conditions": []any{M{"type": "Ready", "status": "True", "observedGeneration": int64(1)}}}
	if !DependenciesReady([]unstructured.Unstructured{*other}, o, d, r.Catalog) {
		t.Fatal("dependência pronta")
	}
	other.SetGeneration(2)
	if DependenciesReady([]unstructured.Unstructured{*other}, o, d, r.Catalog) {
		t.Fatal("geração obsoleta")
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
				workflowStatus(t, r, "Succeeded", "RequireApproval", "Verified", "")
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
func TestPolicyError(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	w := current(t, r)
	w.Object["status"] = M{"phase": "Failed", "nodes": M{"opa": M{"templateName": "opa", "phase": "Error"}}}
	if e := r.Status().Update(context.Background(), w); e != nil {
		t.Fatal(e)
	}
	tick(t, r, 1)
	if reason(t, r) != "PolicyError" {
		t.Fatal(reason(t, r))
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
	o.Object["status"] = M{"execution": M{"applying": true}}
	if e = c.Status().Update(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	stale.Object["status"] = M{"execution": M{"applying": false}}
	if e = c.Status().Update(context.Background(), stale); !apierrors.IsConflict(e) {
		t.Fatalf("esperava conflito, obtido %v", e)
	}
}

func TestRenamedIntegration(t *testing.T) {
	r, _ := fixture(t)
	r.Engine.(*argo.Adapter).Config.TemplateRef.Name = "custom-workflow"
	r.Engine.(*argo.Adapter).Config.Parameters.Input.Intent = "request-json"
	r.Engine.(*argo.Adapter).Config.Parameters.Input.Volume = "workspace"
	r.Engine.(*argo.Adapter).Config.Parameters.Output.Decision = "policy-output"
	r.Engine.(*argo.Adapter).Config.Parameters.Output.Authorization = "authorized-hash"
	r.Engine.(*argo.Adapter).Config.Parameters.Output.Result = "verification"
	tick(t, r, 4)
	w := current(t, r)
	spec := obj(w.Object["spec"])
	if obj(spec["workflowTemplateRef"])["name"] != "custom-workflow" {
		t.Fatal(spec)
	}
	inputs := arr(obj(spec["arguments"])["parameters"])
	if obj(inputs[0])["name"] != "request-json" || obj(inputs[1])["name"] != "workspace" {
		t.Fatal(inputs)
	}
	workflowStatus(t, r, "Running", "AutoApply", "", "hash")
	tick(t, r, 1)
	if getRun(t, r)["authorization"] != "hash" {
		t.Fatal("contrato renomeado não reconhecido")
	}
}
func TestIntegrationChangeBlocksActiveRun(t *testing.T) {
	r, _ := fixture(t)
	tick(t, r, 4)
	old := getRun(t, r)["workflow"]
	r.Engine.(*argo.Adapter).Config.Parameters.Output.Decision = "different-decision"
	tick(t, r, 1)
	if reason(t, r) != "IntegrationChanged" || getRun(t, r)["workflow"] != old {
		t.Fatal("execução não preservada")
	}
}
func TestRenamedPolicyFailure(t *testing.T) {
	r, _ := fixture(t)
	r.Engine.(*argo.Adapter).Config.Templates.Policy = "policy-check"
	tick(t, r, 4)
	w := current(t, r)
	w.Object["status"] = M{"phase": "Failed", "nodes": M{"policy": M{"templateName": "policy-check", "phase": "Error"}}}
	if e := r.Status().Update(context.Background(), w); e != nil {
		t.Fatal(e)
	}
	tick(t, r, 1)
	if reason(t, r) != "PolicyError" {
		t.Fatal(reason(t, r))
	}
}
