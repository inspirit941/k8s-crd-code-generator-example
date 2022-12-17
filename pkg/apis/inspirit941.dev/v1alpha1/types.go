package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Kluster struct {
	// k8s object / resource는 세 개의 main field가 필요함.
	metav1.TypeMeta   `json:",inline"`            // type meta: which particular type of resources it is. client-go의 metav1을 쓸 수 있다.
	metav1.ObjectMeta `json:"metadata,omitempty"` // object meta: details of object (name / namespace 등)

	Spec KlusterSpec `json:"spec,omitempty"` // space spec : configuration or details about that object
}

type KlusterSpec struct {
	// specify the field needed when the operator runs
	// input으로 필요한 값.
	Name    string `json:"name,omitempty"`
	Region  string `json:"region,omitempty"`
	version string `json:"version,omitempty"`

	NodePools []NodePool `json:"nodePools,omitempty"` // digitalOcean api를 보면 size, name, count 값이 required인 array임.
}

type NodePool struct {
	Size  string `json:"size,omitempty"`
	Name  string `json:"name,omitempty"`
	Count int    `json:"count,omitempty"`
}

// kubectl pod list처럼 kluster list 명령어로 리스트 호출하기 위해 정의.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KlusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"` // objectmeta가 아니라 listmeta로 설정해야 에러가 안 남.

	Items []Kluster `json:"items,omitempty"`
}
