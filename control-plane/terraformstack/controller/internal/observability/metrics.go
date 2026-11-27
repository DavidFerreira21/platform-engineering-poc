package observability

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HealthCollector reads the existing informer cache at scrape time. No new
// watches, external collector, API calls per stack, or mutable gauge cache.
type HealthCollector struct {
	reader client.Reader
	namespace string
	descriptors map[string]*prometheus.Desc
}

func NewHealthCollector(reader client.Reader, namespace string) *HealthCollector {
	c := &HealthCollector{reader: reader, namespace: namespace, descriptors: map[string]*prometheus.Desc{}}
	add := func(name, help string, labels ...string) {
		c.descriptors[name] = prometheus.NewDesc("terraformstack_"+name, help, labels, nil)
	}
	add("resource_health_available", "Whether the resource health cache was readable.")
	add("condition", "Current persisted condition and bounded reason.", "namespace", "stack", "condition", "reason")
	add("execution_phase", "Current persisted execution phase.", "namespace", "stack", "phase")
	add("execution_active", "One if an unfinished execution is recorded.", "namespace", "stack")
	add("suspended", "One if the stack is suspended.", "namespace", "stack")
	add("observed_generation_current", "One if status observes the current generation.", "namespace", "stack")
	add("drift_pending", "Reported drift not yet confirmed converged; absent if unknown.", "namespace", "stack")
	add("last_check_timestamp_seconds", "Last completed execution timestamp.", "namespace", "stack")
	add("last_success_timestamp_seconds", "Last confirmed successful apply reconciliation.", "namespace", "stack")
	add("last_drift_timestamp_seconds", "Last reported drift timestamp.", "namespace", "stack")
	add("execution_scheduled_timestamp_seconds", "Time the current attempt was scheduled.", "namespace", "stack")
	add("retry_attempts", "Retry attempt number for the recorded execution.", "namespace", "stack")
	add("active_executions", "Number of stacks with unfinished executions.")
	add("blocked_executions", "Number of stacks with a current blocked condition.")
	return c
}

func (c *HealthCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptors {
		ch <- d
	}
}

func (c *HealthCollector) Collect(ch chan<- prometheus.Metric) {
	emit := func(name string, value float64, labels ...string) {
		ch <- prometheus.MustNewConstMetric(c.descriptors[name], prometheus.GaugeValue, value, labels...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: "platform.example.io", Version: "v1alpha1", Kind: "TerraformStackList"})
	if err := c.reader.List(ctx, list, client.InNamespace(c.namespace)); err != nil {
		emit("resource_health_available", 0)
		return
	}
	emit("resource_health_available", 1)
	active, blocked := 0.0, 0.0
	for i := range list.Items {
		h := ReadHealth(&list.Items[i])
		labels := []string{h.Namespace, h.Stack}
		emit("condition", 1, h.Namespace, h.Stack, h.Condition, h.Reason)
		emit("execution_phase", 1, h.Namespace, h.Stack, h.Phase)
		emit("execution_active", flag(h.Active), labels...)
		emit("suspended", flag(h.Suspended), labels...)
		emit("observed_generation_current", flag(h.Current), labels...)
		if h.DriftKnown {
			emit("drift_pending", flag(h.DriftPending), labels...)
		}
		for name, value := range map[string]float64{
			"last_check_timestamp_seconds": h.LastCheck, "last_success_timestamp_seconds": h.LastSuccess,
			"last_drift_timestamp_seconds": h.LastDrift, "execution_scheduled_timestamp_seconds": h.Scheduled,
		} {
			if value > 0 {
				emit(name, value, labels...)
			}
		}
		emit("retry_attempts", h.Attempt, labels...)
		active += flag(h.Active)
		blocked += flag(h.Condition == "ReconcileBlocked" && h.Current)
	}
	emit("active_executions", active)
	emit("blocked_executions", blocked)
}

func flag(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

type Metrics struct {
	executions *prometheus.CounterVec
	duration *prometheus.HistogramVec
	contractErrors, drift, retries, authorizations *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	counter := func(name, help string, labels ...string) *prometheus.CounterVec {
		c := prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: "terraformstack", Name: name, Help: help}, labels)
		reg.MustRegister(c)
		return c
	}
	m := &Metrics{
		executions: counter("executions_total", "Observed persisted terminal execution transitions.", "phase", "trigger"),
		contractErrors: counter("contract_errors_total", "Observed transitions to invalid contract or identity.", "reason"),
		drift: counter("drift_detected_total", "Observed newly reported drift.", "trigger"),
		retries: counter("retries_total", "Observed newly scheduled retry attempts.", "trigger"),
		authorizations: counter("authorization_requests_total", "Observed authorization grants.", "trigger"),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: "terraformstack", Name: "execution_duration_seconds", Help: "Time from scheduling to terminal result, including waits.", Buckets: []float64{1, 5, 15, 30, 60, 300, 900, 3600, 14400, 86400}}, []string{"phase"}),
	}
	reg.MustRegister(m.duration)
	return m
}

func (m *Metrics) Observe(old, next Data) {
	if m == nil {
		return
	}
	before, after := object(old["execution"]), object(next["execution"])
	newRun := after["workflow"] != before["workflow"] && stringValue(after["workflow"]) != ""
	trigger := boundedTrigger(stringValue(after["trigger"]))
	phase := boundedPhase(stringValue(after["phase"]))
	if newRun && trigger == "Retry" {
		m.retries.WithLabelValues(trigger).Inc()
	}
	if boolean(after["drift"]) && (newRun || !boolean(before["drift"])) {
		m.drift.WithLabelValues(trigger).Inc()
	}
	if stringValue(after["authorization"]) != "" && (newRun || before["authorization"] != after["authorization"]) {
		m.authorizations.WithLabelValues(trigger).Inc()
	}
	_, previousReason := currentCondition(old)
	_, reason := currentCondition(next)
	if (reason == "InvalidExecutionContract" || reason == "InvalidExecution") && (newRun || previousReason != reason) {
		m.contractErrors.WithLabelValues(reason).Inc()
	}
	if stringValue(after["finishedAt"]) != "" && (newRun || stringValue(before["finishedAt"]) == "") {
		m.executions.WithLabelValues(phase, trigger).Inc()
		start, end := timestamp(after["scheduledAt"]), timestamp(after["finishedAt"])
		if start > 0 && end >= start {
			m.duration.WithLabelValues(phase).Observe(end - start)
		}
	}
}
