package controller

import (
 "context"
 "os"
 "slices"

 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
 ctrl "sigs.k8s.io/controller-runtime"
 "sigs.k8s.io/controller-runtime/pkg/builder"
 "sigs.k8s.io/controller-runtime/pkg/client"
 crcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
 "sigs.k8s.io/controller-runtime/pkg/handler"
 "sigs.k8s.io/controller-runtime/pkg/predicate"
 "sigs.k8s.io/controller-runtime/pkg/reconcile"
 "sigs.k8s.io/controller-runtime/pkg/event"
 "sigs.k8s.io/controller-runtime/pkg/log"
)
func LoadCatalog(path string)(Catalog,error){b,e:=os.ReadFile(path);if e!=nil{return Catalog{},e};return readCatalog(b)}
const dependencyIndex="terraformstack.dependsOn"
func (r *Reconciler) SetupWithManager(m ctrl.Manager,workers int, executionObject client.Object, executionHandler handler.EventHandler)error{
 ctx:=context.Background()
 if e:=m.GetFieldIndexer().IndexField(ctx,stackObject(),dependencyIndex,func(o client.Object)[]string{s,e:=specOf(o.(*unstructured.Unstructured));if e!=nil{return nil};names:=[]string{};for _,d:=range s.DependsOn{names=append(names,d.Name)};return names});e!=nil{return e}
 // Eventos de dependências acionam também os pais (teardown reverso).
 neighbors:=handler.EnqueueRequestsFromMapFunc(func(ctx context.Context,o client.Object)[]reconcile.Request{
 if o.GetNamespace()!=r.StackNamespace{return nil}
 // O índice de dependências pertence ao cache, inclusive para objetos unstructured.
 list:=stackList();if e:=m.GetCache().List(ctx,list,client.InNamespace(r.StackNamespace),client.MatchingFields{dependencyIndex:o.GetName()});e!=nil{log.FromContext(ctx).Error(e,"index de dependências");return nil}
 names:=[]string{};for _,item:=range list.Items{names=append(names,item.GetName())}
 s,e:=specOf(o.(*unstructured.Unstructured));if e==nil{for _,d:=range s.DependsOn{if !slices.Contains(names,d.Name){names=append(names,d.Name)}}}
 reqs:=[]reconcile.Request{};for _,name:=range names{reqs=append(reqs,reconcile.Request{NamespacedName:client.ObjectKey{Namespace:r.StackNamespace,Name:name}})};return reqs
 })
 // Status do runner, geração, anotação Git e exclusão precisam gerar eventos.
 // Atualizações sem alteração efetiva de conteúdo não disparam a fila.
 changed:=predicate.Funcs{UpdateFunc:func(e event.UpdateEvent)bool{return e.ObjectOld.GetResourceVersion()!=e.ObjectNew.GetResourceVersion()}}
 return ctrl.NewControllerManagedBy(m).Named("terraformstack").
 For(stackObject(),builder.WithPredicates(changed)).
 Watches(stackObject(),neighbors,builder.WithPredicates(changed)).
 Watches(executionObject,executionHandler,builder.WithPredicates(changed)).
 WithOptions(crcontroller.Options{MaxConcurrentReconciles:workers}).Complete(r)
}
