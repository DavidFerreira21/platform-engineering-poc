package controller

import (
 "fmt"
 "os"
 "strings"

 "k8s.io/apimachinery/pkg/util/validation"
 "sigs.k8s.io/yaml"
 "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution/argo"
)

// Config é o contrato de instalação. Não permite alterar a semântica do
// lifecycle, dos resultados OPA ou do handshake de autorização.
type Config struct {
 Controller ControllerConfig `json:"controller"`
 Workflow argo.WorkflowConfig `json:"workflow"`
}
type ControllerConfig struct {
 StackNamespace string `json:"stackNamespace"`
 Workers int `json:"workers"`
 LeaderElection *bool `json:"leaderElection"`
}
func LoadConfig(path string) (Config,error) {
 b,e:=os.ReadFile(path);if e!=nil{return Config{},e}
 return ParseConfig(b)
}
func ParseConfig(b []byte)(Config,error){
 var c Config
 if e:=yaml.UnmarshalStrict(b,&c);e!=nil{return c,e}
 if e:=c.Validate();e!=nil{return c,e}
 return c,nil
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
func(c Config) Validate()error{
 if e:=validName("controller.stackNamespace",c.Controller.StackNamespace,true);e!=nil{return e}
 if c.Controller.Workers<1||c.Controller.Workers>32{return fmt.Errorf("workers deve estar entre 1 e 32")}
 if c.Controller.LeaderElection==nil{return fmt.Errorf("leaderElection obrigatório")}
 return c.Workflow.Validate()
}
