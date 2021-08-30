/*
Copyright 2021.

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
	"strings"

	v1 "github.com/gepaplexx/oauth-group-sync-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=user.openshift.io,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=user.openshift.io,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=user.openshift.io,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// fetch user instance
	user := &v1.User{}
	err := r.Client.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Reconciling User", "user", user)

	// abort if user has no identities
	if len(user.Identities) == 0 {
		log.Info("User has no identities, skipping")
		return ctrl.Result{}, nil
	}

	// filter primary identity
	identity := user.Identities[0]
	identityGroup := strings.Split(identity, ":")[0]

	// get group instance
	group := &v1.Group{}
	groupSelector := types.NamespacedName{Name: identityGroup}
	err = r.Client.Get(ctx, groupSelector, group)
	if err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		group = nil
	}

	// check if group exists, else create one
	if group == nil {
		log.Info("Group not found! creating group!")
		group = &v1.Group{
			ObjectMeta: metav1.ObjectMeta{
				Name: identityGroup,
			},
		}
		err := r.Create(ctx, group)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// skip if user is already in group
	if isUserInGroup(user, group) {
		log.Info("User is already in group, skipping!")
		return ctrl.Result{}, nil
	}

	// add user to group
	log.Info("Adding user to group")
	group.Users = append(group.Users, user.ObjectMeta.Name)
	err = r.Client.Update(ctx, group)
	if err != nil {
		return ctrl.Result{}, err
	}

	// add label
	if user.ObjectMeta.Annotations == nil {
		user.ObjectMeta.Annotations = make(map[string]string)
	}
	user.ObjectMeta.Annotations["gepaplexx.com/groupsynced"] = "true"
	err = r.Client.Update(ctx, user)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.User{}).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		// For().
		Complete(r)
}

// checks if user is in group
func isUserInGroup(user *v1.User, group *v1.Group) bool {
	for _, member := range group.Users {
		if member == user.ObjectMeta.Name {
			return true
		}
	}
	return false
}
