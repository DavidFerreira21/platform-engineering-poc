package observability

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Observer struct {
	Metrics  *Metrics
	Recorder record.EventRecorder
}

// Committed runs only after a successful status CAS. No payloads, tokens,
// free-form diagnostics, source references or state are emitted.
func (o *Observer) Committed(ctx context.Context, stack *unstructured.Unstructured, old, next Data) {
	run, previous := object(next["execution"]), object(old["execution"])
	condition, reason := currentCondition(next)
	oldCondition, oldReason := currentCondition(old)
	newRun := run["workflow"] != previous["workflow"]
	transition := newRun || condition != oldCondition || reason != oldReason || run["phase"] != previous["phase"]
	if transition {
		log.FromContext(ctx).Info("TerraformStack status transition",
			"component", "controller", "event", "stack.status.updated",
			"workflowRef", stringValue(run["workflowRef"]),
			"workflow", stringValue(run["workflow"]), "intentHash", stringValue(run["intentHash"]),
			"authorizationRequired", object(run["integration"])["authorizationRequired"] != false,
			"phase", boundedPhase(stringValue(run["phase"])), "trigger", boundedTrigger(stringValue(run["trigger"])),
			"attempt", integer(run["attempt"]), "condition", condition, "reason", reason)
	}
	if o == nil {
		return
	}
	o.Metrics.Observe(old, next)
	if o.Recorder == nil || !transition || reason == "Unknown" {
		return
	}
	kind := "Normal"
	if condition == "Failed" || condition == "ReconcileBlocked" || condition == "DriftDetected" {
		kind = "Warning"
	}
	o.Recorder.Eventf(stack, kind, reason, "TerraformStack condition %s; execution phase %s", condition, boundedPhase(stringValue(run["phase"])))
}
