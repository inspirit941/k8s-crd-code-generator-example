package digitalocean

import (
	"context"
	"github.com/digitalocean/godo"
	"github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

// https://docs.digitalocean.com/reference/api/api-reference/#tag/Kubernetes
func Create(c kubernetes.Interface, spec v1alpha1.KlusterSpec) (string, error) {
	token, err := getToken(c, spec.TokenSecret)
	if err != nil {
		return "", err
	}
	// digitalOcean은 토큰을 토대로 K8S secret 정보 가져와서 수행하는 방식
	client := godo.NewFromToken(token)
	request := &godo.KubernetesClusterCreateRequest{
		Name:        spec.Name,
		VersionSlug: spec.Version,
		RegionSlug:  spec.Region,
		// todo: nodepool input 값의 validation, 여러 개의 nodepool 리스트 요청일 경우 리스트로 request 생성하기.
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			{
				Size:  spec.NodePools[0].Size,
				Name:  spec.NodePools[0].Name,
				Count: spec.NodePools[0].Count,
			},
		},
	}
	cluster, _, err := client.Kubernetes.Create(context.Background(), request)
	if err != nil {
		return "", err
	}
	return cluster.ID, nil
}

// call k8s api server to get secret from secretName
func getToken(client kubernetes.Interface, secretName string) (string, error) {
	namespace := strings.Split(secretName, "/")[0]
	name := strings.Split(secretName, "/")[1]
	s, err := client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(s.Data["token"]), nil
}

// digitalOcean에서 생성한 클러스터의 상태 체크용 함수
// https://docs.digitalocean.com/reference/api/api-reference/#operation/kubernetes_get_cluster
func ClusterState(c kubernetes.Interface, spec v1alpha1.KlusterSpec, id string) (string, error) {
	token, err := getToken(c, spec.TokenSecret)
	if err != nil {
		return "", err
	}
	// digitalOcean은 토큰을 토대로 K8S secret 정보 가져와서 수행하는 방식
	client := godo.NewFromToken(token)
	cluster, _, err := client.Kubernetes.Get(context.Background(), id)
	return string(cluster.Status.State), err
}
