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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	groupv1 "github.com/gprossliner/operator-demo/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const conditionRouterConfigurationTriggered = "RouterConfigTriggered"
const finalizerRoute = "group.text.org/finalizer"

//+kubebuilder:rbac:groups=group.test.org,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=group.test.org,resources=routes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=group.test.org,resources=routes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Route object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *RouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Route")

	route := &groupv1.Route{}
	log.Info(fmt.Sprintf("<- GET Route"))
	err := r.Get(ctx, req.NamespacedName, route)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Route resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Route")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("-> GET Route ResourceVersion=%s", route.ObjectMeta.ResourceVersion))

	// Check if the route instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isMarkedToBeDeleted := route.GetDeletionTimestamp() != nil
	if isMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(route, finalizerRoute) {

			// Run finalization logic
			routerConfig, err := findRouterConfig(ctx, r.Client, *route)
			if err != nil {
				return ctrl.Result{}, err
			}

			removeRoute(&routerConfig.Status, route)
			log.Info(fmt.Sprintf("<- POST STATUS RouterConfig ResourceVersion=%s", routerConfig.ObjectMeta.ResourceVersion))
			err = r.Status().Update(ctx, routerConfig)
			if err != nil {
				log.Error(err, "-> POST RouteConfig")
			} else {
				log.Info(fmt.Sprintf("-> POST STATUS Route ResourceVersion=%s", routerConfig.ObjectMeta.ResourceVersion))
			}

			// Remove finalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(route, finalizerRoute)
			err = r.Update(ctx, route)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(route, finalizerRoute) {
		controllerutil.AddFinalizer(route, finalizerRoute)
		err = r.Update(ctx, route)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// check if we already triggered the RouterConfig based on this generation
	cond := meta.FindStatusCondition(route.Status.Conditions, conditionRouterConfigurationTriggered)
	if cond != nil && cond.ObservedGeneration == route.Generation {
		log.Info(fmt.Sprintf("This route has already triggered reconfiguration for genertion %d", route.Generation))
		return ctrl.Result{}, nil
	}

	// find the corresponding RouterConfig
	routerConfig, err := findRouterConfig(ctx, r.Client, *route)
	log.Info(fmt.Sprintf("-> LIST ITEM RouterConfig ResourceVersion=%s", routerConfig.ObjectMeta.ResourceVersion))
	if routerConfig.Spec.RouterConfigName == route.Spec.RouterConfigName {

		// got the correct routerConfig, let's update the routes
		setRoute(&routerConfig.Status, route)

		log.Info(fmt.Sprintf("<- POST STATUS RouterConfig ResourceVersion=%s", routerConfig.ObjectMeta.ResourceVersion))
		err = r.Status().Update(ctx, routerConfig)
		if err != nil {
			log.Error(err, "-> POST RouteConfig")
		} else {
			log.Info(fmt.Sprintf("-> POST STATUS Route ResourceVersion=%s", routerConfig.ObjectMeta.ResourceVersion))
		}

		// set another condition to show we had configured the RouterConfig
		meta.SetStatusCondition(&route.Status.Conditions, metav1.Condition{
			Type:               conditionRouterConfigurationTriggered,
			Status:             metav1.ConditionTrue,
			Reason:             "done",
			Message:            "RouterConfig reconfiguration triggered",
			ObservedGeneration: route.Generation,
		})

		log.Info(fmt.Sprintf("<- POST STATUS Route ResourceVersion=%s", route.ObjectMeta.ResourceVersion))
		err = r.Status().Update(ctx, route)
		if err != nil {
			log.Error(err, "-> POST STATUS RouteConfig")
		} else {
			log.Info(fmt.Sprintf("-> POST STATUS Route ResourceVersion=%s", route.ObjectMeta.ResourceVersion))
		}

	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *RouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupv1.Route{}).
		Complete(r)
}

// findRouterConfig returns the RouterConfig for a given route.
// An error is only returned if the API call returns an error.
// If no matching RouterConfig can be found, nil is returned
func findRouterConfig(ctx context.Context, client client.Client, route groupv1.Route) (*groupv1.RouterConfig, error) {

	log := ctrllog.FromContext(ctx)

	// find the corresponding RouterConfig
	routerConfigs := &groupv1.RouterConfigList{}
	log.Info(fmt.Sprintf("<- LIST RouterConfig"))
	err := client.List(ctx, routerConfigs)
	if err != nil {
		return nil, err
	}

	for _, routerConfig := range routerConfigs.Items {
		log.Info(fmt.Sprintf("-> LIST ITEM RouterConfig ResourceVersion=%s", routerConfig.ObjectMeta.ResourceVersion))
		if routerConfig.Spec.RouterConfigName == route.Spec.RouterConfigName {
			return &routerConfig, nil
		}
	}

	return nil, nil
}

func setRoute(status *groupv1.RouterConfigStatus, route *groupv1.Route) {

	myref := groupv1.RouteReference{
		Namespace:          route.Namespace,
		Name:               route.Name,
		ObservedGeneration: route.Generation,
	}

	for i, ref := range status.Routes {
		if ref.Namespace == myref.Namespace && ref.Name == myref.Name {
			status.Routes[i].ObservedGeneration = myref.ObservedGeneration
			return
		}
	}

	status.Routes = append(status.Routes, myref)
	return
}

func removeRoute(status *groupv1.RouterConfigStatus, route *groupv1.Route) {
	var newRoutes []groupv1.RouteReference
	for _, ref := range status.Routes {
		if !(ref.Namespace == route.Namespace && ref.Name == route.Name) {
			newRoutes = append(newRoutes, ref)
		}
	}

	status.Routes = newRoutes
}
