package controller

import (
 "crypto/sha256"
 "encoding/json"
 "fmt"
 "regexp"
 "slices"
 "time"
 "strconv"
 "strings"

 api "github.com/DavidFerreira21/platform-engineering-poc/terraformstack/api/v1alpha1"
 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
 "k8s.io/apimachinery/pkg/runtime"
)

const revision = "platform.example.io/revision"
const finalizer = "platform.example.io/terraformstack-cleanup"
const uidLabel = "platform.example.io/stack-uid"
type M = map[string]any
type Catalog struct {
 Sources map[string]struct {
 Repository string
 Paths []string
 }
 Targets map[string]struct { Stacks map[string]M }
}
func obj(v any) M { m,_:=v.(map[string]any); if m==nil{return M{}}; return m }
func str(v any) string { s,_:=v.(string);return s }
func yes(v any) bool { b,_:=v.(bool);return b }
func num(v any) int64 { switch n:=v.(type){case int64:return n;case float64:return int64(n);case int:return int64(n)};return 0 }
func arr(v any) []any { a,_:=v.([]any);return a }
func copyMap(m M) M { return runtime.DeepCopyJSON(m) }
func canonical(v any) string { b,_:=json.Marshal(v);return string(b) }
func equal(a,b any) bool {return canonical(a)==canonical(b)}
func stackObject() *unstructured.Unstructured {o:=&unstructured.Unstructured{};o.SetGroupVersionKind(api.StackGVK);return o}
func stackList() *unstructured.UnstructuredList {o:=&unstructured.UnstructuredList{};o.SetGroupVersionKind(api.GroupVersion.WithKind("TerraformStackList"));return o}
func specOf(o *unstructured.Unstructured) (api.StackSpec,error) {var s api.StackSpec;err:=runtime.DefaultUnstructuredConverter.FromUnstructured(obj(o.Object["spec"]),&s);return s,err}
var shaPattern=regexp.MustCompile("^[0-9a-f]{40}$")
var profilePattern=regexp.MustCompile("^[a-z][a-z0-9-]*$")
var durationPattern=regexp.MustCompile("^([0-9]+[smh])+$")
func duration(v string)(time.Duration,error){
 if !durationPattern.MatchString(v){return 0,fmt.Errorf("duração inválida: %s",v)}
 d,e:=time.ParseDuration(v);if e!=nil||d<=0||d>24*time.Hour{return 0,fmt.Errorf("duração fora do limite: %s",v)};return d,nil
}
func fallback(v,d string)string{if v==""{return d};return v}

