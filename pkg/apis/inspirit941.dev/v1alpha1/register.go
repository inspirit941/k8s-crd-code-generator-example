package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CRD가 어떤 group / version인지는 schema.GroupVersion 변수로 정의한다.
var SchemeGroupVersion = schema.GroupVersion{
	Group:   "inspirit941.dev",
	Version: "v1alpha1",
}

var (
	SchemeBuilder runtime.SchemeBuilder
)

// 패키지가 로드될 때 Kluster라는 struct를 등록.
func init() {
	SchemeBuilder.Register(addKnownTypes) // register 안에 파라미터로 들어갈 function이 type을 scheme에 등록하는 역할을 수행함.
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &Kluster{}, &KlusterList{})

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// go struct를 k8s resource 형태로 호출하려면 몇 가지 함수를 implement해야 한다.
// DeepCopy, set / get GroupVersion 필요. deepcopy의 경우 code-generator에서 사용한다고 함
// set / get groupversion은 metav1.TypeMeta와 metav1.ObjectMeta를 변수로 할당한 시점에서 이미 구현됨.
//// CR의 get, list, delete, update 등을 수행하기 위한 clientset
//// informers: 해당 리소스의 상태를 k8s operator가 확인할 수 있도록 만드는 컴포넌트
//// lister: get resources from informers cache
