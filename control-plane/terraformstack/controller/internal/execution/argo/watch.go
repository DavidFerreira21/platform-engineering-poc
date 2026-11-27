package argo

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func stackName(o client.Object) string {
	if name := o.GetLabels()[stackNameKey]; name != "" {
		return name
	}
	return o.GetAnnotations()[stackNameKey]
}

func (a *Adapter) WatchHandler(stackNamespace string) handler.EventHandler {
	return WatchHandlerForNamespaces(stackNamespace, map[string]bool{a.Config.Namespace: true})
}

// WatchHandlerForNamespaces maps only Workflows in the configured execution namespaces.
func WatchHandlerForNamespaces(stackNamespace string, namespaces map[string]bool) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
		if !namespaces[o.GetNamespace()] || o.GetLabels()["platform.example.io/workflow"] != "terraformstack" {
			return nil
		}
		name := stackName(o)
		if name == "" {
			return nil
		}
		return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: stackNamespace, Name: name}}}
	})
}
