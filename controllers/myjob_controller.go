/*
Copyright 2020.

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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	myjobv1beta1 "github.com/sky-big/myjob-operator/api/v1beta1"
)

// MyJobReconciler reconciles a MyJob object
type MyJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=myjob.github.com,resources=myjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=myjob.github.com,resources=myjobs/status,verbs=get;update;patch

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

func (r *MyJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("myjob", req.NamespacedName)

	// your logic here
	j := &myjobv1beta1.MyJob{}
	if err := r.Get(ctx, req.NamespacedName, j); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 设置MyJob的Status的默认值
	if j.StatusSetDefault() {
		if err := r.Status().Update(ctx, j); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// pod不存在则创建pod，如果存在检查pod的状态
	p := &corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, p)
	if err == nil {
		// MyJob的状态还是Pending，但是对应的Pod已经创建，则将MyJob的状态置为Running
		if !isPodCompleted(p) && myjobv1beta1.MyJobRunning != j.Status.Phase {
			j.Status.Phase = myjobv1beta1.MyJobRunning
			if err := r.Status().Update(ctx, j); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("myjob phase changed", "Phase", myjobv1beta1.MyJobRunning)
		}

		// MyJob对应的Pod已经执行完毕，则将MyJob的状态置为Completed
		if isPodCompleted(p) && myjobv1beta1.MyJobRunning == j.Status.Phase {
			j.Status.Phase = myjobv1beta1.MyJobCompleted
			if err := r.Status().Update(ctx, j); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("myjob phase changed", "Phase", myjobv1beta1.MyJobCompleted)
		}

	} else if err != nil && errors.IsNotFound(err) {
		// 创建MyJob对应的Pod
		pod := makePodByMyJob(j)
		if err := controllerutil.SetControllerReference(j, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, pod); err != nil && !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
		}
		logger.Info("myjob create pod success")

	} else {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MyJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)

	// 监视拥有者是MyJob类型的Pod，同时将Pod的拥有者MyJob扔进处理队列中，对MyJob进行调和
	c.Watches(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &myjobv1beta1.MyJob{},
	})

	return c.For(&myjobv1beta1.MyJob{}).
		Complete(r)
}

func isPodCompleted(pod *corev1.Pod) bool {
	if corev1.PodSucceeded == pod.Status.Phase ||
		corev1.PodFailed == pod.Status.Phase ||
		pod.DeletionTimestamp != nil {
		return true
	}
	return false
}

func makePodByMyJob(j *myjobv1beta1.MyJob) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.Name,
			Namespace: j.Namespace,
		},
		Spec: *j.Spec.Template.Spec.DeepCopy(),
	}
}
