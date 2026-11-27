package argo

import (
	"fmt"
	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

type TemplateRef struct {
	Name string `json:"name"`
}
type InputParameters struct {
	Intent string `json:"intent"`
}
type OutputParameters struct {
	State          string `json:"state"`
	Authorization  string `json:"authorization"`
	Result         string `json:"result"`
	ResolvedSource string `json:"resolvedSource"`
}
type Parameters struct {
	Input  InputParameters  `json:"input"`
	Output OutputParameters `json:"output"`
}
type ContractConfig struct {
	Authorization *bool `json:"authorization,omitempty"`
}

type WorkflowConfig struct {
	Contract ContractConfig `json:"contract,omitempty"`
	Namespace          string      `json:"namespace"`
	TemplateRef        TemplateRef `json:"templateRef"`
	ServiceAccountName string      `json:"serviceAccountName"`
	Parameters         Parameters  `json:"parameters"`
}

func validName(label, v string, ns bool) error {
	var errors []string
	if ns {
		errors = validation.IsDNS1123Label(v)
	} else {
		errors = validation.IsDNS1123Subdomain(v)
	}
	if len(errors) > 0 {
		return fmt.Errorf("%s: %s", label, strings.Join(errors, ", "))
	}
	return nil
}
func distinct(label string, values ...string) error {
	seen := map[string]bool{}
	for _, v := range values {
		if e := validName(label, v, false); e != nil {
			return e
		}
		if seen[v] {
			return fmt.Errorf("%s: nome duplicado %q", label, v)
		}
		seen[v] = true
	}
	return nil
}
func (w WorkflowConfig) RequiresAuthorization() bool {
	return w.Contract.Authorization == nil || *w.Contract.Authorization
}

func (w WorkflowConfig) Validate() error {
	if e := validName("workflow.namespace", w.Namespace, true); e != nil {
		return e
	}
	if e := validName("workflow.templateRef.name", w.TemplateRef.Name, false); e != nil {
		return e
	}
	if e := validName("workflow.serviceAccountName", w.ServiceAccountName, false); e != nil {
		return e
	}
	if e := validName("workflow.parameters.input.intent", w.Parameters.Input.Intent, false); e != nil {
		return e
	}
	p := w.Parameters.Output
	outputs := []string{p.Result, p.ResolvedSource}
	if p.State != "" {
		outputs = append(outputs, p.State)
	}
	if w.RequiresAuthorization() {
		outputs = append(outputs, p.Authorization)
	} else if p.Authorization != "" {
		return fmt.Errorf("authorization output must be omitted when authorization is disabled")
	}
	return distinct("workflow.parameters.output", outputs...)
}

func (w WorkflowConfig) Snapshot() execution.Data {
	snapshot := execution.Data{
		"namespace": w.Namespace, "templateRef": w.TemplateRef.Name, "serviceAccountName": w.ServiceAccountName,
		"intent": w.Parameters.Input.Intent, "state": w.Parameters.Output.State,
		"authorization": w.Parameters.Output.Authorization, "result": w.Parameters.Output.Result,
		"resolvedSource": w.Parameters.Output.ResolvedSource,
	}
	if !w.RequiresAuthorization() {
		snapshot["authorizationRequired"] = false
	}
	return snapshot
}
