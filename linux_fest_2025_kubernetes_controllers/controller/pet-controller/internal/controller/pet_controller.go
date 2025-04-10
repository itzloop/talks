/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	linuxfestv2025 "github.com/itzloop/pet-controller/api/v2025"
)

// PetReconciler reconciles a Pet object
type PetReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Recorder record.EventRecorder // 👈 Add this
}

// +kubebuilder:rbac:groups=linuxfest.example.com,resources=pets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=linuxfest.example.com,resources=pets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=linuxfest.example.com,resources=pets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop
func (r *PetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 🐾 Fetch the Pet resource
	var pet linuxfestv2025.Pet
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &pet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 🐣 First-time initialization (100 food + love)
	if !pet.Status.Initialized && pet.Status.Food == 0 && pet.Status.Love == 0 {
		fmt.Println("init", pet.Status.Initialized, pet.Status.Food, pet.Status.Love)

		for range 10 {
			petCopy := pet.DeepCopy()
			petCopy.Status.Food = 100
			petCopy.Status.Love = 100
			pet.Status.ModifiedTime = v1.Now()
			petCopy.Status.Initialized = true

			// 💾 Save initial state
			if err := r.Status().Update(ctx, petCopy); err != nil {
				log.Error(err, "unable to update status")

				if errors.IsConflict(err) {
					// 🔁 Retry if needed
					if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &pet); err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}
					continue
				} else if errors.IsNotFound(err) {
					return ctrl.Result{}, nil
				}

				return ctrl.Result{RequeueAfter: petCopy.Spec.DecayInterval.Duration}, err
			}

			// 🕐 Schedule next decay
			return ctrl.Result{RequeueAfter: petCopy.Spec.DecayInterval.Duration}, nil
		}
	}

	// 🔍 Log the reconcile trigger
	fmt.Println("Reconciling", pet.Name, "gen:", pet.Generation, "rv:", pet.ResourceVersion)

	// 🎯 Check if we should skip reconcile (not enough time passed, no annotations)
	_, feedAnnot := pet.Annotations["linuxfest.example.com/feed"]
	_, petAnnot := pet.Annotations["linuxfest.example.com/pet"]
	if !feedAnnot && !petAnnot && time.Since(pet.Status.ModifiedTime.Time) < pet.Spec.DecayInterval.Duration {
		return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
	} else if !feedAnnot && !petAnnot && time.Since(pet.Status.ModifiedTime.Time) >= pet.Spec.DecayInterval.Duration {
		if pet.Status.Food == 0 {
			return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
		}
	}

	// 🧃 Handle annotation-based feeding/petting
	if feedAnnot || petAnnot {
		var (
			foodDelta, petDelta int
			err                 error
		)

		if feedAnnot {
			foodDelta, err = strconv.Atoi(pet.Annotations["linuxfest.example.com/feed"])
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		if petAnnot {
			petDelta, err = strconv.Atoi(pet.Annotations["linuxfest.example.com/pet"])
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// 🧹 Remove annotations after applying them
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Get(ctx, client.ObjectKey{Name: pet.Name, Namespace: pet.Namespace}, &pet); err != nil {
				return client.IgnoreNotFound(err)
			}
			cpy := pet.DeepCopy()
			delete(cpy.Annotations, "linuxfest.example.com/feed")
			delete(cpy.Annotations, "linuxfest.example.com/pet")
			return r.Update(ctx, cpy)
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		if foodDelta <= 0 && petDelta <= 0 {
			return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
		}

		// 💖 Update status fields with feed/pet deltas
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Get(ctx, client.ObjectKey{Name: pet.Name, Namespace: pet.Namespace}, &pet); err != nil {
				return client.IgnoreNotFound(err)
			}

			cpy := pet.DeepCopy()
			cpy.Status.Food += foodDelta
			if cpy.Status.Food > 100 {
				cpy.Status.Food = 100
			}
			cpy.Status.FedTime = v1.Now()

			cpy.Status.Love += petDelta
			if cpy.Status.Love > 100 {
				cpy.Status.Love = 100
			}
			cpy.Status.PetTime = v1.Now()

			return r.Status().Update(ctx, cpy)
		})

		// 🔁 Schedule next decay
		return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, err
	}

	// 🧓 Otherwise, apply decay to food and love over time
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, client.ObjectKey{Name: pet.Name, Namespace: pet.Namespace}, &pet); err != nil {
			return client.IgnoreNotFound(err)
		}

		cpy := pet.DeepCopy()
		if cpy.Status.Food > cpy.Spec.FoodDecayRate {
			cpy.Status.Food -= cpy.Spec.FoodDecayRate
		} else {
			cpy.Status.Food = 0
		}
		if cpy.Status.Love > cpy.Spec.LoveDecayRate {
			cpy.Status.Love -= cpy.Spec.LoveDecayRate
		} else {
			cpy.Status.Love = 0
		}

		cpy.Status.ModifiedTime = v1.Now()

		if cpy.Status.Food == 0 {
			r.Recorder.Event(cpy, corev1.EventTypeWarning, "Dead", fmt.Sprintf("☠️ %s died", cpy.Spec.Nickname))
			// r.Recorder.Event(cpy, corev1.EventTypeNormal, "Fed", fmt.Sprintf("🐾 Pet food decayed by %d. New food level: %d", cpy.Spec.FoodDecayRate, cpy.Status.Food))
		} else if cpy.Status.Love == 0 {
			r.Recorder.Event(cpy, corev1.EventTypeWarning, "NeedLove", fmt.Sprintf("😢 %s Needs Love and Attention", cpy.Spec.Nickname))
			// r.Recorder.Event(cpy, corev1.EventTypeNormal, "Fed", fmt.Sprintf("🐾 Pet food decayed by %d. New food level: %d", cpy.Spec.FoodDecayRate, cpy.Status.Food))
		} else if cpy.Status.Food < 30 {
			r.Recorder.Event(cpy, corev1.EventTypeWarning, "NeedFood", fmt.Sprintf("😭%s Needs Food", cpy.Spec.Nickname))
		}
		return r.Status().Update(ctx, cpy)
	})

	// 🔁 Requeue for next decay tick
	return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *PetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("pet-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&linuxfestv2025.Pet{}).
		Named("pet").
		Complete(r)
}
