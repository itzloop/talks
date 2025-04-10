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

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
}

// +kubebuilder:rbac:groups=linuxfest.example.com,resources=pets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=linuxfest.example.com,resources=pets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=linuxfest.example.com,resources=pets/finalizers,verbs=update

// Constants for pet state thresholds
const (
	LowThreshold    = 50
	MediumThreshold = 30
	HighThreshold   = 10
	ZeroThreshold   = 0
	DecayAmount     = 10
	RequeueInterval = 10 * time.Second
)

// TODO: Fix an issue when love and food is different
// Reconcile is part of the main kubernetes reconciliation loop
func (r *PetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	print("hello")
	var pet linuxfestv2025.Pet
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &pet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !pet.Status.Initialized && pet.Status.Food == 0 && pet.Status.Love == 0 {
		fmt.Println("init", pet.Status.Initialized, pet.Status.Food, pet.Status.Love)
		for range 10 {
			petCopy := pet.DeepCopy()
			petCopy.Status.Food = 100
			petCopy.Status.Love = 100
			pet.Status.ModifiedTime = v1.Now()
			petCopy.Status.Initialized = true

			if err := r.Status().Update(ctx, petCopy); err != nil {
				log.Error(err, "unable to update status")
				if errors.IsConflict(err) {
					if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &pet); err != nil {
						return ctrl.Result{}, client.IgnoreNotFound(err)
					}

					continue
				} else if errors.IsNotFound(err) {
					return ctrl.Result{}, nil
				}

				return ctrl.Result{RequeueAfter: petCopy.Spec.DecayInterval.Duration}, err
			}

			return ctrl.Result{RequeueAfter: petCopy.Spec.DecayInterval.Duration}, nil
		}

	}

	fmt.Println("Reconciling", pet.Name, "gen:", pet.Generation, "rv:", pet.ResourceVersion)
	// ignore resouceVersion changes since we want to decay love and food
	// at .spec.decayInterval intervals
	_, feedAnnot := pet.Annotations["linuxfest.example.com/feed"]
	_, petAnnot := pet.Annotations["linuxfest.example.com/pet"]

	if !feedAnnot && !petAnnot && time.Since(pet.Status.ModifiedTime.Time) < pet.Spec.DecayInterval.Duration {
		return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
	}

	// if pet.Status.Food == 0 && pet.Status.Love == 0 {
	// 	return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
	// }

	// handle food and love increase by annotation
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

		return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, err
	}

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

		return r.Status().Update(ctx, cpy)
	})
	return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, err

}

// SetupWithManager sets up the controller with the Manager.
func (r *PetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&linuxfestv2025.Pet{}).
		Named("pet").
		Complete(r)
}
