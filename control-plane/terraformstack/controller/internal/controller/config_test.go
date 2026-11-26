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
