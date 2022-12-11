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

