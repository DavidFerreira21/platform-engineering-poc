package controller

import (
	"fmt"
	"os"
	"strings"

	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution/argo"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

// Config é o contrato de instalação, mantido pelo Platform Team.
type Config struct {
	Controller      ControllerConfig    `json:"controller"`
	Workflow        argo.WorkflowConfig `json:"workflow,omitempty"`
	DefaultWorkflow string              `json:"defaultWorkflow,omitempty"`
	Workflows       []NamedWorkflow     `json:"workflows,omitempty"`
}
type NamedWorkflow struct {
	Name string `json:"name"`
	argo.WorkflowConfig
}
type ControllerConfig struct {
	StackNamespace string `json:"stackNamespace"`
	Workers        int    `json:"workers"`
	LeaderElection *bool  `json:"leaderElection"`
}

func LoadConfig(path string) (Config, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return Config{}, e
	}
	return ParseConfig(b)
}
func ParseConfig(b []byte) (Config, error) {
	var c Config
	if e := yaml.UnmarshalStrict(b, &c); e != nil {
		return c, e
	}
	if e := c.Validate(); e != nil {
		return c, e
	}
	return c, nil
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
func (c Config) Validate() error {
	if e := validName("controller.stackNamespace", c.Controller.StackNamespace, true); e != nil {
		return e
	}
	if c.Controller.Workers < 1 || c.Controller.Workers > 32 {
		return fmt.Errorf("workers deve estar entre 1 e 32")
	}
	if c.Controller.LeaderElection == nil {
		return fmt.Errorf("leaderElection obrigatório")
	}
	_, _, err := c.Integrations()
	return err
}

// Integrations resolves the installation allowlist and its default alias.
func (c Config) Integrations() (string, map[string]argo.WorkflowConfig, error) {
	if len(c.Workflows) == 0 {
		if err := c.Workflow.Validate(); err != nil {
			return "", nil, err
		}
		name := c.Workflow.TemplateRef.Name
		if c.DefaultWorkflow != "" && c.DefaultWorkflow != name {
			return "", nil, fmt.Errorf("defaultWorkflow não cadastrado")
		}
		return name, map[string]argo.WorkflowConfig{name: c.Workflow}, nil
	}
	if c.Workflow != (argo.WorkflowConfig{}) {
		return "", nil, fmt.Errorf("workflow e workflows são mutuamente exclusivos")
	}
	configs := make(map[string]argo.WorkflowConfig, len(c.Workflows))
	for _, w := range c.Workflows {
		if err := validName("workflows.name", w.Name, false); err != nil {
			return "", nil, err
		}
		if _, exists := configs[w.Name]; exists {
			return "", nil, fmt.Errorf("workflow duplicado: %s", w.Name)
		}
		if err := w.Validate(); err != nil {
			return "", nil, fmt.Errorf("workflow %s: %w", w.Name, err)
		}
		configs[w.Name] = w.WorkflowConfig
	}
	name := c.DefaultWorkflow
	if name == "" && len(c.Workflows) == 1 {
		name = c.Workflows[0].Name
	}
	if _, exists := configs[name]; !exists {
		return "", nil, fmt.Errorf("defaultWorkflow deve identificar uma integração cadastrada")
	}
	return name, configs, nil
}
