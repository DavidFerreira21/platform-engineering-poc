// Package execution define a comunicação interna entre reconciliação e execução.
package execution

import (
	"context"

	api "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/api/v1alpha1"
)

type Data = map[string]any

type Intent struct {
	Stack              string            `json:"stack"`
	Namespace          string            `json:"namespace"`
	UID                string            `json:"uid"`
	Generation         int64             `json:"generation"`
	RevisionAnnotation string            `json:"revisionAnnotation"`
	Deleting           bool              `json:"deleting"`
	WorkflowRef        *api.WorkflowRef  `json:"workflowRef,omitempty"`
	SourceRef          api.SourceRef     `json:"sourceRef"`
	Context            map[string]string `json:"context"`
	Operation          Operation         `json:"operation"`
}

type Operation string

const (
	Apply   Operation = "apply"
	Destroy Operation = "destroy"
	Retain  Operation = "retain"
)

type Identity struct {
	Workflow   string `json:"workflow"`
	IntentHash string `json:"intentHash"`
}

type Request struct {
	Identity
	Intent          Intent `json:"intent"`
	PinnedReference string `json:"pinnedReference,omitempty"`
	Binding         string `json:"binding,omitempty"`
}

type Resolution struct {
	Reference string `json:"reference"`
	Binding   string `json:"binding"`
}

type Phase string

const (
	Pending                Phase = "Pending"
	Running                Phase = "Running"
	Waiting                Phase = "Waiting"
	AuthorizationRequested Phase = "AuthorizationRequested"
	Applying               Phase = "Applying"
	Succeeded              Phase = "Succeeded"
	Failed                 Phase = "Failed"
	Blocked                Phase = "Blocked"
	Cancelled              Phase = "Cancelled"
)

func (p Phase) Terminal() bool {
	return p == Succeeded || p == Failed || p == Blocked || p == Cancelled
}

type Outcome string

const (
	NoChanges        Outcome = "NoChanges"
	Verified         Outcome = "Verified"
	OutcomeBlocked   Outcome = "Blocked"
	OutcomeCancelled Outcome = "Cancelled"
)

type Diagnostics struct {
	Reason  string            `json:"reason,omitempty"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

type Progress struct {
	Identity
	Phase         Phase `json:"phase"`
	DriftDetected bool  `json:"driftDetected,omitempty"`
	Diagnostics
}

type Authorization struct {
	Identity
	Token string `json:"token"`
}

type Result struct {
	Identity
	Outcome Outcome `json:"outcome"`
	Diagnostics
}

type ResolvedSource struct {
	Identity
	Resolution
}

// Run é a observação do adapter. A fase Applying só é registrada a partir do claim na API.
type Run struct {
	Phase         Phase
	Valid         bool
	ContractError bool
	Resolution    *Resolution
	Authorization string
	Progress      *Progress
	Result        *Result
}

type Engine interface {
	Snapshot() Data
	Validate() error
	Create(context.Context, Request) error
	Get(context.Context, Request) (*Run, error)
	Cancel(context.Context, Request) error
}
