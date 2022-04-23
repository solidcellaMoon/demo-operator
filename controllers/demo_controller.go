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

	"k8s.io/apimachinery/pkg/api/errors" // 추가 패키지
	"k8s.io/apimachinery/pkg/runtime"
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

//+kubebuilder:rbac:groups=demoapp.my.domain,resources=demoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demoapp.my.domain,resources=demoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demoapp.my.domain,resources=demoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Demo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile

var logger = log.Log

func (r *DemoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	// logger 사용법
	logger.WithName("Resource Name")
	logger.WithValues("Resource NameSpace: ", req.Namespace)
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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demoappv1.Demo{}).
		Complete(r)
}
