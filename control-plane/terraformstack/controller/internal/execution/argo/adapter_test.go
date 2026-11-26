package argo

import (
 "context"
 "strings"
 "testing"

 "k8s.io/apimachinery/pkg/runtime"
 "sigs.k8s.io/controller-runtime/pkg/client"
 "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testConfig() WorkflowConfig {
 return WorkflowConfig{
  Namespace: "argo-workflows",
  TemplateRef: TemplateRef{Name:"custom-workflow"},
  Parameters: Parameters{
   Input:InputParameters{Intent:"request-json",Volume:"workspace"},
   Output:OutputParameters{Decision:"decision",Authorization:"hash",Approval:"approval",Result:"result"},
  },
  Templates:Templates{Policy:"policy-check",Approval:"gate"},
 }
}
func TestAdapterLifecycle(t *testing.T) {
 ctx:=context.Background()
 cl:=fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithStatusSubresource(WorkflowObject()).Build()
 a:=&Adapter{Client:cl,Reader:cl,Config:testConfig()}
 if e:=a.Validate();e!=nil{t.Fatal(e)}
 if w,e:=a.Get(ctx,"absent","uid");e!=nil||w!=nil{t.Fatalf("missing: %v %v",w,e)}
 run:=M{"workflow":"test-run","intent":M{"uid":"uid","stack":"stack","serviceAccount":"executor","volume":"pvc"}}
 if e:=a.Create(ctx,run);e!=nil{t.Fatal(e)}
 w,e:=a.Raw(ctx,"test-run");if e!=nil{t.Fatal(e)}
 if w.GetNamespace()!=a.Config.Namespace||w.GetLabels()[uidLabel]!="uid"{t.Fatal(w)}
 spec:=obj(w.Object["spec"])
 if str(obj(spec["workflowTemplateRef"])["name"])!="custom-workflow"||spec["serviceAccountName"]!="executor"{t.Fatal(spec)}
 params:=arr(obj(spec["arguments"])["parameters"])
 if obj(params[0])["name"]!="request-json"||obj(params[1])["name"]!="workspace"||obj(params[1])["value"]!="pvc"{t.Fatal(params)}
 if e:=a.Cancel(ctx,"test-run","wrong");e==nil{t.Fatal("invalid owner accepted")}
 if x,e:=a.Get(ctx,"test-run","wrong");e!=nil||x.Valid{t.Fatalf("invalid identity: %v %v",x,e)}
 if e:=a.Cancel(ctx,"test-run","uid");e!=nil{t.Fatal(e)}
 if e:=a.Cancel(ctx,"test-run","uid");e!=nil{t.Fatal(e)}
 if e:=cl.Get(ctx,client.ObjectKey{Namespace:a.Config.Namespace,Name:"test-run"},w);e!=nil{t.Fatal(e)}
 if obj(w.Object["spec"])["shutdown"]!="Terminate"{t.Fatal(w)}
 w.Object["status"]=M{"phase":"Succeeded"}
 if e:=cl.Status().Update(ctx,w);e!=nil{t.Fatal(e)}
 x,e:=a.Get(ctx,"test-run","uid");if e!=nil||!x.Valid||!x.Terminal||!x.Succeeded{t.Fatalf("%v %v",x,e)}
}

func TestResultNormalization(t *testing.T) {
 w:=WorkflowObject()
 w.SetLabels(map[string]string{"workflows.argoproj.io/actor":"approver"})
 w.Object["status"]=M{
  "outputs":M{"parameters":[]any{
   M{"name":"decision","value":`{"decision":"RequireApproval","hash":"abc","summary":{"count":2},"items":[1,2.5]}`},
   M{"name":"hash","value":"abc"},
   M{"name":"approval","value":"Rejected"},
  }},
  "nodes":M{"gate":M{"type":"Suspend","templateName":"gate","phase":"Running","message":strings.Repeat("x",600)}},
 }
 got,e:=resultOf(w,testConfig());if e!=nil{t.Fatal(e)}
 if got["decision"]!="RequireApproval"||got["applyRequest"]!="abc"||got["approval"]!="Running"||got["approvalActor"]!="approver"||got["outcome"]!="Rejected"{t.Fatal(got)}
 if obj(got["summary"])["count"]!=int64(2)||arr(got["items"])[0]!=int64(1)||arr(got["items"])[1]!=2.5||len(str(got["approvalMessage"]))!=512{t.Fatal(got)}
}
func TestInvalidPolicyOutput(t *testing.T) {
 for _,raw:=range []string{"{","null","[]",`{"decision":"AutoApply"} garbage`} {
  w:=WorkflowObject();w.Object["status"]=M{"outputs":M{"parameters":[]any{M{"name":"decision","value":raw}}}}
  got,e:=resultOf(w,testConfig());if e!=nil||got["decision"]!="Error"{t.Fatalf("%q: %v %v",raw,got,e)}
 }
 w:=WorkflowObject();w.Object["status"]=M{"nodes":M{"policy":M{"templateName":"policy-check","phase":"Error"}}}
 got,e:=resultOf(w,testConfig());if e!=nil||got["decision"]!="Error"{t.Fatalf("%v %v",got,e)}
}
