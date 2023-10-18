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
	kokabieliv1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// ConstellationReconciler reconciles a Constellation object
type ConstellationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kokabie.li,resources=constellations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kokabie.li,resources=constellations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kokabie.li,resources=constellations/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ConstellationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)

	log.Info("Reconciling Constellation", "constellation", req.NamespacedName)
	constellation := &kokabieliv1alpha1.Constellation{}
	err := r.Get(ctx, req.NamespacedName, constellation)

	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Constellation not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Constellation")
		return ctrl.Result{}, err
	}

	constellationResult := &kokabieliv1alpha1.ConstellationResult{
		Name:              constellation.Spec.Name,
		LastUpdated:       metav1.Time{Time: time.Now()},
		Description:       asString(constellation.Spec.Description),
		DataInterfaceList: []kokabieliv1alpha1.ConstellationInterface{},
		DataProcessList:   []kokabieliv1alpha1.ConstellationDataProcess{},
	}

	if constellation.Spec.Filters == nil || len(constellation.Spec.Filters) == 0 {
		var opts []client.ListOption
		err = r.fetch(ctx, constellationResult, opts)
		if err != nil {
			log.Error(err, "Failed to fetch data interfaces and data processes")
			return ctrl.Result{}, err
		}
	} else {
		for _, filter := range constellation.Spec.Filters {
			log.Info("Filter", "filter", filter)

			var opts []client.ListOption
			if filter.Labels != nil && len(filter.Labels) > 0 {
				opts = append(opts, client.MatchingLabels(filter.Labels))
			}

			if filter.Namespaces != nil && len(filter.Namespaces) > 0 {
				for _, namespace := range filter.Namespaces {
					newOpts := append(opts, client.InNamespace(namespace))
					err = r.fetch(ctx, constellationResult, newOpts)
					if err != nil {
						log.Error(err, "Failed to fetch data interfaces and data processes")
						return ctrl.Result{}, err
					}
				}
			} else {
				err = r.fetch(ctx, constellationResult, opts)
				if err != nil {
					log.Error(err, "Failed to fetch data interfaces and data processes")
					return ctrl.Result{}, err
				}
			}
		}
	}
	constellationResult.GenerateMissingInterfaces()

	err = r.Get(ctx, req.NamespacedName, constellation)
	if err != nil {
		log.Error(err, "Failed to re-fetch Constellation")
		return ctrl.Result{}, err
	}
	constellation.Status.ConstellationResult = constellationResult

	if err := r.Status().Update(ctx, constellation); err != nil {
		log.Error(err, "Failed to update Constellation status")
		return ctrl.Result{}, err
	}

	configMap := &corev1.ConfigMap{}
	index := make(map[string]string)
	err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: constellation.Spec.TargetConfigMap}, configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ConfigMap not found, creating new one")
			configMap.Name = constellation.Spec.TargetConfigMap
			configMap.Namespace = req.Namespace
			configMap.Data = map[string]string{
				"index.json": "{}",
			}
			err = r.Create(ctx, configMap)
			if err != nil {
				log.Error(err, "Failed to create ConfigMap")
				return ctrl.Result{}, err
			}
			err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: constellation.Spec.TargetConfigMap}, configMap)
			if err != nil {
				log.Error(err, "Failed to re-fetch ConfigMap")
				return ctrl.Result{}, err
			}
		} else {
			log.Error(err, "Failed to get ConfigMap")
			return ctrl.Result{}, err
		}
	}

	err = json.Unmarshal([]byte(configMap.Data["index.json"]), &index)
	if err != nil {
		log.Error(err, "Failed to unmarshal index.json - assume a default empty index.json")
		index = make(map[string]string)
	}
	index[constellation.Name+".json"] = constellation.Spec.Name
	data, err := json.Marshal(index)
	if err != nil {
		log.Error(err, "Failed to marshal index.json")
		return ctrl.Result{}, err
	}
	configMap.Data["index.json"] = string(data)
	constellationData, err := json.Marshal(constellationResult)
	if err != nil {
		log.Error(err, "Failed to marshal constellation.json")
		return ctrl.Result{}, err
	}
	configMap.Data[constellation.Name+".json"] = string(constellationData)
	err = r.Update(ctx, configMap)

	return ctrl.Result{}, nil
}

func asString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *ConstellationReconciler) fetch(ctx context.Context, constellationResult *kokabieliv1alpha1.ConstellationResult, filterOpts []client.ListOption) error {
	log := logr.FromContext(ctx)

	dataInterfaceList := &kokabieliv1alpha1.DataInterfaceList{}
	dataProcessList := &kokabieliv1alpha1.DataProcessList{}
	err := r.List(ctx, dataInterfaceList, filterOpts...)
	if err != nil {
		log.Error(err, "Failed to list DataInterfaces")
		return err
	}
	err = r.List(ctx, dataProcessList, filterOpts...)
	if err != nil {
		log.Error(err, "Failed to list DataProcesses")
		return err
	}
	constellationResult.AddDataInterfaceList(log, dataInterfaceList.Items)
	constellationResult.AddDataProcessList(log, dataProcessList.Items)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConstellationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kokabieliv1alpha1.Constellation{}).
		Watches(&kokabieliv1alpha1.DataInterface{}, handler.EnqueueRequestsFromMapFunc(r.requeueAllConstellations)).
		Watches(&kokabieliv1alpha1.DataProcess{}, handler.EnqueueRequestsFromMapFunc(r.requeueAllConstellations)).
		Complete(r)
}

func (r *ConstellationReconciler) requeueAllConstellations(_ context.Context, _ client.Object) []reconcile.Request {
	var ret []reconcile.Request
	list := &kokabieliv1alpha1.ConstellationList{}
	err := r.List(context.Background(), list)
	if err != nil {
		return nil
	}
	for _, item := range list.Items {
		ret = append(ret, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.Name,
				Namespace: item.Namespace,
			},
		})
	}
	return ret
}
