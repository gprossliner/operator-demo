/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	groupv1 "github.com/gprossliner/operator-demo/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// RouterConfigReconciler reconciles a RouterConfig object
type RouterConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=group.test.org,resources=routerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=group.test.org,resources=routerconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=group.test.org,resources=routerconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RouterConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *RouterConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile RouteConfig")

	routeConfig := &groupv1.RouterConfig{}
	log.Info(fmt.Sprintf("<- GET RouteConfig"))
	err := r.Get(ctx, req.NamespacedName, routeConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("RouteConfig resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get RouteConfig")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("-> GET RouteConfig ResourceVersion=%s", routeConfig.ObjectMeta.ResourceVersion))

	// we just update some condition here
	meta.SetStatusCondition(&routeConfig.Status.Conditions, metav1.Condition{
		Type:    "reconciled",
		Status:  metav1.ConditionTrue,
		Reason:  "done",
		Message: "reconciliation done",
	})

	log.Info(fmt.Sprintf("<- POST STATUS RouteConfig ResourceVersion=%s", routeConfig.ObjectMeta.ResourceVersion))
	err = r.Status().Update(ctx, routeConfig)
	if err != nil {
		log.Error(err, "-> POST STATUS RouteConfig")
	} else {
		log.Info(fmt.Sprintf("-> POST STATUS RouteConfig ResourceVersion=%s", routeConfig.ObjectMeta.ResourceVersion))

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouterConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupv1.RouterConfig{}).
		Complete(r)
}
