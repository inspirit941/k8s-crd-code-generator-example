package main

import (
	"flag"
	"fmt"
	"github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1"
	klient "github.com/inspirit941/kluster/pkg/client/clientset/versioned"
	"github.com/inspirit941/kluster/pkg/client/informers/externalversions"
	"github.com/inspirit941/kluster/pkg/controller"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"time"
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

	//generated code 사용
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klientset %s", err.Error())
	}
	//fmt.Println(klientset)
	//// 이런 식으로 쓸 수 있다.
	//klusters, err := klientset.Inspirit941V1alpha1().Klusters("").List(context.Background(), metav1.ListOptions{})
	//if err != nil {
	//	log.Printf(err.Error())
	//}
	//fmt.Println(len(klusters.Items))

	// informer를 호출하려면 informerFactory를 사용해야 함.
	informerFactory := externalversions.NewSharedInformerFactory(klientset, 20*time.Minute) // resync 시간은 20분으로 정의.
	// informer를 동작시키려면 chan이 필요.
	ch := make(chan struct{})
	c := controller.NewController(klientset, informerFactory.Inspirit941().V1alpha1().Klusters())

	informerFactory.Start(ch)
	if err := c.Run(ch); err != nil {
		log.Printf("error running controller %s\n", err.Error())
	}
}
