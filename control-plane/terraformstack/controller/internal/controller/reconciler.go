package controller

import (
	"context"
	"fmt"
	"slices"
	"time"

	execengine "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Reconciler struct {
	client.Client
	Reader          client.Reader     // Leitura consistente antes de decisões de autorização.
	Engine          execengine.Engine // Single integration for embedding and tests.
	Engines         map[string]execengine.Engine
	DefaultWorkflow string
	StackNamespace  string
	Now             func() time.Time
}

// resolveEngine uses only installation-authorized aliases. It never mutates the
// reconciler: concurrent stacks may use different integrations.
func (r *Reconciler) resolveEngine(name string) (string, execengine.Engine, error) {
	if len(r.Engines) == 0 && name == "" && r.Engine != nil {
		return "", r.Engine, r.Engine.Validate()
	}
	if name == "" {
		name = r.DefaultWorkflow
	}
	engine, ok := r.Engines[name]
	if !ok || engine == nil {
		return name, nil, fmt.Errorf("integração de Workflow não cadastrada: %s", name)
	}
	return name, engine, engine.Validate()
}

func authorizationRequired(snapshot M) bool {
	enabled, explicit := snapshot["authorizationRequired"].(bool)
	return !explicit || enabled
}

func (r *Reconciler) clock() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}
	return time.Now().UTC()
}
func (r *Reconciler) stamp() string { return r.clock().Format(time.RFC3339) }