// Resolve exclusivamente dados autorizados no catálogo. A intenção serializada
// é o contrato compartilhado com o runner, inclusive nomes e tipos JSON.
func Intent(o *unstructured.Unstructured,c Catalog)(M,error){
 s,e:=specOf(o);if e!=nil{return nil,e}
 deleting:=!o.GetDeletionTimestamp().IsZero()
 d:=M{"stack":o.GetName(),"uid":string(o.GetUID()),"generation":o.GetGeneration(),"deleting":deleting}
 if deleting&&!s.Lifecycle.DestroyResourcesOnDeletion{d["operation"]="retain";return d,nil}
 source,ok:=c.Sources[s.SourceRef.Name];slot:=c.Targets[s.Target.Entity+"/"+s.Target.Environment].Stacks[o.GetName()]
 if !ok||slot==nil||!slices.Contains(source.Paths,s.SourceRef.Path)||str(slot["source"])!=s.SourceRef.Name||str(slot["path"])!=s.SourceRef.Path{return nil,fmt.Errorf("fonte ou destino fora do catálogo")}
 profile:=str(slot["policyProfile"]);if !profilePattern.MatchString(profile){return nil,fmt.Errorf("Policy Profile ausente ou inválido")}
 pinned:=o.GetAnnotations()[revision];rev:=fallback(s.SourceRef.Version,"latest");cfg:=fallback(s.Configuration.Version,"latest")
 if rev=="latest"{rev=pinned};if cfg=="latest"{cfg=pinned}
 if !shaPattern.MatchString(rev)||!shaPattern.MatchString(cfg){return nil,fmt.Errorf("fonte e configuração exigem SHAs Git completos")}
 for _,k:=range []string{"backend","variables","volume","serviceAccount","configurationPath"}{if slot[k]==nil{return nil,fmt.Errorf("campo obrigatório ausente no catálogo: %s",k)};d[k]=slot[k]}
 d["revision"]=rev;d["configuration"]=cfg;d["revisionAnnotation"]=pinned;d["source"]=s.SourceRef.Name;d["repository"]=source.Repository;d["path"]=s.SourceRef.Path;d["target"]=M{"entity":s.Target.Entity,"environment":s.Target.Environment};d["policyProfile"]=profile
 d["operation"]="apply";if deleting||s.Lifecycle.Destroy{d["operation"]="destroy"};return d,nil
}
func retryDelay(p api.RetryPolicy,attempt int64)(time.Duration,error){
 if p.Strategy=="None"||attempt>=int64(p.MaxRetries){return 0,nil}
 d,e:=duration(fallback(p.Interval,"10s"));if e!=nil{return 0,e}
 if fallback(p.Strategy,"ExponentialBackoff")=="ExponentialBackoff"{max,e:=duration(fallback(p.MaxInterval,"5m"));if e!=nil{return 0,e};for i:=int64(0);i<attempt&&d<max;i++{d*=2};if d>max{d=max}}
 return d,nil
}
func execution(d,old M,trigger string,attempt int64)M{
 sequence:=num(old["sequence"])+1
 sum:=sha256.Sum256([]byte(canonical(d)))
 return M{"intent":d,"workflow":fmt.Sprintf("tf-%s-%x-%d",str(d["uid"]),sum[:6],sequence),"sequence":sequence,"attempt":attempt,"trigger":trigger}
}
func conditionTrue(o *unstructured.Unstructured,kind string)bool{
 status:=obj(o.Object["status"])
 for _,v:=range arr(status["conditions"]){c:=obj(v);if str(c["type"])==kind&&str(c["status"])=="True"&&num(c["observedGeneration"])==o.GetGeneration(){return true}}
 return false
}
func DependenciesReady(all []unstructured.Unstructured,o *unstructured.Unstructured,d M,c Catalog)bool{
 s,e:=specOf(o);if e!=nil{return false}
 if d["operation"]=="apply"{
 for _,dep:=range s.DependsOn{
 found:=false
 for i:=range all{other:=&all[i];if other.GetName()!=dep.Name||other.GetName()==o.GetName(){continue};os,e:=specOf(other);if e!=nil||os.Target!=s.Target||!other.GetDeletionTimestamp().IsZero(){return false}
 wanted,e:=Intent(other,c);status:=obj(other.Object["status"])
 found=e==nil&&num(status["observedGeneration"])==other.GetGeneration()&&equal(obj(status["execution"])["intent"],wanted)&&conditionTrue(other,"Ready");break}
 if !found{return false}
 }
 }
 if d["operation"]=="destroy"{for i:=range all{other:=&all[i];os,e:=specOf(other);if e!=nil||os.Target!=s.Target{continue};for _,dep:=range os.DependsOn{if dep.Name==o.GetName()&&!conditionTrue(other,"Destroyed"){return false}}}}
 return true
}
func resultOf(w *unstructured.Unstructured)(M,error){
 result:=M{};if w==nil{return result,nil}
 status:=obj(w.Object["status"]);outputs:=M{}
 for _,p:=range arr(obj(status["outputs"])["parameters"]){m:=obj(p);outputs[str(m["name"])]=m["value"]}
 if raw:=str(outputs["decision-record"]);raw!=""{if e:=json.Unmarshal([]byte(raw),&result);e!=nil||result==nil{return M{"decision":"Error","reason":"Resultado OPA inválido"},nil}}
 for _,v:=range obj(status["nodes"]){n:=obj(v);if n["templateName"]=="opa"&&(n["phase"]=="Failed"||n["phase"]=="Error"){result["decision"]="Error";result["reason"]="OPA falhou ao avaliar o plano"}
 if n["type"]=="Suspend"&&n["templateName"]=="approval"{result["approval"]=n["phase"];if a:=w.GetLabels()["workflows.argoproj.io/actor"];a!=""{result["approvalActor"]=a};if m:=str(n["message"]);m!=""{r:=[]rune(m);if len(r)>512{r=r[:512]};result["approvalMessage"]=string(r)}}}
 if r:=str(outputs["apply-request"]);r!=""&&r!="none"{result["applyRequest"]=r}
 if outputs["approval-result"]=="Rejected"{result["outcome"]="Rejected"}
 if r:=str(outputs["apply-result"]);r!=""{result["outcome"]=r}
 // JSON numbers from external output must use Kubernetes-compatible types.
 return normalize(result),nil
}
func normalize(m M)M{for k,v:=range m{m[k]=normalizeValue(v)};return m}
func normalizeValue(v any)any{switch x:=v.(type){case M:return normalize(x);case []any:for i:=range x{x[i]=normalizeValue(x[i])};case json.Number:if n,e:=strconv.ParseInt(string(x),10,64);e==nil{return n};n,_:=strconv.ParseFloat(string(x),64);return n};return v}
func readCatalog(b []byte)(Catalog,error){var c Catalog;dec:=json.NewDecoder(strings.NewReader(string(b)));dec.UseNumber();e:=dec.Decode(&c);for _,t:=range c.Targets{for _,s:=range t.Stacks{normalize(s)}};return c,e}
