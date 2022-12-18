## Generating ClientSet / Informers / Lister and CRD for Custom Resources | Writing k8s Operator - part 1

- https://youtu.be/89PdRvRUcPU


k8s에서 동작할 Operator
- Custom Resource를 생성하면 DigitalOcean에 kubernetes cluster를 Provision.
- CR 삭제 시 DigitalOcean의 kubernetes cluster 삭제.
- 예시 Repo: https://github.com/viveksinghggits/kluster


### DigitalOcean 에서 사용할 API 확인 (CRD가 필요한 이유)

https://docs.digitalocean.com/reference/api/api-reference/#tag/Kubernetes

- name
- region
- version
- node_pools

즉, 이 네 개의 필드는 required 필드로, 값이 반드시 있어야 함. 
- 이 값을 입력으로 받기 위해서는 k8s의 Custom Resource를 정의해야 한다.
- Custom Resource가 생성 또는 삭제되면, k8s Operator가 특정 작업을 수행하는 것.


네 개의 필드를 Custom Resource에 지정하려면 client-go에서 특정 type / struct를 정의해야 함. 
- 정확히는, k8s object는 특정 group과 version 필드를 필요로 한다.
  - deployment의 경우 group: apps, version: v1

Kluster Object를 정의한다고 하면
- group: inspirit941.dev
- version: v1alpha1

두 가지를 명확히 정의하고, group에 맞게 패키지를 정의한다. 패키지에는 CR에서 받게 될 필드를 의미하는 types.go 파일을 생성한다.
- `pkg/apis/inspirit941.dev/v1alpha1`