// O status inteiro é atualizado com resourceVersion. Isso impede sobrescrever
// o claim feito pelo runner entre a leitura e a escrita.
func (r *Reconciler) report(ctx context.Context, o *unstructured.Unstructured, kind, reason, message string, values M) error {
	old := obj(o.Object["status"])
	status := copyMap(old)
	for k, v := range values {
		if v == nil {
			delete(status, k)
			continue
		}
		status[k] = v
	}
	item := M{"type": kind, "status": "True", "reason": reason, "message": message, "observedGeneration": o.GetGeneration(), "lastTransitionTime": r.stamp()}
	// Mantém a condição corrente em primeiro lugar (contrato dos labs), e
	// conserva as demais condições com False e seus tempos de transição.
	conditions := []any{item}
	for _, v := range arr(old["conditions"]) {
		c := copyMap(obj(v))
		if c["type"] == kind {
			if c["status"] == "True" {
				item["lastTransitionTime"] = c["lastTransitionTime"]
			}
			continue
		}
		if c["status"] != "False" {
			c["lastTransitionTime"] = r.stamp()
		}
		c["status"] = "False"
		c["observedGeneration"] = o.GetGeneration()
		conditions = append(conditions, c)
	}
	status["observedGeneration"] = o.GetGeneration()
	status["conditions"] = conditions
	if equal(status, old) {
		return nil
	}
	o.Object["status"] = status
	e := r.Status().Update(ctx, o)
	if e == nil {
		log.FromContext(ctx).Info("transição TerraformStack", "stack", o.GetName(), "workflowRef", str(obj(status["execution"])["workflowRef"]), "condition", kind, "reason", reason)
	}
	return e
}
func again(e error) (ctrl.Result, error) { return ctrl.Result{Requeue: true}, e }
func done(e error) (ctrl.Result, error)  { return ctrl.Result{}, e }
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace != r.StackNamespace {
		return done(nil)
	}
	o := stackObject()
	if e := r.Reader.Get(ctx, req.NamespacedName, o); e != nil {
		return done(client.IgnoreNotFound(e))
	}
	status := obj(o.Object["status"])
	run := copyMap(obj(status["execution"]))
	deleting := !o.GetDeletionTimestamp().IsZero()
	if deleting && !slices.Contains(o.GetFinalizers(), finalizer) {
		return done(nil)
	}
	d, e := Intent(o)
	s, se := specOf(o)
	if e == nil {
		e = se
	}
	var interval time.Duration
	if e == nil && d["operation"] != "retain" && s.Reconciliation.DriftCheckInterval != "" {
		interval, e = duration(s.Reconciliation.DriftCheckInterval)
	}
	if e == nil && deleting && status["deletionIntent"] != nil && !equal(status["deletionIntent"], d) {
		e = fmt.Errorf("intenção de exclusão diverge do snapshot")
	}
	active := len(run) > 0 && str(run["finishedAt"]) == ""
	desiredRef := ""
	if s.WorkflowRef != nil {
		desiredRef = s.WorkflowRef.Name
	}
	desiredName, desiredEngine, desiredErr := r.resolveEngine(desiredRef)
	// Retain removes only the finalizer; it does not require an executor.
	if !active && e == nil && d["operation"] != string(execengine.Retain) && desiredErr != nil {
		return done(r.report(ctx, o, "ReconcileBlocked", "UnknownWorkflow", "Integração solicitada não cadastrada ou inválida.", M{}))
	}
	engine := desiredEngine
	selectedName := desiredName
	requiresAuthorization := true
	if active {
		selectedName = str(run["workflowRef"])
		if selectedName == "" {
			selectedName = str(obj(obj(run["intent"])["workflowRef"])["name"])
		}
		// Legacy attempts without an alias can only use the current default
		// and must still match their persisted integration snapshot.
		var selectionErr error
		selectedName, engine, selectionErr = r.resolveEngine(selectedName)
		if selectionErr != nil {
			return done(r.report(ctx, o, "ReconcileBlocked", "IntegrationChanged", "Integração da tentativa indisponível; restaurar configuração anterior.", M{}))
		}
		requiresAuthorization = authorizationRequired(engine.Snapshot())
	}
	applying := claimed(status, run)
	// Without the handshake there is no reliable apply-start boundary.
	// Never terminate an existing execution solely because the intent changed.
	canCancel := requiresAuthorization && !applying
	var request execengine.Request
	var w *execengine.Run
	if active {
		if saved := obj(run["integration"]); len(saved) > 0 && !equal(saved, engine.Snapshot()) {
			return done(r.report(ctx, o, "ReconcileBlocked", "IntegrationChanged", "Contrato alterado durante execução; restaurar configuração anterior.", M{}))
		}
		var err error
		request, err = requestFor(run)
		if err != nil {
			return done(err)
		}
		w, err = engine.Get(ctx, request)
		if err != nil {
			return done(err)
		}
		if w != nil && !w.Valid {
			if canCancel {
				delete(run, "authorization")
			}
			return done(r.report(ctx, o, "ReconcileBlocked", "InvalidExecution", "A identidade da execução diverge da solicitação.", M{"execution": run}))
		}
	}

	if e != nil && !(active && !requiresAuthorization && w != nil) {
		if active && w != nil && !w.Phase.Terminal() && canCancel {
			// Persistir cancelamento antes de encerrar o Workflow: disputa CAS com claim.
			if !yes(run["cancelling"]) {
				run["cancelling"] = true
				delete(run, "authorization")
				return again(r.report(ctx, o, "ReconcileBlocked", "InvalidIntent", e.Error(), M{"execution": run}))
			}
			if err := engine.Cancel(ctx, request); err != nil {
				return done(err)
			}
		}
		return done(r.report(ctx, o, "ReconcileBlocked", "InvalidIntent", e.Error(), M{}))
	}
	if e == nil && deleting && status["deletionIntent"] == nil {
		status = copyMap(status)
		status["deletionIntent"] = d
		o.Object["status"] = status
		return again(r.Status().Update(ctx, o))
	}
	if e == nil && !deleting && !slices.Contains(o.GetFinalizers(), finalizer) {
		o.SetFinalizers(append(o.GetFinalizers(), finalizer))
		return again(r.Update(ctx, o))
	}
	// Leitura direta das dependências antes de autorizar; watches continuam event-driven.
	all := stackList()
	if e := r.Reader.List(ctx, all, client.InNamespace(r.StackNamespace)); e != nil {
		return done(e)
	}
	dependency := dependencyReason(all.Items, o, s, str(d["operation"]))
	eligible := dependency == ""
	changed := !equal(d, run["intent"]) || desiredErr != nil || (str(run["workflowRef"]) != "" && desiredName != str(run["workflowRef"]))
	suspended := s.Suspend
	if active {
		switch {
		case w == nil:
			if yes(run["created"]) || yes(run["launchAttempted"]) || applying {
				return done(r.report(ctx, o, "Failed", "ExecutionMissing", "Execução removida; investigar Pods e state.", M{}))
			}
			if !changed && !suspended && eligible && !yes(run["cancelling"]) {
				if !requiresAuthorization {
					// Write ahead before submitting: if submission is ambiguous or
					// the Workflow disappears, fail closed instead of replaying it.
					run["launchAttempted"] = true
					if err := r.report(ctx, o, "Reconciling", "Submitting", request.Workflow, M{"execution": run}); err != nil {
						return done(err)
					}
				}
				err := engine.Create(ctx, request)
				if apierrors.IsAlreadyExists(err) {
					return again(nil)
				}
				return again(err)
			}
		case !w.Phase.Terminal() && (changed || suspended || !eligible || yes(run["cancelling"])) && canCancel:
			if !yes(run["cancelling"]) {
				run["created"], run["cancelling"] = true, true
				delete(run, "authorization")
				return again(r.report(ctx, o, "ReconcilePending", "Cancelling", request.Workflow, M{"execution": run}))
			}
			return done(engine.Cancel(ctx, request))
		}

		contractError := w != nil && (w.ContractError || !resolutionCurrent(run, w))
		if w != nil && !w.Phase.Terminal() && contractError {
			if canCancel && !yes(run["cancelling"]) {
				run["created"], run["cancelling"], run["invalidContract"] = true, true, true
				delete(run, "authorization")
				return again(r.report(ctx, o, "ReconcileBlocked", "InvalidExecutionContract", request.Workflow, M{"execution": run}))
			}
			if canCancel {
				return done(engine.Cancel(ctx, request))
			}
			if !requiresAuthorization {
				run["invalidContract"] = true
				return done(r.report(ctx, o, "ReconcileBlocked", "InvalidExecutionContract", request.Workflow, M{"execution": run}))
			}
		}
		values := M{"execution": run, "result": diagnostics(w)}
		if w != nil {
			run["created"] = true
		}
		if w != nil && !contractError && w.Resolution != nil {
			resolved := M{"reference": w.Resolution.Reference, "binding": w.Resolution.Binding}
			run["resolvedSource"] = resolved
			values["resolvedSource"] = resolved
		}
		if w != nil && w.Progress != nil && w.Progress.DriftDetected && !yes(run["drift"]) && !contractError {
			run["drift"] = true
			values["lastDriftTime"] = r.stamp()
			if !w.Phase.Terminal() {
				return again(r.report(ctx, o, "DriftDetected", "ExternalChange", request.Workflow, values))
			}
		}
		if w != nil && !w.Phase.Terminal() {
			phase, kind, reason := w.Phase, "Reconciling", "Running"
			switch {
			case applying:
				phase, reason = execengine.Applying, "Applying"
			case requiresAuthorization && !contractError && !changed && !suspended && eligible && w.Phase == execengine.AuthorizationRequested && w.Authorization != "":
				run["authorization"] = w.Authorization
				phase, reason = execengine.AuthorizationRequested, "AuthorizationRequested"
			case phase == execengine.Waiting:
				delete(run, "authorization")
				kind, reason = "ExecutionWaiting", "Waiting"
			default:
				delete(run, "authorization")
			}
			run["phase"] = string(phase)
			return done(r.report(ctx, o, kind, reason, request.Workflow, values))
		}

		kind, reason, phase := terminalResult(run, w, applying, contractError)
		run["finishedAt"], run["result"], run["reason"], run["phase"] = r.stamp(), kind, reason, string(phase)
		values["lastCheckTime"] = r.stamp()
		if changed {
			kind, reason = "ReconcilePending", "PreviousExecutionFinished"
		}
		return again(r.report(ctx, o, kind, reason, request.Workflow, values))
	}

	if d["operation"] == "retain" {
		o.SetFinalizers(slices.DeleteFunc(o.GetFinalizers(), func(v string) bool { return v == finalizer }))
		return done(r.Update(ctx, o))
	}
	if suspended {
		return done(r.report(ctx, o, "ReconcilePending", "Suspended", "Novas execuções suspensas.", M{}))
	}
	if !eligible {
		return done(r.report(ctx, o, "ReconcilePending", dependency, "Verifique o grafo e o estado das stacks do namespace.", M{}))
	}
	trigger := "GitOps"
	attempt := int64(0)
	if !changed {
		if deleting && run["result"] == "Destroyed" {
			o.SetFinalizers(slices.DeleteFunc(o.GetFinalizers(), func(v string) bool { return v == finalizer }))
			return done(r.Update(ctx, o))
		}
		finished, err := time.Parse(time.RFC3339, str(run["finishedAt"]))
		if err != nil {
			return done(err)
		}
		var delay time.Duration
		if run["result"] == "Ready" {
			delay = interval
			trigger = "DriftCheck"
		} else if run["result"] == "Failed" {
			delay, err = retryDelay(s.Reconciliation.RetryPolicy, num(run["attempt"]))
			if err != nil {
				return done(err)
			}
			trigger = "Retry"
			attempt = num(run["attempt"]) + 1
		}
		if delay == 0 {
			return done(nil)
		}
		if remaining := finished.Add(delay).Sub(r.clock()); remaining > 0 {
			return ctrl.Result{RequeueAfter: remaining}, nil
		}
	}
	next := execution(d, run, trigger, attempt)
	next["integration"] = desiredEngine.Snapshot()
	if desiredName != "" {
		next["workflowRef"] = desiredName
	}
	resolved := obj(status["resolvedSource"])
	if binding := str(resolved["binding"]); binding != "" {
		next["binding"] = binding
	}
	if !changed {
		ref := str(obj(run["resolvedSource"])["reference"])
		if ref == "" {
			ref = str(run["pinnedReference"])
		}
		if ref != "" {
			next["pinnedReference"] = ref
		}
	} else if d["operation"] == string(execengine.Destroy) {
		if ref := str(resolved["reference"]); ref != "" {
			next["pinnedReference"] = ref
		}
	}
	return again(r.report(ctx, o, "Reconciling", "Scheduling", str(next["workflow"]), M{"execution": next, "claim": nil, "result": M{}}))
}

