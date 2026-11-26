// Package execution defines the internal boundary between reconciliation and
// execution. It does not define a plugin API or a new Kubernetes resource.
package execution

import "context"

type Data = map[string]any

// Run is a normalized observation, not an Argo API object.
// A nil observation means the execution does not exist.
type Run struct {
	Name      string
	Terminal  bool
	Succeeded bool
	Valid     bool
	Result    Data
}

type Engine interface {
	Snapshot() Data
	Validate() error
	Create(context.Context, Data) error
	Get(context.Context, string, string) (*Run, error)
	Cancel(context.Context, string, string) error
}
