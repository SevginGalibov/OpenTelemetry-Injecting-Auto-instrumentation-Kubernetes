package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NamespaceReconciler struct {
	client.Client
	IgnoreList    map[string]bool
	Collector     string
	AuthHeader    string
	SamplerType   string
	SamplerArg    string
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	if r.IgnoreList[req.Name] {
		log.Info("Namespace in ignore list, skipping", "namespace", req.Name)
		return ctrl.Result{}, nil
	}

	ns := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	inst := &unstructured.Unstructured{}
	inst.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "opentelemetry.io",
		Version: "v1alpha1",
		Kind:    "Instrumentation",
	})
	err := r.Get(ctx, client.ObjectKey{Name: "auto-instrumentation", Namespace: ns.Name}, inst)
	if err == nil {
		log.Info("Instrumentation already exists, skipping")
		return ctrl.Result{}, nil
	}

	instObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "opentelemetry.io/v1alpha1",
			"kind":       "Instrumentation",
			"metadata": map[string]interface{}{
				"name":      "auto-instrumentation",
				"namespace": ns.Name,
			},
			"spec": map[string]interface{}{
				"exporter": map[string]interface{}{
					"endpoint": r.Collector,
				},
				"propagators": []interface{}{"tracecontext", "baggage"},
				"sampler": map[string]interface{}{
					"type":     r.SamplerType,
					"argument": r.SamplerArg,
				},
				"env": []interface{}{
					map[string]interface{}{
						"name":  "OTEL_EXPORTER_OTLP_HEADERS",
						"value": r.AuthHeader,
					},
					map[string]interface{}{
						"name":  "OTEL_TRACES_SAMPLER",
						"value": r.SamplerType,
					},
				},
				"python": map[string]interface{}{
					"env": []interface{}{
						map[string]interface{}{
							"name":  "OTEL_EXPORTER_OTLP_ENDPOINT",
							"value": r.Collector,
						},
					},
				},
				"dotnet": map[string]interface{}{
					"env": []interface{}{
						map[string]interface{}{
							"name":  "OTEL_EXPORTER_OTLP_ENDPOINT",
							"value": r.Collector,
						},
						map[string]interface{}{
							"name":  "OTEL_DOTNET_AUTO_TRACES_ENABLED_INSTRUMENTATIONS",
							"value": "AspNetCore,HttpClient,GrpcClient,SqlClient",
						},
					},
				},
				"java": map[string]interface{}{
					"env": []interface{}{
						map[string]interface{}{
							"name":  "OTEL_EXPORTER_OTLP_ENDPOINT",
							"value": r.Collector,
						},
					},
				},
				"nodejs": map[string]interface{}{
					"env": []interface{}{
						map[string]interface{}{
							"name":  "OTEL_EXPORTER_OTLP_ENDPOINT",
							"value": r.Collector,
						},
					},
				},
			},
		},
	}

	if err := r.Create(ctx, instObj); err != nil {
		log.Error(err, "unable to create Instrumentation")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	log.Info("Instrumentation created", "namespace", ns.Name)
	return ctrl.Result{}, nil
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	ignoreList := strings.Split(os.Getenv("IGNORE_NAMESPACES"), ",")
	ignoreMap := make(map[string]bool)
	for _, ns := range ignoreList {
		ignoreMap[strings.TrimSpace(ns)] = true
	}

	collector := os.Getenv("COLLECTOR_ENDPOINT")
	if collector == "" {
		collector = "http://simplest-collector.default.svc.cluster.local:4318"
	}

	authHeader := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if authHeader == "" {
		authHeader = "Authorization=Basic x"
	}

	samplerType := os.Getenv("OTEL_TRACES_SAMPLER")
	if samplerType == "" {
		samplerType = "parentbased_traceidratio"
	}

	samplerArg := os.Getenv("OTEL_TRACES_SAMPLER_ARG")
	if samplerArg == "" {
		samplerArg = "1.0"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		panic(fmt.Sprintf("unable to start manager: %v", err))
	}

	reconciler := &NamespaceReconciler{
		Client:     mgr.GetClient(),
		IgnoreList: ignoreMap,
		Collector:  collector,
		AuthHeader: authHeader,
		SamplerType: samplerType,
		SamplerArg:  samplerArg,
	}

	ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc:  func(e event.CreateEvent) bool { return true },
			UpdateFunc:  func(e event.UpdateEvent) bool { return false },
			DeleteFunc:  func(e event.DeleteEvent) bool { return false },
			GenericFunc: func(e event.GenericEvent) bool { return false },
		}).
		Complete(reconciler)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
