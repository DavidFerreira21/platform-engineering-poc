package argo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testConfig() WorkflowConfig {
	return WorkflowConfig{
		Namespace: "argo-workflows", TemplateRef: TemplateRef{Name: "custom-workflow"}, ServiceAccountName: "executor",
		Parameters: Parameters{
			Input:  InputParameters{Intent: "request-json"},
			Output: OutputParameters{State: "progress", Authorization: "authorization", Result: "result", ResolvedSource: "source"},
		},
	}
}

func testRequest() execution.Request {
	return execution.Request{
		Identity: execution.Identity{Workflow: "test-run", IntentHash: strings.Repeat("a", 64)},
		Intent:   execution.Intent{UID: "uid", Stack: "stack", Namespace: "stacks", Operation: execution.Apply},
	}
}

func TestAdapterLifecycle(t *testing.T) {
	ctx := context.Background()
	cl := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithStatusSubresource(WorkflowObject()).Build()
	a := &Adapter{Client: cl, Reader: cl, Config: testConfig()}
	req := testRequest()
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	if w, err := a.Get(ctx, req); err != nil || w != nil {
		t.Fatalf("missing: %v %v", w, err)
	}
	if err := a.Create(ctx, req); err != nil {
		t.Fatal(err)
	}
	w, err := a.Raw(ctx, req.Workflow)
	if err != nil {
		t.Fatal(err)
	}
	spec := obj(w.Object["spec"])
	if str(obj(spec["workflowTemplateRef"])["name"]) != "custom-workflow" || spec["serviceAccountName"] != "executor" {
		t.Fatal(spec)
	}
	inputs := arr(obj(spec["arguments"])["parameters"])
	if len(inputs) != 1 || obj(inputs[0])["name"] != "request-json" {
		t.Fatal(inputs)
	}
	wrong := req
	wrong.IntentHash = strings.Repeat("b", 64)
	if err := a.Cancel(ctx, wrong); err == nil {
		t.Fatal("invalid owner accepted")
	}
	if got, err := a.Get(ctx, wrong); err != nil || got.Valid {
		t.Fatalf("invalid identity: %v %v", got, err)
	}
	for i := 0; i < 2; i++ {
		if err := a.Cancel(ctx, req); err != nil {
			t.Fatal(err)
		}
	}
	w, err = a.Raw(ctx, req.Workflow)
	if err != nil || obj(w.Object["spec"])["shutdown"] != "Terminate" {
		t.Fatalf("shutdown: %v %v", w, err)
	}
}

func TestLongStackName(t *testing.T) {
	a := &Adapter{Config: testConfig()}
	req := testRequest()
	req.Intent.Stack = strings.Repeat("a", 63) + ".network"
	w, err := a.body(req)
	if err != nil {
		t.Fatal(err)
	}
	if w.GetLabels()[stackNameKey] != "" || stackName(w) != req.Intent.Stack || !a.valid(w, req) {
		t.Fatal(w.Object["metadata"])
	}
}

func TestResultNormalization(t *testing.T) {
	req := testRequest()
	encode := func(fields M) string {
		fields["workflow"], fields["intentHash"] = req.Workflow, req.IntentHash
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	source := encode(M{"reference": "immutable-rev", "binding": "stable-destination"})
	waiting := encode(M{"phase": "Waiting", "driftDetected": true})
	auth := encode(M{"token": "opaque-token"})
	verified := encode(M{"outcome": "Verified", "details": M{"customKey": "message"}})
	noChanges := encode(M{"outcome": "NoChanges"})
	blocked := encode(M{"outcome": "Blocked", "reason": "DestinationMismatch"})
	cancelled := encode(M{"outcome": "Cancelled"})
	for _, test := range []struct {
		name, native string
		outputs      map[string]string
		phase        execution.Phase
		invalid      bool
	}{
		{"pending", "", nil, execution.Pending, false},
		{"running", "Running", nil, execution.Running, false},
		{"failure", "Failed", nil, execution.Failed, false},
		{"waiting", "Running", map[string]string{"source": source, "progress": waiting}, execution.Waiting, false},
		{"authorization", "Running", map[string]string{"source": source, "progress": waiting, "authorization": auth}, execution.AuthorizationRequested, false},
		{"verified", "Succeeded", map[string]string{"source": source, "authorization": auth, "result": verified}, execution.Succeeded, false},
		{"no changes", "Succeeded", map[string]string{"source": source, "result": noChanges}, execution.Succeeded, false},
		{"blocked before resolution", "Succeeded", map[string]string{"result": blocked}, execution.Blocked, false},
		{"cancelled after gate", "Succeeded", map[string]string{"source": source, "authorization": auth, "result": cancelled}, execution.Cancelled, false},
		{"missing result", "Succeeded", nil, execution.Blocked, true},
		{"malformed", "Succeeded", map[string]string{"result": "null"}, execution.Blocked, true},
		{"trailing", "Succeeded", map[string]string{"result": blocked + "{}"}, execution.Blocked, true},
		{"unknown key", "Succeeded", map[string]string{"result": encode(M{"outcome": "Blocked", "unknown": true})}, execution.Blocked, true},
		{"duplicate key", "Succeeded", map[string]string{"result": strings.Replace(blocked, "{", `{"reason":"SecondReason",`, 1)}, execution.Blocked, true},
		{"wrong identity", "Running", map[string]string{"source": source, "authorization": strings.Replace(auth, req.IntentHash, strings.Repeat("b", 64), 1)}, execution.Running, true},
		{"token without revision", "Running", map[string]string{"authorization": auth}, execution.Running, true},
		{"verified without authorization", "Succeeded", map[string]string{"source": source, "result": verified}, execution.Blocked, true},
		{"no changes with authorization", "Succeeded", map[string]string{"source": source, "authorization": auth, "result": noChanges}, execution.Blocked, true},
		{"blocked with authorization", "Succeeded", map[string]string{"source": source, "authorization": auth, "result": blocked}, execution.Blocked, true},
		{"sensitive output", "Succeeded", map[string]string{"result": encode(M{"outcome": "Blocked", "details": M{"awsSecret": "value"}})}, execution.Blocked, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := WorkflowObject()
			var params []any
			for name, value := range test.outputs {
				params = append(params, M{"name": name, "value": value})
			}
			w.Object["status"] = M{"phase": test.native, "outputs": M{"parameters": params}}
			out := observe(w, testConfig().Parameters.Output, req)
			if out.Phase != test.phase || out.ContractError != test.invalid {
				t.Fatalf("%+v", out)
			}
		})
	}

	w := WorkflowObject()
	w.Object["status"] = M{"phase": "Running", "outputs": M{"parameters": []any{M{"name": "source", "value": source}}}}
	for _, mismatch := range []execution.Request{
		{Identity: req.Identity, Intent: req.Intent, Binding: "other-destination"},
		{Identity: req.Identity, Intent: req.Intent, PinnedReference: "other-rev"},
	} {
		if !observe(w, testConfig().Parameters.Output, mismatch).ContractError {
			t.Fatal("referência ou vínculo inesperado aceito")
		}
	}
}
