// Package observability exposes operational signals without owning reconciliation.
package observability

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Data = map[string]any

func object(v any) Data        { m, _ := v.(map[string]any); return m }
func stringValue(v any) string { s, _ := v.(string); return s }
func boolean(v any) bool       { b, _ := v.(bool); return b }
func integer(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	}
	return 0
}

func timestamp(v any) float64 {
	t, err := time.Parse(time.RFC3339Nano, stringValue(v))
	if err != nil {
		return 0
	}
	return float64(t.Unix())
}

func currentCondition(status Data) (string, string) {
	conditions, _ := status["conditions"].([]any)
	for _, entry := range conditions {
		c := object(entry)
		if c["status"] == "True" {
			return boundedCondition(stringValue(c["type"])), boundedReason(stringValue(c["reason"]))
		}
	}
	return "Unknown", "Unknown"
}

func boundedCondition(s string) string {
	switch s {
	case "Ready", "Destroyed", "Failed", "ReconcileBlocked", "ReconcilePending", "Reconciling", "ExecutionWaiting", "DriftDetected":
		return s
	default:
		return "Unknown"
	}
}

func boundedReason(s string) string {
	switch s {
	case "Scheduling", "Submitting", "Running", "Waiting", "AuthorizationRequested", "Applying",
		"Cancelling", "Cancelled", "Verified", "NoChanges", "ExecutionFailed", "ExecutionBlocked",
		"InvalidExecutionContract", "InvalidExecution", "InvalidIntent", "IntegrationChanged",
		"UnknownWorkflow", "ExecutionMissing", "PreviousExecutionFinished", "Suspended",
		"DependencyBlocked", "DependencyCycle", "ExternalChange":
		return s
	default:
		return "Unknown"
	}
}

func boundedPhase(s string) string {
	switch s {
	case "Pending", "Running", "Waiting", "AuthorizationRequested", "Applying", "Succeeded", "Failed", "Blocked", "Cancelled":
		return s
	default:
		return "Unknown"
	}
}

func boundedTrigger(s string) string {
	switch s {
	case "GitOps", "DriftCheck", "Retry":
		return s
	default:
		return "Unknown"
	}
}

// EnrichStatus only persists facts derived from existing status transitions.
// Unknown drift remains unknown, rather than implying that the target is clean.
func EnrichStatus(old, next Data, now time.Time) {
	run := object(next["execution"])
	previous := object(old["execution"])
	stamp := now.UTC().Format(time.RFC3339Nano)
	if len(run) != 0 && run["workflow"] != previous["workflow"] {
		run["scheduledAt"] = stamp
	}
	if boolean(run["drift"]) && (run["workflow"] != previous["workflow"] || !boolean(previous["drift"])) {
		next["driftPending"] = true
	}
	if stringValue(run["finishedAt"]) != "" && (run["workflow"] != previous["workflow"] || stringValue(previous["finishedAt"]) == "") {
		if run["result"] == "Ready" {
			next["lastSuccessTime"] = stamp
			// Only a valid successful result confirms convergence. An absent
			// drift signal alone never establishes a clean target.
			if boolean(next["driftPending"]) {
				next["driftPending"] = false
			}
		}
		if run["result"] == "Destroyed" {
			delete(next, "driftPending")
		}
	}
}

type Health struct {
	Namespace, Stack, Condition, Reason, Phase, Trigger   string
	Active, Suspended, Current, DriftKnown, DriftPending  bool
	Attempt, Scheduled, LastCheck, LastSuccess, LastDrift float64
}

func ReadHealth(o *unstructured.Unstructured) Health {
	status := object(o.Object["status"])
	run := object(status["execution"])
	condition, reason := currentCondition(status)
	drift, known := status["driftPending"].(bool)
	return Health{
		Namespace: o.GetNamespace(), Stack: o.GetName(), Condition: condition, Reason: reason,
		Phase: boundedPhase(stringValue(run["phase"])), Trigger: boundedTrigger(stringValue(run["trigger"])),
		Active:     stringValue(run["workflow"]) != "" && stringValue(run["finishedAt"]) == "",
		Suspended:  boolean(object(o.Object["spec"])["suspend"]),
		Current:    integer(status["observedGeneration"]) == o.GetGeneration() && condition != "Unknown",
		DriftKnown: known, DriftPending: drift, Attempt: float64(integer(run["attempt"])),
		Scheduled: timestamp(run["scheduledAt"]), LastCheck: timestamp(status["lastCheckTime"]),
		LastSuccess: timestamp(status["lastSuccessTime"]), LastDrift: timestamp(status["lastDriftTime"]),
	}
}
