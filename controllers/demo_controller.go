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
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demoappv1 "demo-operator/api/v1"
)

// DemoReconciler reconciles a Demo object
type DemoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// kubebuilder annotation 다시 추가
//+kubebuilder:rbac:groups=demoapp.my.domain,resources=demoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demoapp.my.domain,resources=demoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demoapp.my.domain,resources=demoes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;delete

// SetupWithManager sets up the controller with the Manager.
func (r *DemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demoappv1.Demo{}).  // For에 감시할 CR을 설정합니다.
		Owns(&corev1.Service{}). // Owns는 서브로 감시할 대상입니다. (서브 감시 대상이 삭제되면 reconcile 되도록)
		Owns(&appsv1.Deployment{}).
		Complete(r)

	// 여기서 서브로 감시할 대상에 추가된 service와 deploy는
	// 추후 임의로 삭제하면 다시 복구됩니다.
}

func (r *DemoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx) // logger 정의
	cr := &demoappv1.Demo{}        // CR 객체 정의
	svc := &corev1.Service{}       // svc 객체 정의
	dply := &appsv1.Deployment{}   // deploy 객체 정의

	// 클러스터에서 해당 CR이 있는지 확인합니다.
	err := r.Client.Get(ctx, req.NamespacedName, cr)

	if err != nil { // Get CR에 에러가 있는 경우

		if errors.IsNotFound(err) { // 변경사항인 cr이 k8s에 존재하지 않는 경우
			logger.Info("CR is Deleted")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to get CR")
		return ctrl.Result{}, err // 기타 에러 처리
	}

	size := cr.Spec.Size // deploy.Spec.Replicas 값으로 들어감.

	// 클러스터에서 cr용 service가 있는지 확인합니다.
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}, svc)

	if err != nil { // Get service에 에러가 있는 경우

		if errors.IsNotFound(err) { // 해당 service가 클러스터에 없는 경우 생성합니다.
			newSvc := r.createService(cr)
			err = r.Create(ctx, newSvc)

			if err != nil { // 생성중 에러가 발생하면 이벤트 큐에 다시 넣음.
				logger.Info("failed to create Service", "svc.namespace", newSvc.Namespace, "svc.name", newSvc.Name)
				return ctrl.Result{}, err
				// 이벤트 큐 - return에서 err 포함하거나, ctrl.Result{}에 Requeue 옵션 포함
			}

			logger.Info("Service Created", "svc.namespace", newSvc.Namespace, "svc.name", newSvc.Name)
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to Get Service")
		return ctrl.Result{}, err // 기타 에러 처리
	}

	// 클러스터에서 cr용 deploy가 있는지 확인합니다.
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}, dply)

	if err != nil { // Get deploy에 에러가 있는 경우

		if errors.IsNotFound(err) { // 해당 deploy가 클러스터에 없는 경우 생성합니다.
			newDply := r.createDeployment(cr)
			err = r.Create(ctx, newDply)

			if err != nil {
				logger.Info("failed to create Service", "deploy.namespace", newDply.Namespace, "deploy.name", newDply.Name)
				return ctrl.Result{}, err
			}

			logger.Info("Deployment Created", "deploy.namespace", newDply.Namespace, "deploy.name", newDply.Name)
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to Get Deployment")
		return ctrl.Result{}, err
	}

	// deploy 정의할 때 사용한 replicas 값과 cr.Spec.Size 값이 다른 경우
	if *dply.Spec.Replicas != size {

		dply.Spec.Replicas = &size // replicas를 변경된 값으로 맞춰줌.

		// deploy에 replicas 값 변경을 반영합니다.
		logger.Info("changed replicas Size", "deploy.namespace", dply.Namespace, "deploy.Name", dply.Name, "Size", size)
		err = r.Client.Update(ctx, dply)

		if err != nil {
			logger.Error(err, "Error in Updating ReplicaSet", "deploy.Namespace", dply.Namespace,
				"deploy.Name", dply.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// status.Nodes
	podList := &corev1.PodList{}
	label := getLabelForCR(cr.Name)
	listOps := []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
		client.MatchingLabels(label),
	}

	err = r.Client.List(ctx, podList, listOps...)
	if err != nil {
		logger.Error(err, "Failed to list Pods.", "demo.Namespace", cr.Namespace, "demo.Name", cr.Name)
		return ctrl.Result{}, err
	}

	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, cr.Status.Nodes) {
		logger.Info("Update pod list in demo", "podNames", podNames)
		cr.Status.Nodes = podNames
		err := r.Client.Status().Update(ctx, cr)
		if err != nil {
			logger.Error(err, "Failed to update Demo Status.")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
