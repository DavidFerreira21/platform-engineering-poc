package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	strictjson "sigs.k8s.io/json"
)

type M = execution.Data

const uidLabel = "platform.example.io/stack-uid"
const stackNameKey = "platform.example.io/stack-name"

type Adapter struct {
	Client client.Client
	Reader client.Reader
	Config WorkflowConfig
}

var _ execution.Engine = (*Adapter)(nil)

func obj(v any) M      { m, _ := v.(map[string]any); return m }
func str(v any) string { s, _ := v.(string); return s }
func arr(v any) []any  { a, _ := v.([]any); return a }

func (a *Adapter) Validate() error { return a.Config.Validate() }
func (a *Adapter) Snapshot() M     { return a.Config.Snapshot() }

func WorkflowObject() *unstructured.Unstructured {
	o := &unstructured.Unstructured{}
	o.SetGroupVersionKind(schema.GroupVersionKind{Group: "argoproj.io", Version: "v1alpha1", Kind: "Workflow"})
	return o
}

func (a *Adapter) Raw(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	w := WorkflowObject()
	err := a.Reader.Get(ctx, client.ObjectKey{Namespace: a.Config.Namespace, Name: name}, w)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	return w, err
}

func (a *Adapter) body(req execution.Request) (*unstructured.Unstructured, error) {
	intent, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	w := WorkflowObject()
	w.SetName(req.Workflow)
	w.SetNamespace(a.Config.Namespace)
	labels := map[string]string{
		"platform.example.io/workflow": "terraformstack",
		uidLabel:                       req.Intent.UID,
	}
	if len(req.Intent.Stack) <= 63 {
		labels[stackNameKey] = req.Intent.Stack
	} else {
		w.SetAnnotations(map[string]string{stackNameKey: req.Intent.Stack})
	}
	w.SetLabels(labels)
	w.Object["spec"] = M{
		"serviceAccountName":  a.Config.ServiceAccountName,
		"workflowTemplateRef": M{"name": a.Config.TemplateRef.Name},
		"arguments":           M{"parameters": []any{M{"name": a.Config.Parameters.Input.Intent, "value": string(intent)}}},
	}
	return w, nil
}

func (a *Adapter) Create(ctx context.Context, req execution.Request) error {
	w, err := a.body(req)
	if err != nil {
		return err
	}
	err = a.Client.Create(ctx, w)
	if err == nil {
		log.FromContext(ctx).Info("workflow created", "event", "execution.created", "workflow", req.Workflow, "intentHash", req.IntentHash)
	}
	return err
}

func (a *Adapter) valid(w *unstructured.Unstructured, req execution.Request) bool {
	labels, spec := w.GetLabels(), obj(w.Object["spec"])
	if w.GetName() != req.Workflow || w.GetNamespace() != a.Config.Namespace ||
		labels["platform.example.io/workflow"] != "terraformstack" ||
		labels[uidLabel] != req.Intent.UID || stackName(w) != req.Intent.Stack ||
		spec["serviceAccountName"] != a.Config.ServiceAccountName ||
		obj(spec["workflowTemplateRef"])["name"] != a.Config.TemplateRef.Name ||
		obj(spec["workflowTemplateRef"])["clusterScope"] == true {
		return false
	}
	params := arr(obj(spec["arguments"])["parameters"])
	if len(params) != 1 || obj(params[0])["name"] != a.Config.Parameters.Input.Intent {
		return false
	}
	expected, err := json.Marshal(req)
	return err == nil && str(obj(params[0])["value"]) == string(expected)
}

func (a *Adapter) Get(ctx context.Context, req execution.Request) (*execution.Run, error) {
	w, err := a.Raw(ctx, req.Workflow)
	if err != nil || w == nil {
		return nil, err
	}
	if !a.valid(w, req) {
		log.FromContext(ctx).Info("workflow identity invalid", "event", "execution.identity.invalid", "workflow", req.Workflow, "intentHash", req.IntentHash)
		return &execution.Run{}, nil
	}
	return observeWithDiagnostic(w, a.Config.Parameters.Output, req, func(reason, output string) {
		log.FromContext(ctx).Info("execution contract invalid", "event", "execution.contract.invalid", "workflow", req.Workflow, "intentHash", req.IntentHash, "reason", reason, "output", output)
	}), nil
}

func (a *Adapter) Cancel(ctx context.Context, req execution.Request) error {
	w, err := a.Raw(ctx, req.Workflow)
	if err != nil || w == nil {
		return err
	}
	if !a.valid(w, req) {
		return fmt.Errorf("identidade de execução inválida")
	}
	spec := obj(w.Object["spec"])
	if spec["shutdown"] == "Terminate" {
		return nil
	}
	old := w.DeepCopy()
	spec["shutdown"] = "Terminate"
	err = a.Client.Patch(ctx, w, client.MergeFromWithOptions(old, client.MergeFromWithOptimisticLock{}))
	if err == nil {
		log.FromContext(ctx).Info("workflow termination requested", "event", "execution.cancel.requested", "workflow", req.Workflow, "intentHash", req.IntentHash)
	}
	return err
}

var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:@/+~-]{0,255}$`)
var code = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,63}$`)

