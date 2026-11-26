package controller

import (
 "context"
 "fmt"
 "slices"
 "time"

 apierrors "k8s.io/apimachinery/pkg/api/errors"
 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
 "k8s.io/apimachinery/pkg/runtime/schema"
 ctrl "sigs.k8s.io/controller-runtime"
 "sigs.k8s.io/controller-runtime/pkg/client"
 "sigs.k8s.io/controller-runtime/pkg/log"
)

var workflowGVK=schema.GroupVersionKind{Group:"argoproj.io",Version:"v1alpha1",Kind:"Workflow"}
type Reconciler struct {
 client.Client
 Reader client.Reader // Leitura consistente antes de decisões de autorização.
 Catalog Catalog
 StackNamespace, WorkflowNamespace string
 Now func()time.Time
}
func (r *Reconciler) clock()time.Time{if r.Now!=nil{return r.Now().UTC()};return time.Now().UTC()}
func (r *Reconciler) stamp()string{return r.clock().Format(time.RFC3339)}
func workflowObject()*unstructured.Unstructured{o:=&unstructured.Unstructured{};o.SetGroupVersionKind(workflowGVK);return o}
func (r *Reconciler) getWorkflow(ctx context.Context,name string)(*unstructured.Unstructured,error){
 if name==""{return nil,nil};w:=workflowObject();e:=r.Reader.Get(ctx,client.ObjectKey{Namespace:r.WorkflowNamespace,Name:name},w);if apierrors.IsNotFound(e){return nil,nil};return w,e
}
func (r *Reconciler) body(run M)*unstructured.Unstructured{
 d:=obj(run["intent"]);w:=workflowObject();w.SetName(str(run["workflow"]));w.SetNamespace(r.WorkflowNamespace);w.SetLabels(map[string]string{"platform.example.io/workflow":"terraformstack",uidLabel:str(d["uid"]),"platform.example.io/stack-name":str(d["stack"])})
 w.Object["spec"]=M{"serviceAccountName":d["serviceAccount"],"workflowTemplateRef":M{"name":"terraformstack"},"arguments":M{"parameters":[]any{M{"name":"intent","value":canonical(d)},M{"name":"volume","value":d["volume"]}}}};return w
}
func terminal(p string)bool{return p=="Succeeded"||p=="Failed"||p=="Error"}
func (r *Reconciler) terminate(ctx context.Context,w *unstructured.Unstructured)error{
 if obj(w.Object["spec"])["shutdown"]=="Terminate"{return nil}
 old:=w.DeepCopy();obj(w.Object["spec"])["shutdown"]="Terminate";return r.Patch(ctx,w,client.MergeFromWithOptions(old,client.MergeFromWithOptimisticLock{}))
}

