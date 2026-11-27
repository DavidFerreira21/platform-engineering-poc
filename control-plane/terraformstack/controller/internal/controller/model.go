package controller

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"time"

	api "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/api/v1alpha1"
	execengine "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const revision = "platform.example.io/revision"
const finalizer = "platform.example.io/terraformstack-cleanup"

type M = map[string]any

func obj(v any) M {
	m, _ := v.(map[string]any)
	if m == nil {
		return M{}
	}
	return m
}
func str(v any) string { s, _ := v.(string); return s }
func yes(v any) bool   { b, _ := v.(bool); return b }
func num(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	}
	return 0
}
func arr(v any) []any                         { a, _ := v.([]any); return a }
func copyMap(m M) M                           { return runtime.DeepCopyJSON(m) }
func canonical(v any) string                  { b, _ := json.Marshal(v); return string(b) }
func equal(a, b any) bool                     { return canonical(a) == canonical(b) }
func StackObject() *unstructured.Unstructured { return stackObject() }
func stackObject() *unstructured.Unstructured {
	o := &unstructured.Unstructured{}
	o.SetGroupVersionKind(api.StackGVK)
	return o
}
func stackList() *unstructured.UnstructuredList {
	o := &unstructured.UnstructuredList{}
	o.SetGroupVersionKind(api.GroupVersion.WithKind("TerraformStackList"))
	return o
}
func specOf(o *unstructured.Unstructured) (api.StackSpec, error) {
	var s api.StackSpec
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj(o.Object["spec"]), &s)
	return s, err
}

var pathPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_./-]*$`)
var durationPattern = regexp.MustCompile("^([0-9]+[smh])+$")

func duration(v string) (time.Duration, error) {
	if !durationPattern.MatchString(v) {
		return 0, fmt.Errorf("duração inválida: %s", v)
	}
	d, e := time.ParseDuration(v)
	if e != nil || d <= 0 || d > 24*time.Hour {
		return 0, fmt.Errorf("duração fora do limite: %s", v)
	}
	return d, nil
}
func fallback(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

// Intent registra a intenção declarativa. Referências e contexto são opacos.
func Intent(o *unstructured.Unstructured) (M, error) {
	s, err := specOf(o)
	if err != nil {
		return nil, err
	}
	deleting := !o.GetDeletionTimestamp().IsZero()
	op := execengine.Apply
	switch {
	case deleting && !s.Lifecycle.DestroyResourcesOnDeletion:
		op = execengine.Retain
	case deleting || s.Lifecycle.Destroy:
		op = execengine.Destroy
	}
	if op != execengine.Retain && (s.SourceRef.Repository == "" ||
		len(s.SourceRef.Path) > 512 || !pathPattern.MatchString(s.SourceRef.Path) || !fs.ValidPath(s.SourceRef.Path)) {
		return nil, fmt.Errorf("fonte ou caminho inválido")
	}
	s.SourceRef.Version = fallback(s.SourceRef.Version, "latest")
	if s.Context == nil {
		s.Context = map[string]string{}
	}
	intent := &execengine.Intent{
		Stack: o.GetName(), Namespace: o.GetNamespace(), UID: string(o.GetUID()),
		Generation: o.GetGeneration(), RevisionAnnotation: o.GetAnnotations()[revision],
		Deleting: deleting, SourceRef: s.SourceRef, Context: s.Context, Operation: op,
	}
	return runtime.DefaultUnstructuredConverter.ToUnstructured(intent)
}

func requestFor(run M) (execengine.Request, error) {
	req := execengine.Request{
		Identity:        execengine.Identity{Workflow: str(run["workflow"]), IntentHash: str(run["intentHash"])},
		PinnedReference: str(run["pinnedReference"]),
		Binding:         str(run["binding"]),
	}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj(run["intent"]), &req.Intent)
	return req, err
}

func claimed(status, run M) bool {
	claim := obj(status["claim"])
	token := str(run["authorization"])
	return token != "" && claim["authorization"] == token &&
		claim["workflow"] == run["workflow"] && claim["intentHash"] == run["intentHash"]
}

func retryDelay(p api.RetryPolicy, attempt int64) (time.Duration, error) {
	if p.Strategy == "None" || attempt >= int64(p.MaxRetries) {
		return 0, nil
	}
	d, e := duration(fallback(p.Interval, "10s"))
	if e != nil {
		return 0, e
	}
	if fallback(p.Strategy, "ExponentialBackoff") == "ExponentialBackoff" {
		max, e := duration(fallback(p.MaxInterval, "5m"))
		if e != nil {
			return 0, e
		}
		for i := int64(0); i < attempt && d < max; i++ {
			d *= 2
		}
		if d > max {
			d = max
		}
	}
	return d, nil
}
func execution(d, old M, trigger string, attempt int64) M {
	sequence := num(old["sequence"]) + 1
	sum := sha256.Sum256([]byte(canonical(d)))
	return M{"intent": d, "intentHash": fmt.Sprintf("%x", sum), "workflow": fmt.Sprintf("tf-%s-%x-%d", str(d["uid"]), sum[:6], sequence), "sequence": sequence, "attempt": attempt, "trigger": trigger, "phase": string(execengine.Pending)}
}
func conditionTrue(o *unstructured.Unstructured, kind string) bool {
	status := obj(o.Object["status"])
	for _, v := range arr(status["conditions"]) {
		c := obj(v)
		if str(c["type"]) == kind && str(c["status"]) == "True" && num(c["observedGeneration"]) == o.GetGeneration() {
			return true
		}
	}
	return false
}

// dependencyReason usa apenas o grafo do namespace e as intenções vigentes.
func dependencyReason(all []unstructured.Unstructured, o *unstructured.Unstructured, spec api.StackSpec, op string) string {
	if op == string(execengine.Retain) {
		return ""
	}
	type entry struct {
		object *unstructured.Unstructured
		spec   api.StackSpec
	}
	nodes := make(map[string]entry, len(all))
	for i := range all {
		item := &all[i]
		if item.GetNamespace() == o.GetNamespace() {
			if s, err := specOf(item); err == nil {
				nodes[item.GetName()] = entry{item, s}
			}
		}
	}
	visiting, visited := map[string]bool{}, map[string]bool{}
	var cycle func(string) bool
	cycle = func(name string) bool {
		if visiting[name] {
			return true
		}
		if visited[name] {
			return false
		}
		visiting[name] = true
		for _, dep := range nodes[name].spec.DependsOn {
			if cycle(dep.Name) {
				return true
			}
		}
		delete(visiting, name)
		visited[name] = true
		return false
	}
	if cycle(o.GetName()) {
		return "DependencyCycle"
	}
	current := func(other *unstructured.Unstructured, condition string) bool {
		status := obj(other.Object["status"])
		wanted, err := Intent(other)
		return err == nil && num(status["observedGeneration"]) == other.GetGeneration() &&
			equal(obj(status["execution"])["intent"], wanted) && conditionTrue(other, condition)
	}
	if op == string(execengine.Apply) {
		for _, dep := range spec.DependsOn {
			other := nodes[dep.Name].object
			if other == nil || !other.GetDeletionTimestamp().IsZero() || !current(other, "Ready") {
				return "DependencyBlocked"
			}
		}
	}
	if op == string(execengine.Destroy) {
		for _, node := range nodes {
			for _, dep := range node.spec.DependsOn {
				if dep.Name == o.GetName() && !current(node.object, "Destroyed") {
					return "DependencyBlocked"
				}
			}
		}
	}
	return ""
}
