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

### SubResources / additional printer columes

- spec 필드에서 cluster 생성에 쓰이는 configuration 값을 입력받은 상태.
- 하지만 spec에 '사용자가 입력한 값 말고', controller에서 spec에 특정 필드를 추가하거나 수정하는 식의 동작이 필요할 수 있음. 
  - 예컨대 create로 클러스터를 생성하고 id값을 받음. 다른 로직에서 clusterID가 쓰이는 경우가 많으므로, 이 값을 Custom Resource instance에 저장해두고 다른 로직에서 활용할 수 있으면 좋다.
- Custom Resource가 생성되었을 때, '실제 DigitalOcean 클러스터 생성은 완료되었는지', 'digitalOcean 클러스터의 현재 상태' 와 같은 정보를 `k get kluster` 에서 활용할 수 있다면 좋다.

status 필드를 추가할 예정.
- spec은 사용자에 의해서만 수정, status는 controller에 의해서만 수정되도록 구조 설정
  - status라는 Subresources를 정의하고, Controller는 subresource인 status 수정만 가능하도록 role 생성.

subresource 예시 
- Pod의 Subresource는 logs.
- k8s Resource 예시
  - `apis/apps/v1/namespace/<ns>/deployments`
  - `v1/namespaces/<ns>/pods/<podname>/logs`


```yaml
apiVersion:
kind:
metadata:
spec:
  ...
  ...
  ...
status: # 정보 추가
  clusterId
  progress
  kubeconfig

---
role-test:
  resources: Kluster/status
```

subresource인 Status 필드가 추가되었으니, UpdateStatus() 와 같은 subresource 관련 메소드가 필요. code generator 실행해서 업데이트한다.

printer column : kubebuilder 공식문서에서 'additional printer columns' 확인하면 알 수 있음.
- types.go 참고

### Event Recorder for Kluster and routines to handle objects from queue

subresource인 progress를 creating으로 변경하는 것까지는 진행한 상태. 실제로 digitalocean api에서 클러스터가 생성되면 progress를 다른 걸로 변경해줘야 함.
- https://github.com/kanisterio/kanister 의 poll 패키지 -> Wait() 메소드를 사용할 예정.
- go get github.com/kanisterio/kanister/poll 로 설치. -> graymeta/stow 라는 디펜던시 버전에러가 있으므로, replace 명령어로 버전을 맞춰준다.

Event Recorder
- 예컨대 `kubectl create deployment nginx --image nginx` 로 deployment를 생성한 뒤 describe로 상태를 확인해보면 Events 필드가 있다.
  - 'deployment에서 pod scale을 얼마로 조정했는지' 같은 이벤트가 발생
  - 누가 이 이벤트를 발생시켰는지 (From -> i.e. deployment-controller) 
  - 이벤트가 발생하고 얼마나 시간이 지났는지 (Age)
  - 이벤트 타입은 무엇인지 (Normal / Error 등등)
  - 왜 발생했는지 (Reason -> i.e. Scheduled, Pulling...)


예시의 경우 kluster가 생성될 때, status 필드가 업데이트될 때 event를 생성할 수 있음.

#### concurrency issue with operator

지금 코드는 controller.Run() 메소드에서 하나의 goroutine이 돌고 있음. 요청이 들어오는 queue를 처리하는 daemon process가 하나인 셈
- digitalocean에서 create api 호출 후 status를 업데이트하는 로직이 하나의 daemon process에서 처리되므로, 
- 동시에 여러 Custom resource를 생성하면 digital ocean api 호출까지 시간이 오래 걸림

-> run 내부에서 process를 여러 개 만들어두거나, run 메소드의 파라미터로 goroutine 개수를 지정하는 식으로 리팩토링할 수 있음.


### Packaging (Containerize & RBAC) k8s operator

- Dockerfile 생성. `docker build -t <dockerhub_repo>/<name>:tag` 로 빌드
- 생성된 docker image를 push한 뒤, `kubectl create deployment kluster --image <dockerhub-image> --dry-run=client -oyaml > deploy.yaml`
- deploy.yaml 파일 배포.
  - digitalocean 토큰값을 클러스터에 secret으로 배포되어 있어야 함 `kubectl create secret generic dosecret --from-literal token="token_value"`
  - CRD yaml도 배포되어 있어야 함.

![스크린샷 2022-12-24 오후 4 28 01](https://user-images.githubusercontent.com/26548454/209425907-a9333d15-a87d-4941-8f98-906d3e68b49f.png)
- inClusterConfig() 메소드를 사용했을 때 발생할 수 있는 에러.
  - 별다른 설정을 하지 않았으므로 default ServiceAccount를 사용했는데, 이 account는 리소스에 접근할 권한이 없음.
  - 따라서 권한이 있는 serviceAccount를 생성하고, deployment가 해당 serviceaccount를 사용하도록 수정.

cf. serviceAccount가 controller에서 사용하는 k8s 리소스에는 전부 접근권한이 있어야 함.
- secret / CRD (kluster) / events / subresource (status)
- 이 중 secret은 clusterRole이 아니라, 특정 namespace에 배포된 secret에만 접근할 수 있도록 rolebinding 형태로 추가해야 함.

RBAC -> 생성한 뒤 yaml 파일을 일부 수정한 내용이 있음
- `kubectl create serviceaccount kluster-sa --dry-run=client -oyaml > sa.yaml`
- `kubectl create clusterrole kluster-clusterrole --resource Kluster --verb list,watch,get --dry-run=client -oyaml` -> 이후에 yaml 추가로 수정.
- `kubectl create clusterrolebinding kluster-crb --serviceaccount default:kluster-sa --clusterrole kluster-clusterrole --dry-run=client -oyaml > manifests/clusterrolebinding.yaml`
- `kubectl create role kluster-role --resource secret --verb get --dry-run=client -oyaml`

