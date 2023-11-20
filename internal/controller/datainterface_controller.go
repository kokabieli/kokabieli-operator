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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sort"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logr "sigs.k8s.io/controller-runtime/pkg/log"

	kokabieliv1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
)

// DataInterfaceReconciler reconciles a DataInterface object
type DataInterfaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kokabie.li,resources=datainterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kokabie.li,resources=datainterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kokabie.li,resources=datainterfaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DataInterfaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	log.Info("Reconciling DataInterface", "datainterface", req.NamespacedName)

	dataInterface := &kokabieliv1alpha1.DataInterface{}
	err := r.Get(ctx, req.NamespacedName, dataInterface)

	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Interface not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Interface")
		return ctrl.Result{}, err
	}

	ref := dataInterfaceReference(dataInterface)
	if dataInterface.Status.UsedReference != ref {
		err = r.Get(ctx, req.NamespacedName, dataInterface)
		if err != nil {
			log.Error(err, "Failed to refetch Interface")
			return ctrl.Result{}, err
		}
		dataInterface.Status.UsedReference = ref
		err = r.Status().Update(ctx, dataInterface)
		if err != nil {
			log.Error(err, "Failed to update Interface status")
			return ctrl.Result{}, err
		}
	}

	usedInDataProcesses, err := getProcessesForInterface(ctx, r.Client, ref)
	if err != nil {
		log.Error(err, "Failed to get processes for Interface")
		return ctrl.Result{}, err
	}

	if !equalNamespacedNames(dataInterface.Status.UsedInDataProcesses, usedInDataProcesses) {
		err = r.Get(ctx, req.NamespacedName, dataInterface)
		if err != nil {
			log.Error(err, "Failed to refetch Interface")
			return ctrl.Result{}, err
		}
		if !equalNamespacedNames(dataInterface.Status.UsedInDataProcesses, usedInDataProcesses) {
			dataInterface.Status.UsedInDataProcesses = usedInDataProcesses
			err = r.Status().Update(ctx, dataInterface)
			if err != nil {
				log.Error(err, "Failed to update Interface status")
				return ctrl.Result{}, err
			}
		}
	}

	err = r.checkForDuplicates(ctx, dataInterface)
	if err != nil {
		log.Error(err, "Failed to check for duplicates")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DataInterfaceReconciler) checkForDuplicates(ctx context.Context, di *kokabieliv1alpha1.DataInterface) error {
	log := logr.FromContext(ctx)
	allInterfaces := &kokabieliv1alpha1.DataInterfaceList{}
	err := r.List(ctx, allInterfaces)
	if err != nil {
		log.Error(err, "Failed to list all interfaces")
		return err
	}
	// Check for duplicate references
	for _, dataInterface := range allInterfaces.Items {
		if dataInterface.Status.UsedReference == "" {
			continue
		}
		if di.Status.UsedReference == dataInterface.Status.UsedReference && (dataInterface.Namespace != di.Namespace || dataInterface.Name != di.Name) {

			condition := metav1.Condition{
				Type:    "UniqueReference",
				Status:  metav1.ConditionFalse,
				Reason:  "failed_duplicate_check",
				Message: "DuplicateReference for object: " + dataInterface.Namespace + "/" + dataInterface.Name,
			}
			if !meta.IsStatusConditionFalse(di.Status.Conditions, condition.Type) {
				err = r.Get(ctx, types.NamespacedName{Namespace: di.Namespace, Name: di.Name}, di)
				if err != nil {
					log.Error(err, "Failed to refetch Interface")
					return err
				}
				log.Error(err, "Duplicate reference",
					"reference", di.Status.UsedReference,
					"object", dataInterface.Namespace+"/"+dataInterface.Name)
				meta.SetStatusCondition(&di.Status.Conditions, condition)
				err = r.Status().Update(ctx, di)
				if err != nil {
					log.Error(err, "Failed to update Interface status")
					return err
				}
				return nil
			}
			return nil
		}
	}

	condition := metav1.Condition{
		Type:    "UniqueReference",
		Status:  metav1.ConditionTrue,
		Reason:  "successful_duplicate_check",
		Message: "UniqueReference " + di.Status.UsedReference + " is unique",
	}
	if !meta.IsStatusConditionTrue(di.Status.Conditions, condition.Type) {
		if !meta.IsStatusConditionTrue(di.Status.Conditions, condition.Type) {

			err = r.Get(ctx, types.NamespacedName{Namespace: di.Namespace, Name: di.Name}, di)
			if err != nil {
				log.Error(err, "Failed to refetch Interface")
				return err
			}

			meta.SetStatusCondition(&di.Status.Conditions, condition)
			err = r.Status().Update(ctx, di)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func getProcessesForInterface(ctx context.Context, c client.Client, ref string) ([]kokabieliv1alpha1.NamespacedName, error) {

	dataProcesses := &kokabieliv1alpha1.DataProcessList{}
	err := c.List(ctx, dataProcesses)
	if err != nil {
		return nil, err
	}
	var usedInDataProcesses []kokabieliv1alpha1.NamespacedName
	for _, dataProcess := range dataProcesses.Items {
		refName := kokabieliv1alpha1.NamespacedName{Namespace: dataProcess.Namespace, Name: dataProcess.Name}
		for _, input := range dataProcess.Spec.Inputs {
			if input.BuildTargetReference(dataProcess.Namespace) == ref {
				usedInDataProcesses = append(usedInDataProcesses, refName)
			}
		}
		for _, output := range dataProcess.Spec.Outputs {
			if output.BuildTargetReference(dataProcess.Namespace) == ref {
				usedInDataProcesses = append(usedInDataProcesses, refName)
			}
		}
	}
	sort.Slice(usedInDataProcesses, func(i, j int) bool {
		if usedInDataProcesses[i].Namespace != usedInDataProcesses[j].Namespace {
			return usedInDataProcesses[i].Namespace < usedInDataProcesses[j].Namespace
		}
		return usedInDataProcesses[i].Name < usedInDataProcesses[j].Name
	})
	return usedInDataProcesses, nil
}

func equalNamespacedNames(processes []kokabieliv1alpha1.NamespacedName, processes2 []kokabieliv1alpha1.NamespacedName) bool {
	if len(processes) != len(processes2) {
		return false
	}
	for i, process := range processes {
		if process != processes2[i] {
			return false
		}
	}
	return true
}

func dataInterfaceReference(dataInterface *kokabieliv1alpha1.DataInterface) string {
	if dataInterface.Spec.Reference != nil {
		if dataInterface.Spec.Namespaced {
			return dataInterface.Namespace + "/" + *dataInterface.Spec.Reference
		}
		return *dataInterface.Spec.Reference
	}
	return dataInterface.Namespace + "/" + dataInterface.Name
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kokabieliv1alpha1.DataInterface{}).
		Complete(r)
}
