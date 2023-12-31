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
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logr "sigs.k8s.io/controller-runtime/pkg/log"

	kokabieliv1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
)

// DataProcessReconciler reconciles a DataProcess object
type DataProcessReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kokabie.li,resources=dataprocesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kokabie.li,resources=dataprocesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kokabie.li,resources=dataprocesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DataProcessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	log.Info("Reconciling DataProcess", "dataprocess", req.NamespacedName)

	dataProcess := &kokabieliv1alpha1.DataProcess{}
	err := r.Get(ctx, req.NamespacedName, dataProcess)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("DataProcess resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch DataProcess")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dataInterfaces := &kokabieliv1alpha1.DataInterfaceList{}
	err = r.List(ctx, dataInterfaces)
	if err != nil {
		log.Error(err, "unable to list DataInterfaces")
		return ctrl.Result{}, err
	}

	availableReferences := map[string]bool{}
	for _, item := range dataInterfaces.Items {
		if item.Status.UsedReference != "" {
			availableReferences[item.Status.UsedReference] = true
		}
	}

	var notFound []string

	for _, input := range dataProcess.Spec.Inputs {
		if _, available := availableReferences[input.BuildTargetReference(dataProcess.Namespace)]; !available {
			notFound = append(notFound, input.BuildTargetReference(dataProcess.Namespace))
		}
	}
	for _, output := range dataProcess.Spec.Outputs {
		if _, available := availableReferences[output.BuildTargetReference(dataProcess.Namespace)]; !available {
			notFound = append(notFound, output.BuildTargetReference(dataProcess.Namespace))
		}
	}

	sort.Strings(notFound)

	if !reflect.DeepEqual(dataProcess.Status.MissingDataInterfaces, notFound) {
		err = r.Get(ctx, req.NamespacedName, dataProcess)
		if err != nil {
			log.Error(err, "unable to re-fetch DataProcess")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if !reflect.DeepEqual(dataProcess.Status.MissingDataInterfaces, notFound) {
			dataProcess.Status.MissingDataInterfaces = notFound
			err = r.Status().Update(ctx, dataProcess)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if !dataProcess.Status.Loaded {
		err = r.Get(ctx, req.NamespacedName, dataProcess)
		if err != nil {
			log.Error(err, "unable to re-fetch DataProcess")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		dataProcess.Status.Loaded = true
		err = r.Status().Update(ctx, dataProcess)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataProcessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kokabieliv1alpha1.DataProcess{}).
		Watches(&kokabieliv1alpha1.DataInterface{}, handler.EnqueueRequestsFromMapFunc(r.requeAffectedProcessors)).
		Complete(r)
}

func (r *DataProcessReconciler) requeAffectedProcessors(_ context.Context, object client.Object) []reconcile.Request {
	var ret []reconcile.Request

	dataInterface := object.(*kokabieliv1alpha1.DataInterface)

	if dataInterface.Status.UsedReference == "" {
		return ret
	}
	usedInDataProcesses, err := getProcessesForInterface(context.Background(), r.Client, dataInterface.Status.UsedReference)
	if err != nil {
		return ret
	}
	for _, process := range usedInDataProcesses {
		ret = append(ret, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: process.Namespace,
				Name:      process.Name,
			},
		})
	}
	return ret
}
