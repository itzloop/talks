# Kubernetes Controllers

## Build a simple controller

- Working with kubebuilder
```bash
mkdir -p linux-fest-controller && cd linux-fest-controller
kubebuilder init --domain=example.com --repo=github.com/itzloop/linux-fest-controller
kubebuilder create api --group=aut --version=v2025 --kind=Sample
```

- Installing CRDs

```bash
make manifests
make install
```

- Running the controller
```bash
make run
```

- Copy this to reconcile code
```go
func (r *MyKindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the MyKind resource
	var mykind autv2025.MyKind
	if err := r.Get(ctx, req.NamespacedName, &mykind); err != nil {
		// The resource may have been deleted after the reconcile request.
		// Return and don't requeue.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 🐾 Log something useless — just to show that Reconcile is triggered
	log.Info("Reconcile loop running", "name", mykind.Name, "namespace", mykind.Namespace)

	// 🔁 Don’t requeue — this controller literally does nothing
	return ctrl.Result{}, nil
}
```

- Apply fluffy

```bash
k apply -f something
```

- Cleanup
```bash
k delete -f fluffy.yaml
make uninstall
```

## Showcasing the demo controller


