// Package v1alpha1 define a identidade da API. O schema canônico permanece
// no CRD versionado; campos de execução preservam o JSON do runner.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersion = schema.GroupVersion{Group: "platform.example.io", Version: "v1alpha1"}
var StackGVK = GroupVersion.WithKind("TerraformStack")

type SourceRef struct {
	Repository string `json:"repository"`
	Path       string `json:"path"`
	Version    string `json:"version,omitempty"`
}
type Dependency struct {
	Name string `json:"name"`
}
type RetryPolicy struct {
	Strategy    string `json:"strategy,omitempty"`
	Interval    string `json:"interval,omitempty"`
	MaxInterval string `json:"maxInterval,omitempty"`
	MaxRetries  int    `json:"maxRetries,omitempty"`
}
type StackSpec struct {
	SourceRef SourceRef         `json:"sourceRef"`
	Context   map[string]string `json:"context,omitempty"`
	DependsOn []Dependency      `json:"dependsOn,omitempty"`
	Suspend   bool              `json:"suspend,omitempty"`
	Lifecycle struct {
		Destroy                    bool `json:"destroy,omitempty"`
		DestroyResourcesOnDeletion bool `json:"destroyResourcesOnDeletion,omitempty"`
	} `json:"lifecycle,omitempty"`
	Reconciliation struct {
		DriftCheckInterval string      `json:"driftCheckInterval,omitempty"`
		RetryPolicy        RetryPolicy `json:"retryPolicy,omitempty"`
	} `json:"reconciliation,omitempty"`
}
type StackStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}