func safeDiagnostics(d execution.Diagnostics) bool {
	if (d.Reason != "" && !code.MatchString(d.Reason)) || len(d.Message) > 256 ||
		strings.ContainsFunc(d.Message, unicode.IsControl) || len(d.Details) > 12 {
		return false
	}
	for key, val := range d.Details {
		k := strings.ToLower(key)
		if !code.MatchString(key) || strings.Contains(k, "secret") || strings.Contains(k, "token") ||
			strings.Contains(k, "credential") || strings.Contains(k, "password") ||
			len(val) > 160 || strings.ContainsFunc(val, unicode.IsControl) {
			return false
		}
	}
	return true
}

func decode(raw string, into any) bool {
	if raw == "" || len(raw) > 4096 {
		return false
	}
	strict, err := strictjson.UnmarshalStrict([]byte(raw), into)
	return err == nil && len(strict) == 0
}

func observe(w *unstructured.Unstructured, names OutputParameters, req execution.Request) *execution.Run {
	return observeWithDiagnostic(w, names, req, func(string, string) {})
}

func observeWithDiagnostic(w *unstructured.Unstructured, names OutputParameters, req execution.Request, diagnostic func(string, string)) *execution.Run {
	status := obj(w.Object["status"])
	native := str(status["phase"])
	phase := execution.Running
	terminal := native == "Succeeded" || native == "Failed" || native == "Error"
	if native == "" || native == "Pending" {
		phase = execution.Pending
	}
	out := &execution.Run{Valid: true, Phase: phase}
	invalid := func(reason, output string) *execution.Run {
		diagnostic(reason, output)
		if terminal {
			phase = execution.Blocked
		}
		return &execution.Run{Valid: true, Phase: phase, ContractError: true}
	}
	if native != "" && native != "Pending" && native != "Running" && !terminal {
		return invalid("InvalidNativePhase", "")
	}

	outputs := map[string]string{}
	for _, p := range arr(obj(status["outputs"])["parameters"]) {
		param := obj(p)
		key := str(param["name"])
		switch key {
		case names.State, names.Authorization, names.Result, names.ResolvedSource:
			value, ok := param["value"].(string)
			if _, exists := outputs[key]; exists || !ok {
				return invalid("DuplicateOrNonStringOutput", key)
			}
			outputs[key] = value
		}
	}

	if raw := outputs[names.ResolvedSource]; raw != "" {
		var source execution.ResolvedSource
		if !decode(raw, &source) || source.Identity != req.Identity ||
			!opaque.MatchString(source.Reference) || !opaque.MatchString(source.Binding) ||
			(req.PinnedReference != "" && source.Reference != req.PinnedReference) ||
			(req.Binding != "" && source.Binding != req.Binding) {
			return invalid("InvalidResolvedSource", names.ResolvedSource)
		}
		out.Resolution = &source.Resolution
	}
	if raw := outputs[names.State]; raw != "" {
		var progress execution.Progress
		if !decode(raw, &progress) || progress.Identity != req.Identity ||
			(progress.Phase != execution.Running && progress.Phase != execution.Waiting) ||
			!safeDiagnostics(progress.Diagnostics) || out.Resolution == nil {
			return invalid("InvalidProgress", names.State)
		}
		out.Progress = &progress
	}
	if raw := outputs[names.Authorization]; raw != "" {
		var approval execution.Authorization
		if !decode(raw, &approval) || approval.Identity != req.Identity ||
			!opaque.MatchString(approval.Token) || out.Resolution == nil {
			return invalid("InvalidAuthorization", names.Authorization)
		}
		out.Authorization = approval.Token
	}
	if raw := outputs[names.Result]; raw != "" {
		var result execution.Result
		if !decode(raw, &result) || result.Identity != req.Identity || !safeDiagnostics(result.Diagnostics) {
			return invalid("InvalidResult", names.Result)
		}
		switch result.Outcome {
		case execution.NoChanges:
			if req.Intent.Operation != execution.Apply || out.Resolution == nil || out.Authorization != "" ||
				(out.Progress != nil && out.Progress.Phase == execution.Waiting) || (terminal && native != "Succeeded") {
				return invalid("InconsistentNoChanges", names.Result)
			}
			if terminal {
				out.Phase = execution.Succeeded
			}
		case execution.Verified:
			if out.Resolution == nil || out.Authorization == "" || (terminal && native != "Succeeded") {
				return invalid("InconsistentVerified", names.Result)
			}
			if terminal {
				out.Phase = execution.Succeeded
			}
		case execution.OutcomeBlocked:
			if out.Authorization != "" {
				return invalid("UnexpectedAuthorization", names.Result)
			}
			if terminal {
				out.Phase = execution.Blocked
			}
		case execution.OutcomeCancelled:
			if terminal {
				out.Phase = execution.Cancelled
			}
		default:
			return invalid("UnknownOutcome", names.Result)
		}
		out.Result = &result
		return out
	}
	switch {
	case native == "Succeeded":
		return invalid("MissingTerminalResult", names.Result)
	case terminal:
		out.Phase = execution.Failed
	case out.Authorization != "":
		out.Phase = execution.AuthorizationRequested
	case out.Progress != nil && out.Progress.Phase == execution.Waiting:
		out.Phase = execution.Waiting
	}
	return out
}
