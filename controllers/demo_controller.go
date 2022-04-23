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
	"time"

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

func (r *DemoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	// logger 예시
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Resource Changed")

	// CR로 정의한 객체를 가져오기 위한 struct의 ref를 받아옵니다.
	cr := &demoappv1.Demo{}

	// 데이터를 서버에서 받아와 cr에 넣어줍니다.
	err := r.Client.Get(ctx, req.NamespacedName, cr)

	// cr에 변경이 존재하거나 err가 발생한 경우 진행되는 로직입니다.
	// 변경사항이 존재한다는 것으로 생각해야 합니다.
	if err != nil {
		// 변경사항인 cr이 k8s에 존재하는지 확인합니다.
		if errors.IsNotFound(err) {
			logger.Info("Deleted - Demo CR")
			return ctrl.Result{}, nil
		}

		// GET 함수 처리
		logger.Error(err, "GET CR Error occurred")
		return ctrl.Result{}, err
	}

	// 1. service 객체를 만들어줍니다.
	svc := &corev1.Service{}

	// 서버에서 cr로 만들어진 service를 받아옵니다.
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}, svc)

	// service를 받아왔더니 변경사항이 존재합니다.
	if err != nil {
		// 서비스가 found 되지 않은 경우 생성합니다.

		if errors.IsNotFound(err) {
			// 서비스 생성!!
			newSvc := r.createService(cr)
			err = r.Create(ctx, newSvc)

			if err != nil {
				logger.Info("failed to create Service", "svc.namespace", newSvc.Namespace, "svc.name", newSvc.Name)
				return ctrl.Result{}, err
			}

			logger.Info("Service Created", "svc.namespace", newSvc.Namespace, "svc.name", newSvc.Name)

			// Requeue를 설정해주면 이벤트큐에 다시 올라가 다시 로직이 진행됩니다...
			return ctrl.Result{RequeueAfter: time.Second * 2}, nil
		}

		logger.Error(err, "Failed to Get Service")
		return ctrl.Result{}, err
	}

	// 2. Deployment 생성
	dply := &appsv1.Deployment{}

	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}, dply)

	// deploy 받아왔더니 변경사항이 존재합니다...
	if err != nil {
		// deploy가 found 되지 않은 경우 생성합니다.
		if errors.IsNotFound(err) {
			// deploy 생성!!
			newDply := r.createDeployment(cr)
			err = r.Create(ctx, newDply)

			if err != nil {
				logger.Info("failed to create Service", "deploy.namespace", newDply.Namespace, "deploy.name", newDply.Name)
				return ctrl.Result{}, err
			}

			logger.Info("Deployment Created", "deploy.namespace", newDply.Namespace, "deploy.name", newDply.Name)

			// Requeue를 설정해주면 이벤트큐에 다시 올라가 다시 로직이 진행됩니다...
			return ctrl.Result{RequeueAfter: time.Second * 2}, nil
		}

		logger.Error(err, "Failed to Get Deployment")
		return ctrl.Result{}, err
	}

	// deploy 생성할 때 cr.Spec.Size로 replicaSet 생성을 해주었음.
	// cr.Spec.Size가 변경되어 apply된 경우 해당 로직으로 제어함.
	size := cr.Spec.Size

	// deploy 정의할 때 사용한 replicas와 cr.Spec.Size 값이 다른 경우
	if *dply.Spec.Replicas != size {

		// replicas를 변경된 값으로 맞춰줌.
		dply.Spec.Replicas = &size

		// custom controller를 만들더라도 기존 k8s controll loop는 정상적으로 돌아갑니다...
		// replicaset만 변경해서 pod 수를 제어합니다.
		logger.Info("changed replicas Size", "ReplicaSet.namespace", cr.Namespace, "Size", size)

		err = r.Client.Update(ctx, dply)

		if err != nil {
			logger.Error(err, "Error in Updating ReplicaSet", "Deploy.Namespace", dply.Namespace,
				"Deploy.Name", dply.Name)
			return ctrl.Result{}, err
		}

		// Requeue를 설정해주면 이벤트큐에 다시 올라가 다시 로직이 진행됩니다...
		return ctrl.Result{RequeueAfter: time.Second * 2}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demoappv1.Demo{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)

	// For에 감시할 CR을 설정합니다.
	// Owns는 서브로 감시할 대상입니다. (서브 감시 대상이 삭제되면 reconcile 되도록)
	// 서브로 감시할 대상에 추가된 service와 deploy는
	// 추후 사용자가 임의로 삭제하면 다시 복구됩니다.
}
