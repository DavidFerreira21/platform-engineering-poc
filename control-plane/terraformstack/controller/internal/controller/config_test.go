package controller

import (
	"os"
	"strings"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	b, e := os.ReadFile("../../controller-config.yaml")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ParseConfig(b); e != nil {
		t.Fatal(e)
	}
	for _, v := range []struct{ old, new string }{
		{"workers: 4", "workers: 0"},
		{"leaderElection: true", "unknown: true"},
		{"state: execution-state", "state: execution-result"},
		{"name: terraformstack", "name: ''"},
		{"serviceAccountName: terraformstack-executor", "serviceAccountName: ''"},
		{"namespace: argo-workflows", "namespace: ../argo"},
		{"workers: 4", "workers: 4\n  allowDeny: true"},
	} {
		t.Run(v.new, func(t *testing.T) {
			if _, e := ParseConfig([]byte(strings.Replace(string(b), v.old, v.new, 1))); e == nil {
				t.Fatal("configuração inválida aceita")
			}
		})
	}
}

func TestMultipleWorkflowConfig(t *testing.T) {
	b, err := os.ReadFile("../../controller-config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	base := string(b)
	second := `
  - name: alternate
    namespace: another-namespace
    templateRef:
      name: custom-template
    serviceAccountName: custom-executor
    contract:
      authorization: false
    parameters:
      input:
        intent: request
      output:
        result: result
        resolvedSource: resolution
`
	c, err := ParseConfig([]byte(base + second))
	if err != nil {
		t.Fatal(err)
	}
	name, integrations, err := c.Integrations()
	if err != nil || name != "terraformstack" || len(integrations) != 2 ||
		!integrations[name].RequiresAuthorization() || integrations["alternate"].RequiresAuthorization() {
		t.Fatalf("unexpected integration config: %s %v %v", name, integrations, err)
	}
	for _, invalid := range []string{
		strings.Replace(base+second, "defaultWorkflow: terraformstack", "defaultWorkflow: missing", 1),
		strings.Replace(base+second, "name: alternate", "name: terraformstack", 1),
		strings.Replace(base+second, "authorization: false", "authorization: true", 1),
		strings.Replace(base+second, "defaultWorkflow: terraformstack\n", "", 1),
	} {
		if _, err := ParseConfig([]byte(invalid)); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
	single := strings.Replace(base, "defaultWorkflow: terraformstack\n", "", 1)
	single = strings.Replace(single, "    contract:\n      authorization: true\n", "", 1)
	c, err = ParseConfig([]byte(single))
	if err != nil {
		t.Fatal(err)
	}
	name, integrations, err = c.Integrations()
	if err != nil || name != "terraformstack" || !integrations[name].RequiresAuthorization() {
		t.Fatal("missing authorization must default to true")
	}
}
