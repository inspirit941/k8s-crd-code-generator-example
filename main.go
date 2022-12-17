package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1"
	klient "github.com/inspirit941/kluster/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

func main() {
	k := v1alpha1.Kluster{}
	fmt.Println(k)

	// Kluster 사용
	var kubeconfig *string
	if home, err := os.UserHomeDir(); err != nil {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("building config from flags, %s", err.Error())
	}

	// generated code 사용
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klientset %s", err.Error())
	}
	fmt.Println(klientset)
	// 이런 식으로 쓸 수 있다.
	klusters, err := klientset.Inspirit941V1alpha1().Klusters("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf(err.Error())
	}
	fmt.Println(len(klusters.Items))
}
