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

package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MyJobPending = "pending"

	MyJobRunning = "running"

	MyJobCompleted = "completed"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MyJobSpec defines the desired state of MyJob
type MyJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Template v1.PodTemplateSpec `json:"template" protobuf:"bytes,6,opt,name=template"`
}

// MyJobStatus defines the observed state of MyJob
type MyJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Phase string `json:"phase,omitempty"`
}

func (j *MyJobStatus) SetDefault(job *MyJob) bool {
	changed := false

	if job.Status.Phase == "" {
		job.Status.Phase = MyJobPending
		changed = true
	}

	return changed
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MyJob is the Schema for the myjobs API
type MyJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyJobSpec   `json:"spec,omitempty"`
	Status MyJobStatus `json:"status,omitempty"`
}

func (j *MyJob) StatusSetDefault() bool {
	return j.Status.SetDefault(j)
}

// +kubebuilder:object:root=true

// MyJobList contains a list of MyJob
type MyJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyJob{}, &MyJobList{})
}
