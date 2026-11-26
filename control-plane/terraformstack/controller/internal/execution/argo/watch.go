package argo

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (a *Adapter) WatchHandler(stackNamespace string) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		if o.GetNamespace() != a.Config.Namespace || o.GetLabels()["platform.example.io/workflow"] != "terraformstack" {
			return nil
		}
		name := o.GetLabels()["platform.example.io/stack-name"]
		if name == "" {
			return nil
		}
		return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: stackNamespace, Name: name}}}
	})
}
