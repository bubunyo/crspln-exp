# Learning Crossplane by building its equivalent by hand

Goal: understand what Crossplane actually automates, by building the same
capability twice — once as a hand-rolled Kubernetes operator, once using
Crossplane's own primitives — and comparing them directly.

Both halves integrate with the existing `app/` (the job-queue/webhook service
from Part 1), so the loop is real and observable, not a toy example.

## Module A — Hand-rolled operator (`operator/`)

No Crossplane. Just `controller-runtime`, two custom resources, and a
reconcile loop written by hand.

1. **CRD fundamentals** — define the `WebhookTarget` Go type (group
   `platform.crspln-exp.io/v1alpha1`), hand-write its CRD YAML, apply an
   instance, and observe that nothing happens yet.
   Concept: a CRD is just a schema. No behavior exists until something
   watches it.

2. **The reconcile loop** — write the `WebhookTarget` controller: an
   embedded HTTP catcher plus a reconciler that updates `.status.received`.
   Run it locally (out-of-cluster, against the real cluster via kubeconfig)
   and watch the status field update live.
   Concept: watches, reconcile loops, status subresources.

3. **A second CRD referencing the first** — `ScheduledJob`, with
   `spec.targetRef` naming a `WebhookTarget`. On a timer, it resolves the
   target's status and calls the real app's `/jobs` endpoint.
   Concept: cross-resource references, calling external services from a
   controller, `RequeueAfter` as a minimal scheduler.

4. **Ship it** — Dockerize the operator, write least-privilege RBAC, deploy
   it in-cluster through the same CI + GitOps pattern already used for
   `app/`.
   Concept: running a controller for real, RBAC scoping.

## Module B — The same shape of problem, the Crossplane way

5. **Install Crossplane + a provider** — Crossplane core via Helm,
   `provider-family-gcp` + `provider-gcp-storage`, and a `ProviderConfig`
   authenticated via GKE Workload Identity (keyless, same rationale as the
   GitHub Actions WIF setup from Part 1).
   Concept: a provider is a fleet of controllers you didn't have to write.

6. **XRD** — define `XObjectStore` / `ObjectStore`. Compare the
   schema-authoring experience directly against hand-writing
   `WebhookTarget`'s CRD in Lesson 1.
   Concept: an XRD *is* a CRD under the hood — just generated for you.

7. **Composition (patch-and-transform)** — map `XObjectStore` onto a real
   `Bucket` managed resource, apply a claim, and watch an actual GCS bucket
   get created.
   Concept: this *is* the reconcile loop from Lessons 2-3, except
   declarative, driven by Crossplane's generic reconciler instead of code
   you wrote.

8. **(Stretch) Composition Function in Go** — replace Lesson 7's YAML
   patch-and-transform with real Go (`function-sdk-go`) performing the same
   mapping.
   Concept: even Crossplane's declarative layer can drop back to
   hand-written Go when the logic gets complex — ties directly back to
   Module A.

## How we're working through this

One lesson at a time. Each lesson ends with a concrete, real check against
the actual cluster before moving to the next — not a code dump followed by
"does this look right?".
