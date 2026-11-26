package main

import (
	"flag"
	"os"
	"time"

	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/controller"
	"github.com/DavidFerreira21/platform-engineering-poc/terraformstack/internal/execution/argo"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "/etc/terraformstack-config/controller-config.yaml", "contrato de integração")
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	logger := ctrl.Log.WithName("setup")
	config, e := controller.LoadConfig(configPath)
	if e != nil {
		logger.Error(e, "configuração inválida")
		os.Exit(1)
	}
	stackNS, wfNS := config.Controller.StackNamespace, config.Workflow.Namespace
	workers, leader := config.Controller.Workers, *config.Controller.LeaderElection
	lease := 15 * time.Second
	renew := 10 * time.Second
	retry := 2 * time.Second
	manager, e := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Cache:          cache.Options{ByObject: map[client.Object]cache.ByObject{controller.StackObject(): {Namespaces: map[string]cache.Config{stackNS: {}}}, argo.WorkflowObject(): {Namespaces: map[string]cache.Config{wfNS: {}}}}},
		LeaderElection: leader, LeaderElectionID: "terraformstack-controller.platform.example.io", LeaderElectionNamespace: stackNS, LeaseDuration: &lease, RenewDeadline: &renew, RetryPeriod: &retry,
		Metrics: metrics.Options{BindAddress: ":8080"}, HealthProbeBindAddress: ":8081",
	})
	if e != nil {
		logger.Error(e, "manager")
		os.Exit(1)
	}
	engine := &argo.Adapter{Client: manager.GetClient(), Reader: manager.GetAPIReader(), Config: config.Workflow}
	r := &controller.Reconciler{Client: manager.GetClient(), Reader: manager.GetAPIReader(), StackNamespace: stackNS, Engine: engine}
	if e = r.SetupWithManager(manager, workers, argo.WorkflowObject(), engine.WatchHandler(stackNS)); e != nil {
		logger.Error(e, "controller")
		os.Exit(1)
	}
	if e = manager.AddHealthzCheck("healthz", healthz.Ping); e != nil {
		logger.Error(e, "healthz")
		os.Exit(1)
	}
	if e = manager.AddReadyzCheck("readyz", healthz.Ping); e != nil {
		logger.Error(e, "readyz")
		os.Exit(1)
	}
	if e = manager.Start(ctrl.SetupSignalHandler()); e != nil {
		logger.Error(e, "manager encerrado")
		os.Exit(1)
	}
}
