package controller

import (
	klientset "github.com/inspirit941/kluster/pkg/client/clientset/versioned"
	informer "github.com/inspirit941/kluster/pkg/client/informers/externalversions/inspirit941.dev/v1alpha1"
	klister "github.com/inspirit941/kluster/pkg/client/listers/v1alpha1/internalversion"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

// required Field to run a Custom controller.
type Controller struct {
	// clientset for custom resource 'kluster'
	klient klientset.Interface

	// informer maintains local cache. cache가 initialized/synced 되었는지 확인.
	// kluster has synced or not.
	klusterSynced cache.InformerSynced
	// lister
	kLister klister.KlusterLister
	// queue. object의 Create / delete 작업을 순차적으로 수행하기.
	wq workqueue.RateLimitingInterface
}

func NewController(klient klientset.Interface, klusterInformer informer.KlusterInformer) *Controller {
	c := &Controller{
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
	}

	// register functions.
	// resource에 특정 이벤트가 들어올 때 실행될 함수를 정의하는 영역.
	// 리소스 이벤트가 들어오면 workqueue에 먼저 들어가고,
	//  goroutine에서 queue 값을 받아서 실행하는 방식.
	// 따라서 eventHandler는 queue에 이벤트 값을 넣는 역할.
	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd, // 리소스가 생성될 때
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

// workqueue로부터 값을 consume받아 처리하는 goroutine
func (c *Controller) Run(ch chan struct{}) error {
	// check if local cache has been initialized at least once.
	if ok := cache.WaitForCacheSync(ch, c.klusterSynced); !ok {
		// 캐시가 싱크되지 않음
		log.Println("cache was not synced")
	}
	// goroutine consumes from workqueue
	go wait.Until(c.worker, time.Second, ch) // 채널이 closed되기 전까지 run 'f' every period.
	<-ch
	return nil
}

func (c *Controller) worker() {
	// kluster resource 관련 로직. continuously하게 수행.
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	// get resource from cache
	item, shutDown := c.wq.Get()
	if shutDown {
		// logs as well
		return false
	}

	// 함수가 동작 끝나면 workqueue에서 제거.
	defer c.wq.Forget(item)

	// get namespace / name from cache
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error %s calling Namespace key func on cache for item", err.Error())
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("error during splitting key into namespace / name: %s", err.Error())
		return false
	}

	// ns, name 확인했으니 lister로 exact Object 조회
	kluster, err := c.kLister.Klusters(ns).Get(name)
	if err != nil {
		log.Printf("error during get Kluster resouces from lister: %s", err.Error())
		return false
	}
	log.Printf("Kluster spec from Resource : %+v", kluster.Spec)

	return true
}

// 리소스 생성 이벤트가 들어올 때.
func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)
}
