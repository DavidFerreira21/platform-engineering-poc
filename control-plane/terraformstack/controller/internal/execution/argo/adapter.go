package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type M = execution.Data

const uidLabel = "platform.example.io/stack-uid"

type Adapter struct {
	Client client.Client
	Reader client.Reader
	Config WorkflowConfig
}

var _ execution.Engine = (*Adapter)(nil)

func obj(v any) M                  { m, _ := v.(map[string]any); return m }
func str(v any) string             { s, _ := v.(string); return s }
func arr(v any) []any              { a, _ := v.([]any); return a }
func canonical(v any) string       { b, _ := json.Marshal(v); return string(b) }
func (a *Adapter) Validate() error { return a.Config.Validate() }
func (a *Adapter) Snapshot() M     { return a.Config.Snapshot() }
func WorkflowObject() *unstructured.Unstructured {
	o := &unstructured.Unstructured{}
	o.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Version: "v1alpha1", Kind: "Workflow"})
	return o
}
func (a *Adapter) Raw(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	if name == "" {
		return nil, nil
	}
	w := WorkflowObject()
	e := a.Reader.Get(ctx, client.ObjectKey{Namespace: a.Config.Namespace, Name: name}, w)
	if apierrors.IsNotFound(e) {
		return nil, nil
	}
	return w, e
}
func (a *Adapter) body(run M) *unstructured.Unstructured {
	d := obj(run["intent"])
	w := WorkflowObject()
	w.SetName(str(run["workflow"]))
	w.SetNamespace(a.Config.Namespace)
	w.SetLabels(map[string]string{"platform.example.io/workflow": "terraformstack", uidLabel: str(d["uid"]), "platform.example.io/stack-name": str(d["stack"])})
	w.Object["spec"] = M{"serviceAccountName": d["serviceAccount"], "workflowTemplateRef": M{"name": a.Config.TemplateRef.Name}, "arguments": M{"parameters": []any{M{"name": a.Config.Parameters.Input.Intent, "value": canonical(d)}, M{"name": a.Config.Parameters.Input.Volume, "value": d["volume"]}}}}
	return w
}

func (a *Adapter) Create(ctx context.Context, run M) error {
	return a.Client.Create(ctx, a.body(run))
}
func (a *Adapter) valid(w *unstructured.Unstructured, uid string) bool {
	return w.GetLabels()[uidLabel] == uid && str(obj(obj(w.Object["spec"])["workflowTemplateRef"])["name"]) == a.Config.TemplateRef.Name
}
func (a *Adapter) Get(ctx context.Context, name, uid string) (*execution.Run, error) {
	w, e := a.Raw(ctx, name)
	if e != nil || w == nil {
		return nil, e
	}
	phase := str(obj(w.Object["status"])["phase"])
	result, e := resultOf(w, a.Config)
	if e != nil {
		return nil, e
	}
	return &execution.Run{Name: w.GetName(), Valid: a.valid(w, uid), Terminal: phase == "Succeeded" || phase == "Failed" || phase == "Error", Succeeded: phase == "Succeeded", Result: result, Signals: signalsOf(result)}, nil
}
func (a *Adapter) Cancel(ctx context.Context, name, uid string) error {
	w, e := a.Raw(ctx, name)
	if e != nil || w == nil {
		return e
	}
	if !a.valid(w, uid) {
		return fmt.Errorf("identidade de execução inválida")
	}
	spec := obj(w.Object["spec"])
	if spec["shutdown"] == "Terminate" {
		return nil
	}
	old := w.DeepCopy()
	spec["shutdown"] = "Terminate"
	return a.Client.Patch(ctx, w, client.MergeFromWithOptions(old, client.MergeFromWithOptimisticLock{}))
}
func resultOf(w *unstructured.Unstructured, integration WorkflowConfig) (M, error) {
	p := integration.Parameters.Output
	result := M{}
	if w == nil {
		return result, nil
	}
	status := obj(w.Object["status"])
	outputs := M{}
	for _, p := range arr(obj(status["outputs"])["parameters"]) {
		m := obj(p)
		outputs[str(m["name"])] = m["value"]
	}
	if raw := str(outputs[p.Decision]); raw != "" {
		if e := decodeDecision(raw, &result); e != nil || result == nil {
			return M{"decision": "Error", "reason": "Resultado OPA inválido"}, nil
		}
	}
	for _, v := range obj(status["nodes"]) {
		n := obj(v)
		if n["templateName"] == integration.Templates.Policy && (n["phase"] == "Failed" || n["phase"] == "Error") {
			result["decision"] = "Error"
			result["reason"] = "OPA falhou ao avaliar o plano"
		}
		if n["type"] == "Suspend" && n["templateName"] == integration.Templates.Approval {
			result["approval"] = n["phase"]
			if a := w.GetLabels()["workflows.argoproj.io/actor"]; a != "" {
				result["approvalActor"] = a
			}
			if m := str(n["message"]); m != "" {
				r := []rune(m)
				if len(r) > 512 {
					r = r[:512]
				}
				result["approvalMessage"] = string(r)
			}
		}
	}
	if r := str(outputs[p.Authorization]); r != "" && r != "none" {
		result["applyRequest"] = r
	}
	if outputs[p.Approval] == "Rejected" {
		result["outcome"] = "Rejected"
	}
	if r := str(outputs[p.Result]); r != "" {
		result["outcome"] = r
	}
	return result, nil
}

// signalsOf translates the existing reference pipeline's protocol.
func signalsOf(result M) execution.Signals {
	s := execution.Signals{
		Waiting:       result["approval"] == "Running",
		DriftDetected: result["drift"] == true,
	}
	switch {
	case result["outcome"] == "Cancelled":
		s.Outcome, s.Reason = execution.OutcomeCancelled, "Cancelled"
	case result["outcome"] == "Rejected":
		s.Outcome, s.Reason = execution.OutcomeBlocked, "ApprovalRejected"
	case result["decision"] == "Deny":
		s.Outcome, s.Reason = execution.OutcomeBlocked, "PolicyDenied"
	case result["decision"] == "Error":
		s.Outcome, s.Reason = execution.OutcomeBlocked, "PolicyError"
	case result["decision"] == "NoChanges":
		s.Outcome, s.Reason = execution.OutcomeNoChanges, "NoChanges"
	case result["outcome"] == "Verified":
		s.Outcome, s.Reason = execution.OutcomeVerified, "Verified"
	}
	request := str(result["applyRequest"])
	if s.Outcome == execution.OutcomeUnknown && !s.Waiting && request != "" &&
		request == str(result["hash"]) &&
		(result["decision"] == "AutoApply" || result["decision"] == "RequireApproval") {
		s.AuthorizationRequest = request
	}
	return s
}

func decodeDecision(raw string, result *M) error {
	// Preserve integral JSON values when writing Kubernetes status.
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if e := dec.Decode(result); e != nil {
		return e
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		return fmt.Errorf("resultado contém JSON adicional ou inválido")
	}
	for k, v := range *result {
		(*result)[k] = normalizeValue(v)
	}
	return nil
}
func normalizeValue(v any) any {
	switch x := v.(type) {
	case json.Number:
		if n, e := x.Int64(); e == nil {
			return n
		}
		n, _ := x.Float64()
		return n
	case M:
		for k, val := range x {
			x[k] = normalizeValue(val)
		}
	case []any:
		for i, val := range x {
			x[i] = normalizeValue(val)
		}
	}
	return v
}
