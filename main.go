package main

import (
	"flag"
	"fmt"
	"github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1"
	klient "github.com/inspirit941/kluster/pkg/client/clientset/versioned"
	"github.com/inspirit941/kluster/pkg/client/informers/externalversions"
	"github.com/inspirit941/kluster/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
		// containerize할 경우 이 코드는 동작하지 않음. 따라서 kubeconfig가 생성되지 않음.
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("building config from flags, trying to build inclusterconfig. %s", err.Error())
		// kubeconfig 파일이 없는 경우
		config, err = rest.InClusterConfig() // kubernetes pod에 mount된 secret Account를 사용
		if err != nil {
			log.Printf("error %s building inclusterconfig: %s", err.Error())
		}
	}

	//generated code 사용
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klientset %s", err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("getting k8s client %s", err.Error())
	}

	// informer를 호출하려면 informerFactory를 사용해야 함.
	informerFactory := externalversions.NewSharedInformerFactory(klientset, 20*time.Minute) // resync 시간은 20분으로 정의.
	// informer를 동작시키려면 chan이 필요.
	ch := make(chan struct{})
	c := controller.NewController(client, klientset, informerFactory.Inspirit941().V1alpha1().Klusters())

	informerFactory.Start(ch)
	if err := c.Run(ch); err != nil {
		log.Printf("error running controller %s\n", err.Error())
	}
}
