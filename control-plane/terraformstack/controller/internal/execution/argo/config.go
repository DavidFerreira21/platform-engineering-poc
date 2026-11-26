package argo

import (
 "fmt"
 "strings"
 "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution"
 "k8s.io/apimachinery/pkg/util/validation"
)
type TemplateRef struct { Name string `json:"name"` }
type InputParameters struct {
 Intent string `json:"intent"`
 Volume string `json:"volume"`
}
type OutputParameters struct {
 Decision string `json:"decision"`
 Authorization string `json:"authorization"`
 Approval string `json:"approval"`
 Result string `json:"result"`
}
type Parameters struct {
 Input InputParameters `json:"input"`
 Output OutputParameters `json:"output"`
}
type Templates struct {
 Policy string `json:"policy"`
 Approval string `json:"approval"`
}
type WorkflowConfig struct {
 Namespace string `json:"namespace"`
 TemplateRef TemplateRef `json:"templateRef"`
 Parameters Parameters `json:"parameters"`
 Templates Templates `json:"templates"`
}

func validName(label,v string,ns bool)error{
 var errors []string
 if ns{errors=validation.IsDNS1123Label(v)}else{errors=validation.IsDNS1123Subdomain(v)}
 if len(errors)>0{return fmt.Errorf("%s: %s",label,strings.Join(errors,", "))}
 return nil
}
func distinct(label string,values ...string)error{
 seen:=map[string]bool{}
 for _,v:=range values{
  if e:=validName(label,v,false);e!=nil{return e}
  if seen[v]{return fmt.Errorf("%s: nome duplicado %q",label,v)}
  seen[v]=true
 }
 return nil
}
func(w WorkflowConfig) Validate()error{
 if e:=validName("workflow.namespace",w.Namespace,true);e!=nil{return e}
 if e:=validName("workflow.templateRef.name",w.TemplateRef.Name,false);e!=nil{return e}
 if e:=distinct("workflow.parameters.input",w.Parameters.Input.Intent,w.Parameters.Input.Volume);e!=nil{return e}
 p:=w.Parameters.Output
 if e:=distinct("workflow.parameters.output",p.Decision,p.Authorization,p.Approval,p.Result);e!=nil{return e}
 return distinct("workflow.templates",w.Templates.Policy,w.Templates.Approval)
}

func (w WorkflowConfig) Snapshot() execution.Data{
 return execution.Data{"namespace":w.Namespace,"templateRef":w.TemplateRef.Name,
 "intent":w.Parameters.Input.Intent,"volume":w.Parameters.Input.Volume,
 "decision":w.Parameters.Output.Decision,"authorization":w.Parameters.Output.Authorization,
 "approvalResult":w.Parameters.Output.Approval,"result":w.Parameters.Output.Result,
 "policyTemplate":w.Templates.Policy,"approvalTemplate":w.Templates.Approval}
}
