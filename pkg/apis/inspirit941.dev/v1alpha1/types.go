package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Kluster struct {
	// k8s object / resource는 세 개의 main field가 필요함.
	metav1.TypeMeta   // type meta: which particular type of resources it is. client-go의 metav1을 쓸 수 있다.
	metav1.ObjectMeta // object meta: details of object (name / namespace 등)

	// space spec : configuration or details about that object
	Spec KlusterSpec
}

type KlusterSpec struct {
	// specify the field needed when the operator runs
	// input으로 필요한 값.
	Name    string
	Region  string
	version string

	NodePools []NodePool // digitalOcean api를 보면 size, name, count 값이 required인 array임.
}

type NodePool struct {
	Size  string
	Name  string
	Count int
}
