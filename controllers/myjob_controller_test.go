package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	myjobv1beta1 "github.com/sky-big/myjob-operator/api/v1beta1"
)

var _ = Describe("MyJob controller", func() {
	const (
		MyjobName      = "test-myjob"
		MyjobNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating MyJob", func() {
		It("Should be success", func() {
			By("By creating a new MyJob")
			ctx := context.Background()

			// 0. 创建 myjob
			cronJob := &myjobv1beta1.MyJob{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "myjob.github.com/v1beta1",
					Kind:       "MyJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      MyjobName,
					Namespace: MyjobNamespace,
				},
				Spec: myjobv1beta1.MyJobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								v1.Container{
									Name:    "pi",
									Image:   "perl",
									Command: []string{"perl", "-Mbignum=bpi", "-wle", "print bpi(2000)"},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, cronJob)).Should(Succeed())

			myjobKey := types.NamespacedName{Name: MyjobName, Namespace: MyjobNamespace}
			createdMyjob := &myjobv1beta1.MyJob{}

			// 1. 验证 myjob 创建成功
			Eventually(func() bool {
				err := k8sClient.Get(ctx, myjobKey, createdMyjob)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdMyjob.Name).Should(Equal(MyjobName))

			// 2. 验证 myjob 创建 pod
			myPodKey := types.NamespacedName{Name: MyjobName, Namespace: MyjobNamespace}
			myPod := &v1.Pod{}
			Consistently(func() (string, error) {
				err := k8sClient.Get(ctx, myPodKey, myPod)
				if err != nil {
					return "", err
				}
				return myPod.Name, nil
			}, duration, interval).Should(Equal(MyjobName))

			// 3. 验证 myjob 状态变为 Running
			runningMyjob := &myjobv1beta1.MyJob{}
			Consistently(func() bool {
				err := k8sClient.Get(ctx, myjobKey, runningMyjob)
				if err != nil {
					return false
				}
				return runningMyjob.Status.Phase == myjobv1beta1.MyJobRunning
			}, duration, interval).Should(BeTrue())

			// 4. Mock Pod 工作完成
			mockPod := &v1.Pod{}
			Consistently(func() bool {
				err := k8sClient.Get(ctx, myPodKey, mockPod)
				if err != nil {
					return false
				}

				copy := mockPod.DeepCopy()
				copy.Status.Phase = v1.PodSucceeded
				err = k8sClient.Status().Update(context.TODO(), copy)
				if err != nil {
					return false
				}
				return true
			}, duration, interval).Should(BeTrue())

			// 5. 验证 myjob 状态变为 Completed
			completedMyjob := &myjobv1beta1.MyJob{}
			Consistently(func() bool {
				err := k8sClient.Get(ctx, myjobKey, completedMyjob)
				if err != nil {
					return false
				}
				return completedMyjob.Status.Phase == myjobv1beta1.MyJobCompleted
			}, duration, interval).Should(BeTrue())
		})
	})
})
