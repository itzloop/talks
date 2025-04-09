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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	// ignore resouceVersion changes since we want to decay love and food
	// at .spec.decayInterval intervals
	if time.Since(pet.Status.ModifiedTime.Time) < pet.Spec.DecayInterval.Duration {
		return ctrl.Result{}, nil
	}

	fmt.Println("Reconciling", pet.Name, "gen:", pet.Generation, "rv:", pet.ResourceVersion)

	if pet.Status.Food == 0 && pet.Status.Love == 0 {
		return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
	}

	for range 10 {
		petCopy := pet.DeepCopy()
		if petCopy.Status.Food > petCopy.Spec.FoodDecayRate {
			petCopy.Status.Food -= petCopy.Spec.FoodDecayRate
		} else {
			petCopy.Status.Food = 0
		}

		if petCopy.Status.Love > petCopy.Spec.LoveDecayRate {
			petCopy.Status.Love -= petCopy.Spec.LoveDecayRate
		} else {
			petCopy.Status.Love = 0
		}

		petCopy.Status.ModifiedTime = v1.Now()

		if err := r.Status().Update(ctx, petCopy); err != nil {
			log.Error(err, "unable to update status")
			if errors.IsConflict(err) {
				if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &pet); err != nil {
					return ctrl.Result{}, client.IgnoreNotFound(err)
				}

				continue
			}
		}

		fmt.Println(petCopy.Name, petCopy.Namespace, petCopy.Spec.DecayInterval, petCopy.Status)
		return ctrl.Result{RequeueAfter: petCopy.Spec.DecayInterval.Duration}, nil
	}

	// log.Info("requeuing pet", "pet_name", pet.Spec.Nickname, "requeue_after", pet.Spec.DecayInterval)
	return ctrl.Result{RequeueAfter: pet.Spec.DecayInterval.Duration}, nil
	// var petList linuxfestv2025.PetList
	// if err := r.List(ctx, &petList); err != nil {
	// 	return ctrl.Result{}, client.IgnoreNotFound(err)
	// }

	// now := metav1.Now()
	// var minInterval time.Duration = math.MaxInt64
	// for _, pet := range petList.Items {

	// 	minInterval = min(minInterval, pet.Spec.DecayInterval.Duration)

	// 	// Create a copy of the pet for processing
	// 	petCopy := pet.DeepCopy()

	// 	// Skip if pet is already dead
	// 	if petCopy.Status.IsDead {
	// 		if petCopy.Status.Food != 0 {
	// 			petCopy.Status.IsDead = false
	// 		} else {
	// 			continue
	// 		}
	// 	} else {
	// 		if petCopy.Status.Food == 0 && petCopy.Status.Love == 0 {
	// 			petCopy.Status.Love = 100
	// 		}

	// 		if petCopy.Status.Food == 0 {
	// 			petCopy.Status.Food = 100
	// 		}

	// 	}

	// 	// Check if we need to decay food and love
	// 	// Only decay if the pet hasn't been modified recently
	// 	if petCopy.Status.ModifiedTime.IsZero() || time.Since(petCopy.Status.ModifiedTime.Time) >= pet.Spec.DecayInterval.Duration {
	// 		// log.Info("modified time", "modified_time", petCopy.Status.ModifiedTime)
	// 		// Decay food
	// 		if petCopy.Status.Food > ZeroThreshold {
	// 			petCopy.Status.Food -= petCopy.Spec.FoodDecayRate
	// 			if petCopy.Status.Food < ZeroThreshold {
	// 				petCopy.Status.Food = ZeroThreshold
	// 			}
	// 			// controllerModified = true
	// 		}

	// 		// Decay love
	// 		if petCopy.Status.Love > ZeroThreshold {
	// 			petCopy.Status.Love -= petCopy.Spec.LoveDecayRate
	// 			if petCopy.Status.Love < ZeroThreshold {
	// 				petCopy.Status.Love = ZeroThreshold
	// 			}
	// 			// controllerModified = true
	// 		}

	// 		// Update ModifiedTime if controller made changes
	// 		petCopy.Status.ModifiedTime = now
	// 		// // Check for attention needed events
	// 		// if petCopy.Spec.Food <= LowThreshold || petCopy.Spec.Love <= LowThreshold {
	// 		// 	log.Info("Pet needs attention (low resources)",
	// 		// 		"pet", petCopy.Spec.Name,
	// 		// 		"food", petCopy.Spec.Food,
	// 		// 		"love", petCopy.Spec.Love)
	// 		// } else if petCopy.Spec.Food <= MediumThreshold || petCopy.Spec.Love <= MediumThreshold {
	// 		// 	log.Info("Pet needs attention (medium resources)",
	// 		// 		"pet", petCopy.Spec.Name,
	// 		// 		"food", petCopy.Spec.Food,
	// 		// 		"love", petCopy.Spec.Love)
	// 		// } else if petCopy.Spec.Food <= HighThreshold || petCopy.Spec.Love <= HighThreshold {
	// 		// 	log.Info("Pet needs attention (high resources)",
	// 		// 		"pet", petCopy.Spec.Name,
	// 		// 		"food", petCopy.Spec.Food,
	// 		// 		"love", petCopy.Spec.Love)
	// 		// }
	// 	}

	// 	// Check for death
	// 	if petCopy.Status.Food > ZeroThreshold {
	// 		petCopy.Status.IsDead = false
	// 	} else if petCopy.Status.Food == ZeroThreshold && petCopy.Status.Love == ZeroThreshold {
	// 		petCopy.Status.IsDead = true
	// 		log.Info("Pet has died",
	// 			"pet", petCopy.Name,
	// 			"nickname", petCopy.Spec.Nickname,
	// 			"food", petCopy.Status.Food,
	// 			"love", petCopy.Status.Love)
	// 	}
	// 	// Update the pet in the cluster if changes were made
	// 	// if controllerModified || petCopy.Status.IsDead {
	// 	// 	if err := r.Status().Update(ctx, petCopy); err != nil {
	// 	// 		log.Error(err, "unable to update Pet")
	// 	// 		return ctrl.Result{}, err
	// 	// 	}
	// 	// }

	// 	name, ns := petCopy.Name, petCopy.Namespace
	// 	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	// 		var pet v2025.Pet
	// 		err := r.Get(ctx, types.NamespacedName{
	// 			Name:      name,
	// 			Namespace: ns,
	// 		}, &pet)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		return r.Status().Update(ctx, &pet)
	// 	})
	// 	if err != nil {
	// 		log.Error(err, "unable to update status")
	// 		continue
	// 	}
	// 	for range 10 {
	// 		if err := r.Status().Update(ctx, petCopy); err != nil {
	// 			log.Error(err, "unable to update status")
	// 			if errors.IsConflict(err) {
	// 				continue
	// 			}
	// 		}

	// 		break
	// 	}

	// }

	// Check if food or love was modified by another source
	// If so, we don't update ModifiedTime
	// controllerModified := false

	// Requeue after interval
	// return ctrl.Result{RequeueAfter: minInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&linuxfestv2025.Pet{}).
		Named("pet").
		Complete(r)
}