// O status inteiro é atualizado com resourceVersion. Isso impede sobrescrever
// o claim feito pelo runner entre a leitura e a escrita.
func (r *Reconciler) report(ctx context.Context,o *unstructured.Unstructured,kind,reason,message string,values M)error{
 old:=obj(o.Object["status"]);status:=copyMap(old);for k,v:=range values{status[k]=v}
 item:=M{"type":kind,"status":"True","reason":reason,"message":message,"observedGeneration":o.GetGeneration(),"lastTransitionTime":r.stamp()}
 // Mantém a condição corrente em primeiro lugar (contrato dos labs), e
 // conserva as demais condições com False e seus tempos de transição.
 conditions:=[]any{item}
 for _,v:=range arr(old["conditions"]){c:=copyMap(obj(v));if c["type"]==kind{if c["status"]=="True"{item["lastTransitionTime"]=c["lastTransitionTime"]};continue};if c["status"]!="False"{c["lastTransitionTime"]=r.stamp()};c["status"]="False";c["observedGeneration"]=o.GetGeneration();conditions=append(conditions,c)}
 status["observedGeneration"]=o.GetGeneration();status["conditions"]=conditions
 if equal(status,old){return nil}
 o.Object["status"]=status
 e:=r.Status().Update(ctx,o)
 if e==nil{log.FromContext(ctx).Info("transição TerraformStack","stack",o.GetName(),"condition",kind,"reason",reason)}
 return e
}
func again(e error)(ctrl.Result,error){return ctrl.Result{Requeue:true},e}
func done(e error)(ctrl.Result,error){return ctrl.Result{},e}
func (r *Reconciler) Reconcile(ctx context.Context,req ctrl.Request)(ctrl.Result,error){
 if req.Namespace!=r.StackNamespace{return done(nil)}
 o:=stackObject();if e:=r.Reader.Get(ctx,req.NamespacedName,o);e!=nil{return done(client.IgnoreNotFound(e))}
 status:=obj(o.Object["status"]);run:=copyMap(obj(status["execution"]));deleting:=!o.GetDeletionTimestamp().IsZero()
 if deleting&&!slices.Contains(o.GetFinalizers(),finalizer){return done(nil)}
 d,e:=Intent(o,r.Catalog);s,se:=specOf(o);if e==nil{e=se}
 var interval time.Duration
 if e==nil&&d["operation"]!="retain"&&s.Reconciliation.DriftCheckInterval!=""{interval,e=duration(s.Reconciliation.DriftCheckInterval)}
 if e==nil&&deleting&&status["deletionIntent"]!=nil&&!equal(status["deletionIntent"],d){e=fmt.Errorf("intenção de exclusão diverge do snapshot")}
 active:=len(run)>0&&str(run["finishedAt"])==""
 var w *unstructured.Unstructured
 if active{var err error;w,err=r.getWorkflow(ctx,str(run["workflow"]));if err!=nil{return done(err)}}
 phase:="";if w!=nil{phase=str(obj(w.Object["status"])["phase"])}
 if e!=nil{
  if active&&w!=nil&&!terminal(phase)&&!yes(run["applying"]){
   // Persistir cancelamento antes de encerrar o Workflow: disputa CAS com claim.
   if !yes(run["cancelling"]){run["cancelling"]=true;return again(r.report(ctx,o,"ReconcileBlocked","InvalidIntent",e.Error(),M{"execution":run}))}
   if err:=r.terminate(ctx,w);err!=nil{return done(err)}
  }
  return done(r.report(ctx,o,"ReconcileBlocked","InvalidIntent",e.Error(),M{}))
 }
 if deleting&&status["deletionIntent"]==nil{status=copyMap(status);status["deletionIntent"]=d;o.Object["status"]=status;return again(r.Status().Update(ctx,o))}
 if !deleting&&!slices.Contains(o.GetFinalizers(),finalizer){o.SetFinalizers(append(o.GetFinalizers(),finalizer));return again(r.Update(ctx,o))}
 // A lista vem do cache dos informers (sem polling periódico global).
 all:=stackList();if e:=r.List(ctx,all,client.InNamespace(r.StackNamespace));e!=nil{return done(e)}
 eligible:=DependenciesReady(all.Items,o,d,r.Catalog)
 changed:=!equal(d,run["intent"]);suspended:=s.Suspend
 if active{
  if w==nil{
   if yes(run["created"])||yes(run["applying"]){return done(r.report(ctx,o,"Failed","WorkflowMissing","Execução removida; investigar Pods e state.",M{}))}
   if !changed&&!suspended&&eligible&&!yes(run["cancelling"]){
    // O nome determinístico impede duplicação mesmo após uma resposta perdida.
    err:=r.Create(ctx,r.body(run));if apierrors.IsAlreadyExists(err){return again(nil)};return again(err)
   }
  }else if !terminal(phase){
   if (changed||suspended||!eligible||yes(run["cancelling"]))&&!yes(run["applying"]){
    if !yes(run["cancelling"]){run["created"]=true;run["cancelling"]=true;return again(r.report(ctx,o,"ReconcilePending","Cancelling",w.GetName(),M{"execution":run}))}
    return done(r.terminate(ctx,w))
   }
   result,err:=resultOf(w);if err!=nil{return done(err)}
   run["created"]=true
   if yes(result["drift"])&&!yes(run["drift"]){run["drift"]=true;return again(r.report(ctx,o,"DriftDetected","ExternalChange",w.GetName(),M{"execution":run,"policy":result,"lastDriftTime":r.stamp()}))}
   request:=str(result["applyRequest"])
   if !changed&&!suspended&&eligible&&!yes(run["cancelling"])&&request!=""&&request==str(result["hash"])&&(result["decision"]=="AutoApply"||result["decision"]=="RequireApproval"){run["authorization"]=request}
   kind,reason:="Reconciling","Planning";if result["approval"]=="Running"{kind,reason="ApprovalRequired","AwaitingApproval"}else if yes(run["applying"]){reason="Applying"}
   return done(r.report(ctx,o,kind,reason,w.GetName(),M{"execution":run,"policy":result}))
  }
  result,err:=resultOf(w);if err!=nil{return done(err)}
  kind,reason:="Failed","WorkflowFailed"
  switch{
  case yes(run["cancelling"])||w==nil||result["outcome"]=="Cancelled":kind,reason="ReconcileBlocked","Cancelled"
  case result["outcome"]=="Rejected":kind,reason="ReconcileBlocked","ApprovalRejected"
  case result["decision"]=="Deny":kind,reason="ReconcileBlocked","PolicyDenied"
  case result["decision"]=="Error":kind,reason="ReconcileBlocked","PolicyError"
  case phase=="Succeeded"&&result["decision"]=="NoChanges"&&obj(run["intent"])["operation"]=="apply":kind,reason="Ready","NoChanges"
  case phase=="Succeeded"&&result["outcome"]=="Verified":kind,reason="Ready","Verified";if obj(run["intent"])["operation"]=="destroy"{kind="Destroyed"}
  }
  values:=M{"policy":result,"lastCheckTime":r.stamp()}
  if yes(result["drift"])&&!yes(run["drift"]){values["lastDriftTime"]=r.stamp()}
  run["finishedAt"]=r.stamp();run["result"]=kind;run["reason"]=reason;run["drift"]=yes(result["drift"]);values["execution"]=run
  if changed{kind,reason="ReconcilePending","PreviousExecutionFinished"}
  return again(r.report(ctx,o,kind,reason,str(run["workflow"]),values))
 }
 if d["operation"]=="retain"{o.SetFinalizers(slices.DeleteFunc(o.GetFinalizers(),func(v string)bool{return v==finalizer}));return done(r.Update(ctx,o))}
 if suspended{return done(r.report(ctx,o,"ReconcilePending","Suspended","Novas execuções suspensas.",M{}))}
 if !eligible{return done(r.report(ctx,o,"ReconcilePending","DependencyBlocked","Dependência pendente ou ainda ativa.",M{}))}
 trigger:="GitOps";attempt:=int64(0)
 if !changed{
  if deleting&&run["result"]=="Destroyed"{o.SetFinalizers(slices.DeleteFunc(o.GetFinalizers(),func(v string)bool{return v==finalizer}));return done(r.Update(ctx,o))}
  finished,err:=time.Parse(time.RFC3339,str(run["finishedAt"]));if err!=nil{return done(err)}
  var delay time.Duration
  if run["result"]=="Ready"{delay=interval;trigger="DriftCheck"}else if run["result"]=="Failed"{delay,err=retryDelay(s.Reconciliation.RetryPolicy,num(run["attempt"]));if err!=nil{return done(err)};trigger="Retry";attempt=num(run["attempt"])+1}
  if delay==0{return done(nil)}
  if remaining:=finished.Add(delay).Sub(r.clock());remaining>0{return ctrl.Result{RequeueAfter:remaining},nil}
 }
 next:=execution(d,run,trigger,attempt)
 return again(r.report(ctx,o,"Reconciling","Scheduling",str(next["workflow"]),M{"execution":next,"resolvedSource":M{"name":d["source"],"version":d["revision"],"configurationVersion":d["configuration"],"digest":d["revision"]},"policy":M{}}))
}