func resolutionCurrent(run M, w *execengine.Run) bool {
	if w.Resolution != nil {
		seen := obj(run["resolvedSource"])
		if expected := str(run["binding"]); expected != "" && expected != w.Resolution.Binding {
			return false
		}
		if pinned := str(run["pinnedReference"]); pinned != "" && pinned != w.Resolution.Reference {
			return false
		}
		if len(seen) != 0 && (seen["reference"] != w.Resolution.Reference || seen["binding"] != w.Resolution.Binding) {
			return false
		}
	} else if len(obj(run["resolvedSource"])) != 0 {
		return false
	}
	token := str(run["authorization"])
	return token == "" || token == w.Authorization
}

func terminalResult(run M, w *execengine.Run, applying, contractError bool) (string, string, execengine.Phase) {
	switch {
	case contractError || yes(run["invalidContract"]):
		return "ReconcileBlocked", "InvalidExecutionContract", execengine.Blocked
	case yes(run["cancelling"]) || w == nil || (w.Phase == execengine.Cancelled && !applying):
		return "ReconcileBlocked", "Cancelled", execengine.Cancelled
	case w.Phase == execengine.Blocked && !applying:
		return "ReconcileBlocked", "ExecutionBlocked", execengine.Blocked
	case w.Phase == execengine.Succeeded && w.Result != nil:
		switch {
		case w.Result.Outcome == execengine.NoChanges && !applying && obj(run["intent"])["operation"] == string(execengine.Apply):
			return "Ready", "NoChanges", execengine.Succeeded
		case w.Result.Outcome == execengine.Verified && ((authorizationRequired(obj(run["integration"])) && applying && w.Authorization == str(run["authorization"])) || (!authorizationRequired(obj(run["integration"])) && !applying && w.Authorization == "" && w.Resolution != nil)):
			kind := "Ready"
			if obj(run["intent"])["operation"] == string(execengine.Destroy) {
				kind = "Destroyed"
			}
			return kind, "Verified", execengine.Succeeded
		}
		return "ReconcileBlocked", "InvalidExecutionContract", execengine.Blocked
	}
	return "Failed", "ExecutionFailed", execengine.Failed
}

func diagnostics(w *execengine.Run) M {
	output := M{}
	if w == nil {
		return output
	}
	details := M{}
	if w.Progress != nil {
		for k, v := range w.Progress.Details {
			details[k] = v
		}
		if w.Progress.Reason != "" {
			output["reason"] = w.Progress.Reason
		}
		if w.Progress.Message != "" {
			output["message"] = w.Progress.Message
		}
	}
	if w.Result != nil {
		output["outcome"] = string(w.Result.Outcome)
		for k, v := range w.Result.Details {
			details[k] = v
		}
		if w.Result.Reason != "" {
			output["reason"] = w.Result.Reason
		}
		if w.Result.Message != "" {
			output["message"] = w.Result.Message
		}
	}
	// O adapter valida cada envelope. Preferir o diagnóstico terminal se a
	// combinação ultrapassar o limite de campos do status.
	if len(details) > 12 {
		clear(details)
		for k, v := range w.Result.Details {
			details[k] = v
		}
	}
	if len(details) > 0 {
		output["details"] = details
	}
	return output
}