types를 정의했지만, 이렇게 정의한 go struct만 가지고서는 이게 k8s object인지 알 수 없음. 따라서 k8s object의 scheme로 인식시킨 뒤 register해줘야 함.
- k8s의 code-generator 프로젝트의 register.go 참고 (https://github.com/kubernetes/code-generator/blob/master/examples/crd/apis/example/v1/register.go)

Code Generater가 만들어주는 컴포넌트들은 아래와 같다.
- DeepCopy 메소드
- clientset
- informers
- lister


- tag: the way to control the behavior of your code generator
  - code generator 동작을 위해 파라미터나 flag를 입력하지만, tag 값도 입력으로 쓸 수 있다.
  - global tags: doc.go 파일에서 정의. specify global tags for your api
  - local tags

doc.go의 맨 앞줄에 주석으로 표시한 값의 의미
```go
// +k8s:deepcopy-gen=package -> deepcopy 메소드는 패키지에 정의된 모든 타입을 대상으로 만들어야 한다.
// +k8s:defaulter-gen=TypeMeta -> 
// +groupName=inspirit941.dev
```

local tag는 types.go에서 특정 struct에 주석을 붙이는 식으로 동작. 주석 의미는 아래와 같다
- Kluster struct의 clientset을 생성한다

```go
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Kluster struct {
	metav1.TypeMeta   
	metav1.ObjectMeta 
	Spec KlusterSpec
}
```

code Generator(https://github.com/kubernetes/code-generator)를 로컬에 받고, execDir 환경변수를 code-generator 다운받은 path로 설정한다.
- execDir=/your/path/code-generator
- `"${execDir}"/generate-groups.sh all github.com/inspirit941/kluster/pkg/client github.com/inspirit941/kluster/pkg/apis inspirit941.dev:v1alpha1 --go-header-file "${execDir}/hack/boilerplate.go.txt"`
  - all 파라미터: 필요로 하는 4개 컴포넌트 (deepcopy object, clientset, informers, lister) 전부 생성한다.
  - github.com/... : 컴포넌트들이 어느 모듈에 들어갈 것인지. 모듈명이 들어가야 한다. (go mod로 생성한 거)
  - .../pkg/apis : 참조해야 할 types들은 모듈 내 어느 path에 있는지
  - `inspirit941.dev:v1alpha1` : group과 version 명시
  - boilerplate.go.txt : 생성된 코드에 라이센스 정보를 명시하기 위한 로직.

위 명령어를 실행하면 deepcopy func, clientset, lister, informer 파일이 생성된다. github.com/inspirit941/kluster/pkg에 생성된다.
- GOPATH/github.com/inspirit941/kluster/pkg... 에 생성되었음. github 아카이브를 위해 이 repo에 생성된 값을 복사해넣음.

kubernetes-native resource를 client-go에서 호출해 쓸 때는 kubernetes.NewForConfig()을 보통 쓰지만, <Br>
CRD를 import해서 쓸 때는 보통 clientset 모듈을 호출해서 쓴다.

cf. https://github.com/viveksinghggits/kluster/issues/1. 위 sh로 코드 생성해서 go mod tidy를 하면 
- `k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset: module k8s.io/kubernetes@latest found (v1.26.0), but does not contain package k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset`
- 에러가 뜬다. informer 값을 로컬에서 전부 지우면 go mod tidy가 동작함.

k8s가 registered된 type을 토대로 뭔가 응답을 보내려면 CRD를 정의해줘야 한다.
- crd 없이 생성하면 type을 인식하기만 할 뿐 특정 작업을 수행하지는 못하는 듯. (the server could not find the requested resources 에러)
- 따라서 CRD를 controller-gen 으로 만들어주고, types에 정의된 struct에 json 옵션을 추가해준다.

-> controller-gen을 사용해서 CRD를 생성한다. ![referenec](https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/reference/controller-gen.md)
- controller-gen을 Cli로 설치하려면
  - https://github.com/kubernetes-sigs/controller-tools를 clone한다
  - `go build ./cmd/controller-gen`
  - 만들어진 binary를 /usr/local/bin 디렉토리에 추가.
- `controller-gen paths=github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1 crd:trivialVersions=true crd:crdVersions=v1 output:crd:artifacts:config=manifests`
- -> 강의에서는 이렇게 명령어 입력. 내 로컬에서는 `controller-gen paths=github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1 +crd:crdVersions=v1 +output:artifacts:config=manifests`
  - paths: api 정보가 있는 Package
  - crdVersions: v1권장. crdVersion v1beta1은 k8s 1.22 버전부터 deprecated되었다. (https://github.com/k8s-operatorhub/community-operators/discussions/468)
  - config=manifest -> manifest 형태로 generate

터미널을 입력하면 루트 디렉토리에 ./manifests/inspirit941.dev_klusters.yaml 파일 하나가 생성됨.
- 이렇게 생성된 yaml 파일을 클러스터에 배포한 뒤 main.go를 실행하면, (the server could not find the requested resources 에러)가 발생하지 않는다.

이제 아래의 yaml을 CR로 배포할 수 있다. 배포한 리소스는 `kubectl get kluster` 로도 확인 가능.

```yaml
apiVersion: inspirit941.dev/v1alpha1
kind: Kluster
metadata:
  name: kluster-0
spec:
  name: kluster-0
  region: "ap-south-1"
  version: "1234"
```

요약
- CR / CRD는 k8s cluster supported resource.
- 해당 리소스를 k8s에 Register하기 -> register.go의 AddKnownTypes 함수의 파라미터로 custom하게 정의한 go struct 추가.
- CR의 behavior 정의하기
  - metav1.TypeMeta, metav1.ObjectMeta 등 몇몇 메소드는 타입으로 정의하면 기본 제공
  - deepcopy는 직접 구현해야 함 / clientset, informers, lister 구현 필요 -> code-generator 사용
  - CRD는 controller-gen으로 구현 가능.

---

## Add / del handler func & token field in kluster CRD

https://www.youtube.com/watch?v=MOutOgdXfnA

controller 패키지 생성, 코드 작성
- CR에 생성 / 삭제 이벤트 발생 시 EventHandler, workqueue에서 이벤트 받아 처리하는 goroutine 실행함수.
- lister에 Resource() 메소드 구현이 빠져 있어서 추가.


## Calling DigitalOcean API on CustomResource's add event

유의: production 환경에서는 이런 방식으로 코드를 짜는 것보다는 interface로 구현에 필요한 메소드만 정의하고, cloud provider가 interface를 구현하는 식으로 접근하는 게 맞음

```go
type ... interface {
	CreateCluster(...)
	DeleteCluster(...)
	GetKubeConfig(...)
}

// cloud provider struct가 Interface implement할 수 있도록.
type AKS struct {}
type EKS struct {}
type DigitalOcean struct {
	
}

func (d *DigitalOcean) CreateCluster(...) {
	
}
```