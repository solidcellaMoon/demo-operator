package controllers

import (
	demoappv1 "demo-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Label을 메소드로 모듈화하여 사용
func getLabelForCR(crName string) map[string]string {
	return map[string]string{"app": crName}
}

// Service를 생성하고, 컨트롤러에 등록해 cr이 삭제된 경우 함께 삭제되도록 합니다.
func (r *DemoReconciler) createService(d *demoappv1.Demo) *corev1.Service {

	label := getLabelForCR(d.Name)

	// service yaml을 하드코딩으로 정의
	newSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: label,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: 80},
				},
			},
		},
	} // svc 정의 끝

	// cr이 삭제됐을때 svc가 남아있는걸 막기 위해 ref에 추가
	ctrl.SetControllerReference(d, newSvc, r.Scheme)
	return newSvc
}

// Deployment를 생성하고 컨트롤러에 등록해 cr이 삭제되면 함께 삭제되도록 합니다.
func (r *DemoReconciler) createDeployment(d *demoappv1.Demo) *appsv1.Deployment {

	label := getLabelForCR(d.Name)
	size := d.Spec.Size // CR.Spec.Size 정의 내용을 사용

	// Deployment yaml을 하드코딩으로 정의
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
						Name:  "nginx",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 80,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					}},
				},
			},
		},
	} // deploy 정의 끝

	// cr이 삭제됐을때 deploy가 남아있는걸 막기 위해 ref에 추가
	ctrl.SetControllerReference(d, newDply, r.Scheme)
	return newDply
}

// pod Name List
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, p := range pods {
		podNames = append(podNames, p.Name)
	}
	return podNames
}
