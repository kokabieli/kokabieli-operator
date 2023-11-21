/*
Copyright 2023 Florian Schrag.

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
	kokabieliv1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/examples/configfile/custom/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randChars(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// DataSetReconciler reconciles a DataSet object
type DataSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kokabie.li,resources=datasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kokabie.li,resources=datasets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kokabie.li,resources=datasets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DataSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	clog := log.FromContext(ctx)

	set := &kokabieliv1alpha1.DataSet{}
	err := r.Get(ctx, req.NamespacedName, set)
	if err != nil {
		if apierrors.IsNotFound(err) {
			clog.Info("DataSet not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		clog.Error(err, "Failed to get DataSet")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if set.Status.Conditions == nil || len(set.Status.Conditions) == 0 {
		meta.SetStatusCondition(&set.Status.Conditions, metav1.Condition{Type: "Available", Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, set); err != nil {
			clog.Error(err, "Failed to update DataSet status")
			return ctrl.Result{}, err
		}

		if err := r.Get(ctx, req.NamespacedName, set); err != nil {
			clog.Error(err, "Failed to re-fetch set")
			return ctrl.Result{}, err
		}
	}

	// check for duplicate interfaces and processes
	for i, iSpec := range set.Spec.Interfaces {
		for j, jSpec := range set.Spec.Interfaces {
			if i != j && iSpec.Name == jSpec.Name {
				meta.SetStatusCondition(&set.Status.Conditions, metav1.Condition{Type: "Degraded", Status: metav1.ConditionFalse, Reason: "DuplicateInterfaces", Message: fmt.Sprintf("Duplicate interface name %s", iSpec.Name)})
				if err := r.Status().Update(ctx, set); err != nil {
					clog.Error(err, "Failed to update DataSet status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}
		}
	}
	for i, iSpec := range set.Spec.Processes {
		for j, jSpec := range set.Spec.Processes {
			if i != j && iSpec.Name == jSpec.Name {
				meta.SetStatusCondition(&set.Status.Conditions, metav1.Condition{Type: "Degraded", Status: metav1.ConditionFalse, Reason: "DuplicateProcesses", Message: fmt.Sprintf("Duplicate process name %s", iSpec.Name)})
				if err := r.Status().Update(ctx, set); err != nil {
					clog.Error(err, "Failed to update DataSet status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			}
		}
	}

	// Update or create the Interfaces
	for _, iSpec := range set.Spec.Interfaces {
		iref, found := set.Status.Interfaces[iSpec.Name]
		if !found {
			if set.Status.Interfaces == nil {
				set.Status.Interfaces = make(map[string]corev1.ObjectReference)
			}
			iref = corev1.ObjectReference{
				Kind:       "DataInterface",
				APIVersion: v1alpha1.GroupVersion.String(),
				Name:       req.Name + "-" + randChars(5),
				Namespace:  req.Namespace,
			}
			set.Status.Interfaces[iSpec.Name] = iref
			if err := r.Status().Update(ctx, set); err != nil {
				clog.Error(err, "Failed to update DataSet status")
				return ctrl.Result{}, err
			}
			if err := r.Get(ctx, req.NamespacedName, set); err != nil {
				clog.Error(err, "Failed to re-fetch set")
				return ctrl.Result{}, err
			}
		}
		intf := &kokabieliv1alpha1.DataInterface{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: iref.Namespace, Name: iref.Name}, intf); err != nil {
			if apierrors.IsNotFound(err) {
				intf = &kokabieliv1alpha1.DataInterface{
					ObjectMeta: metav1.ObjectMeta{
						Name:      iref.Name,
						Namespace: iref.Namespace,
						Labels:    mergeLabels(iSpec.Labels, set.Labels),
					},
					Spec: kokabieliv1alpha1.DataInterfaceSpec{
						Name:        iSpec.Name,
						Reference:   iSpec.Reference,
						Type:        iSpec.Type,
						Description: iSpec.Description,
					},
				}
				if err := ctrl.SetControllerReference(set, intf, r.Scheme); err != nil {
					clog.Error(err, "Failed to set controller reference")
					return ctrl.Result{}, err
				}
				if err := r.Create(ctx, intf); err != nil {
					clog.Error(err, "Failed to create DataInterface")
					return ctrl.Result{}, err
				}
			} else {
				clog.Error(err, "Failed to get DataInterface")
				return ctrl.Result{}, err
			}
		}

		intf.Spec.Name = iSpec.Name
		intf.Spec.Type = iSpec.Type
		intf.Spec.Reference = iSpec.Reference
		intf.Spec.Description = iSpec.Description
		intf.Spec.Namespaced = iSpec.Namespaced || set.Spec.Namespaced
		intf.Labels = mergeLabels(iSpec.Labels, set.Labels)
		if err := r.Update(ctx, intf); err != nil {
			clog.Error(err, "Failed to update DataInterface")
			return ctrl.Result{}, err
		}
	}

	// Update or create the Processes
	for _, pSpec := range set.Spec.Processes {
		pref, found := set.Status.Processes[pSpec.Name]
		if !found {
			if set.Status.Processes == nil {
				set.Status.Processes = make(map[string]corev1.ObjectReference)
			}
			pref = corev1.ObjectReference{
				Kind:       "DataProcess",
				APIVersion: v1alpha1.GroupVersion.String(),
				Name:       req.Name + "-" + randChars(5),
				Namespace:  req.Namespace,
			}
			set.Status.Processes[pSpec.Name] = pref
			if err := r.Status().Update(ctx, set); err != nil {
				clog.Error(err, "Failed to update DataSet status")
				return ctrl.Result{}, err
			}
			if err := r.Get(ctx, req.NamespacedName, set); err != nil {
				clog.Error(err, "Failed to re-fetch set")
				return ctrl.Result{}, err
			}
		}
		intf := &kokabieliv1alpha1.DataProcess{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: pref.Namespace, Name: pref.Name}, intf); err != nil {
			if apierrors.IsNotFound(err) {
				intf = &kokabieliv1alpha1.DataProcess{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pref.Name,
						Namespace: pref.Namespace,
						Labels:    mergeLabels(pSpec.Labels, set.Labels),
					},
					Spec: kokabieliv1alpha1.DataProcessSpec{
						Name:        pSpec.Name,
						Type:        pSpec.Type,
						Description: pSpec.Description,
						Inputs:      pSpec.Inputs,
						Outputs:     pSpec.Outputs,
					},
				}
				if err := ctrl.SetControllerReference(set, intf, r.Scheme); err != nil {
					clog.Error(err, "Failed to set controller reference")
					return ctrl.Result{}, err
				}
				if err := r.Create(ctx, intf); err != nil {
					clog.Error(err, "Failed to create DataProcess")
					return ctrl.Result{}, err
				}
			} else {
				clog.Error(err, "Failed to get DataProcess")
				return ctrl.Result{}, err
			}
		}

		intf.Spec.Name = pSpec.Name
		intf.Spec.Type = pSpec.Type
		intf.Spec.Description = pSpec.Description
		intf.Spec.Inputs = pSpec.Inputs
		intf.Spec.Outputs = pSpec.Outputs
		intf.Labels = mergeLabels(pSpec.Labels, set.Labels)
		if err := r.Update(ctx, intf); err != nil {
			clog.Error(err, "Failed to update DataProcess")
			return ctrl.Result{}, err
		}
	}

	// yay success
	meta.SetStatusCondition(&set.Status.Conditions, metav1.Condition{Type: "Available",
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Nested DataInterfaces and DataProcesses successfully created")})

	if err := r.Status().Update(ctx, set); err != nil {
		clog.Error(err, "Failed to update DataSet status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func mergeLabels(primaryLabels map[string]string, secondaryLabels map[string]string) map[string]string {
	if primaryLabels == nil {
		return secondaryLabels
	}
	if secondaryLabels == nil {
		return primaryLabels
	}
	labels := make(map[string]string)
	for k, v := range secondaryLabels {
		labels[k] = v
	}
	// primary overwrites secondary
	for k, v := range primaryLabels {
		labels[k] = v
	}
	return labels
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kokabieliv1alpha1.DataSet{}).
		Complete(r)
}
