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

	appsv1 "k8s.io/api/apps/v1"          // k8s apps
	corev1 "k8s.io/api/core/v1"          // k8s core
	"k8s.io/apimachinery/pkg/api/errors" // 추가 패키지
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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

			logger.Info("Service Created", "deploy.namespace", newDply.Namespace, "deploy.name", newDply.Name)

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

// Service를 생성하고, 컨트롤러에 등록해 cr이 삭제된 경우 함께 삭제되도록 합니다.
// yaml을 하드코딩하였음.
func (r *DemoReconciler) createService(d *demoappv1.Demo) *corev1.Service {

	// Label은 여러 곳에서 사용하는 pod의 정보가 담긴 데이터입니다.
	// 메소드로 모듈화시켜 정적자원처럼 사용하려고 합니다.
	label := GetLabelForCR(d.Name)

	// service struct는 metadata, spec 등을 구현할 수 있도록 정의
	newSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP, // cluster IP
			Selector: label,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
			},
		},
	} // svc 정의 끝

	// cr이 삭제됐을때 svc가 남아있는걸 막기 위해 ref에 추가
	ctrl.SetControllerReference(d, newSvc, r.Scheme)
	return newSvc
}

// !! CR.Spec.Size 정의 내용을 사용하는 곳입니다 !!
// Deployment를 생성하고 컨트롤러에 등록해 cr이 삭제되면 함께 삭제되도록 합니다.
// yaml을 golang으로 하드코딩한 형태입니다...
func (r *DemoReconciler) createDeployment(d *demoappv1.Demo) *appsv1.Deployment {

	label := GetLabelForCR(d.Name)
	size := d.Spec.Size

	// Deployment 정의
	newDply := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: label,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "nginx:latest",
						Name:  "nginx-app",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8080,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					}},
				},
			},
		},
	} // 정의 끝

	ctrl.SetControllerReference(d, newDply, r.Scheme)
	return newDply
}

// pod의 Label은 pod를 인식하는 데이터입니다.
// 이것을 메소드로 모듈화하여 정적자원처럼 사용합니다.
func GetLabelForCR(name string) map[string]string {
	return map[string]string{"app": "nginx-app"}
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
	// 추후 사용자가 임의로 삭제하면 다시 복구됩니다~!
	// 현재 cr이 삭제됐을 때 svc와 dep가 함께 삭제되지는 않습니다.
	// 해당 로직이 필요한 경우, 컨트롤러 ref에 추가합니다.
}
