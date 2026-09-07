package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "operator/api/v1alpha1"
)

type catcher struct {
	mu       sync.Mutex
	received map[types.NamespacedName]int
}

func newCatcher() *catcher {
	return &catcher{received: map[types.NamespacedName]int{}}
}

func (c *catcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/hooks/"), "/")
	if len(parts) != 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key := types.NamespacedName{Namespace: parts[0], Name: parts[1]}
	c.mu.Lock()
	c.received[key]++
	c.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (c *catcher) count(key types.NamespacedName) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.received[key]
}

type webhookTargetReconciler struct {
	client.Client
	catcher *catcher
}

func (r *webhookTargetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var target platformv1alpha1.WebhookTarget
	if err := r.Get(ctx, req.NamespacedName, &target); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	endpoint := fmt.Sprintf("http://operator.%s.svc.cluster.local/hooks/%s/%s", req.Namespace, req.Namespace, req.Name)
	received := r.catcher.count(req.NamespacedName)

	if target.Status.Endpoint != endpoint || target.Status.Received != received {
		target.Status.Endpoint = endpoint
		target.Status.Received = received
		if err := r.Status().Update(ctx, &target); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		slog.Error("unable to add scheme", "err", err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme})
	if err != nil {
		slog.Error("unable to start manager", "err", err)
		os.Exit(1)
	}

	c := newCatcher()

	err = ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.WebhookTarget{}).
		Complete(&webhookTargetReconciler{Client: mgr.GetClient(), catcher: c})
	if err != nil {
		slog.Error("unable to create controller", "err", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("catcher listening", "addr", ":8090")
		if err := http.ListenAndServe(":8090", c); err != nil {
			slog.Error("catcher server error", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		slog.Error("manager exited with error", "err", err)
		os.Exit(1)
	}
}
