# Learning Crossplane by building its equivalent by hand

Goal: understand what Crossplane actually automates, by building the same
capability twice — once as a hand-rolled Kubernetes operator, once using
Crossplane's own primitives — and comparing them directly.

Assumes: working Go, Kubernetes basics (objects, kubectl, what a CRD is at a
user level). Lessons skip that and go straight at controller-runtime and
Crossplane internals.

Both halves integrate with the existing `app/` (the job-queue/webhook service
from Part 1), so the loop is real and observable, not a toy example.

## Module A — Hand-rolled operator (`operator/`)

1. **A CRD is a real extension of the API server, not a YAML fiction.**
   *(done)* The moment you apply the CRD, `WebhookTarget` becomes a
   first-class resource: stored in etcd, versioned, watchable, RBAC-controlled
   — exactly like `Pod`. The Go type is optional client-side plumbing on top
   of that; you could operate this resource with pure `kubectl`/YAML forever
   and never write a line of Go. The one design choice actually worth
   understanding here: `.status` as a separate subresource, so a controller
   (reporting observed reality) and a human/CI (declaring desired state) can
   write to the same object without racing or needing overlapping RBAC.

2. **The `Manager`/`Reconciler` machinery.** Wire up a real `Reconciler`
   (`Reconcile(ctx, req ctrl.Request) (ctrl.Result, error)`) and trace the
   actual plumbing: `Watch` registers an informer → the informer's event
   handlers don't call your code directly, they enqueue a bare
   `Request{Namespace, Name}` onto a rate-limited workqueue → a worker pulls
   the request and calls `Reconcile`, which re-fetches the *current* object
   from the cache — never the object from the event that triggered it.
   Concept: `Request` deliberately carries no object data, so N rapid
   changes to one resource coalesce into a single reconcile of current
   state, and reconciliation is level-triggered, not edge-triggered — which
   is what makes a controller self-healing after a restart (it re-converges
   from whatever state exists, it doesn't need to "catch up" on missed
   events). Code: make the `WebhookTarget` reconciler patch `.status`, run
   it out-of-cluster against the real kubeconfig, watch status update live —
   then kill and restart the process to see it re-converge without having
   "missed" anything.

3. **Cross-resource reads + external I/O from inside a reconciler.**
   `ScheduledJob.spec.targetRef` — the reconciler does a live `Get` for the
   referenced `WebhookTarget`, reads its status, then makes a real HTTP
   `POST` to the app's `/jobs`. Concept: reconcilers doing external I/O is
   normal, but the operation must be safe to retry, since the same state can
   be reconciled more than once; `ctrl.Result{RequeueAfter: ...}` is the
   mechanism for "call me again later" — no cron library, no external
   timers, just the reconcile loop scheduling its own next invocation.

4. **RBAC is the actual security boundary, not the Go code.** Write the
   least-privilege `Role`/`RoleBinding` — get/list/watch on both CRDs,
   update specifically on their `/status` subresource, get on the app's
   `Service` — and see that a manifest/RBAC mismatch fails at *runtime*
   (403 from the API server) with no compile-time signal at all. Ties back
   to Lesson 1: the YAML is the real, enforced contract; the Go code's
   correctness is irrelevant if the RBAC doesn't grant what it needs.

## Module B — The same shape of problem, Crossplane's way

5. **Install Crossplane + a provider.** Crossplane core via Helm,
   `provider-family-gcp` + `provider-gcp-storage`, `ProviderConfig` via GKE
   Workload Identity (keyless, same rationale as the GitHub Actions WIF
   setup). Concept: a `Provider` package is a container running exactly the
   same shape of `Manager`/`Reconciler` code as Module A, just written by
   Upbound and shipped as a package instead of hand-written — go read its
   pod logs and watch it reconcile.

6. **XRD generates a CRD.** Apply an XRD and watch a new
   `CustomResourceDefinition` appear (`kubectl get crd`) — same shape as
   Lesson 1's hand-written one, now synthesized. Concept: an XRD produces
   *two* kinds — a cluster-scoped composite resource (XR) and a namespaced
   claim — analogous to how a `PersistentVolumeClaim` claims a
   cluster-scoped `PersistentVolume`; the split exists for multi-tenancy
   (claims are what namespaced users touch, the XR is the real object
   underneath).

7. **Composition is a generic reconciler you don't write.** Patch-and-transform
   Composition renders child resources (here, a `Bucket`) from a template —
   functionally the same shape as Lesson 3's hand-written reconciler
   (read desired state, compute + apply derived resources), except the
   reconciler is Crossplane's own generic engine and your "code" is a YAML
   template. Apply a claim, watch the `Bucket` managed resource appear, watch
   a real GCS bucket get created.

8. **(Stretch) Composition Function in Go.** Replace Lesson 7's YAML
   patch-and-transform with the same transform written in Go
   (`function-sdk-go`), which Crossplane calls out to over gRPC per
   reconcile. Concept: closes the loop — the same Go skills from Module A,
   now writing the piece of logic Crossplane's reconciler delegates to when
   declarative patches aren't expressive enough.

## How we're working through this

One lesson at a time, each ending with a concrete check against the real
cluster before moving on — and the reasoning explained as part of building
each piece, not summarized afterward.
