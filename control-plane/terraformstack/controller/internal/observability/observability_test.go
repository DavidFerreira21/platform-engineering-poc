package observability

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDriftLifecycle(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first := Data{"execution": Data{"workflow": "one", "phase": "Running"}}
	EnrichStatus(Data{}, first, now)
	if first["driftPending"] != nil || object(first["execution"])["scheduledAt"] == nil {
		t.Fatal(first)
	}
	drift := Data{"execution": Data{"workflow": "one", "drift": true, "phase": "Running"}}
	EnrichStatus(first, drift, now)
	if drift["driftPending"] != true {
		t.Fatal(drift)
	}
	failed := Data{"driftPending": true, "execution": Data{"workflow": "one", "drift": true, "finishedAt": now.Format(time.RFC3339), "result": "Failed"}}
	EnrichStatus(drift, failed, now)
	if failed["driftPending"] != true || failed["lastSuccessTime"] != nil {
		t.Fatal(failed)
	}
	success := Data{"driftPending": true, "execution": Data{"workflow": "two", "finishedAt": now.Format(time.RFC3339), "result": "Ready"}}
	EnrichStatus(failed, success, now)
	if success["driftPending"] != false || success["lastSuccessTime"] == nil {
		t.Fatal(success)
	}
}

func TestMetricsTransitions(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	before := Data{"execution": Data{"workflow": "one", "phase": "Running"}}
	after := Data{"execution": Data{"workflow": "one", "phase": "Succeeded", "trigger": "GitOps", "finishedAt": "2026-01-01T00:01:00Z", "scheduledAt": "2026-01-01T00:00:00Z"}}
	m.Observe(before, after)
	m.Observe(after, after)
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, family := range families {
		if family.GetName() == "terraformstack_executions_total" {
			found = true
			if family.Metric[0].GetCounter().GetValue() != 1 {
				t.Fatal("duplicate terminal transition")
			}
		}
	}
	if !found {
		t.Fatal("missing counter")
	}
}

func TestResourceHealthAndDeletion(t *testing.T) {
	o := &unstructured.Unstructured{Object: Data{
		"apiVersion": "platform.example.io/v1alpha1", "kind": "TerraformStack",
		"metadata": Data{"name": "network", "namespace": "platform-system", "generation": int64(2)},
		"status":   Data{"observedGeneration": int64(2), "driftPending": true, "conditions": []any{Data{"type": "Failed", "status": "True", "reason": "ExecutionFailed"}}},
	}}
	h := ReadHealth(o)
	if h.Condition != "Failed" || !h.Current || !h.DriftKnown || !h.DriftPending {
		t.Fatal(h)
	}
	reader := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithObjects(o).Build()
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewHealthCollector(reader, "platform-system"))
	check := func(want bool) {
		t.Helper()
		families, err := reg.Gather()
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, family := range families {
			if family.GetName() == "terraformstack_drift_pending" {
				found = true
			}
		}
		if found != want {
			t.Fatal("unexpected drift series")
		}
	}
	check(true)
	if err := reader.Delete(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	check(false)
}

func TestEventsNoPayloadOrDuplicates(t *testing.T) {
	recorder := record.NewFakeRecorder(5)
	observer := &Observer{Recorder: recorder}
	o := &unstructured.Unstructured{}
	o.SetName("network")
	old := Data{}
	next := Data{"conditions": []any{Data{"type": "Failed", "status": "True", "reason": "ExecutionFailed", "message": "private-payload"}}}
	observer.Committed(context.Background(), o, old, next)
	observer.Committed(context.Background(), o, next, next)
	select {
	case event := <-recorder.Events:
		if !strings.Contains(event, "Warning ExecutionFailed") || strings.Contains(event, "private-payload") {
			t.Fatal(event)
		}
	default:
		t.Fatal("missing event")
	}
	select {
	case event := <-recorder.Events:
		t.Fatal("duplicate event", event)
	default:
	}
}
