package controller

import (
	"context"
	"github.com/inspirit941/kluster/pkg/apis/inspirit941.dev/v1alpha1"
	klientset "github.com/inspirit941/kluster/pkg/client/clientset/versioned"
	klusterscheme "github.com/inspirit941/kluster/pkg/client/clientset/versioned/scheme"
	informer "github.com/inspirit941/kluster/pkg/client/informers/externalversions/inspirit941.dev/v1alpha1"
	klister "github.com/inspirit941/kluster/pkg/client/listers/inspirit941.dev/v1alpha1"
	"github.com/inspirit941/kluster/pkg/digitalocean"
	"github.com/kanisterio/kanister/pkg/poll"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

// required Field to run a Custom controller.
type Controller struct {
	// clientset for custom resource 'kluster'
	klient klientset.Interface
	// k8s apiserver 접근할 client도 추가한다.
	client kubernetes.Interface

	// informer maintains local cache. cache가 initialized/synced 되었는지 확인.
	// kluster has synced or not.
	klusterSynced cache.InformerSynced
	// lister
	kLister klister.KlusterLister
	// queue. object의 Create / delete 작업을 순차적으로 수행하기.
	wq workqueue.RateLimitingInterface
	// Event Recorder
	recorder record.EventRecorder
}

func NewController(client kubernetes.Interface, klient klientset.Interface, klusterInformer informer.KlusterInformer) *Controller {
	// 이벤트를 생성할 때 "어떤 컴포넌트가 이벤트를 생성했는지"를 추가해줘야 함.
	// -> Controller / Operator의 type을 code-generator가 Event code를 생성할 때 같이 넣어주는 것.
	// Custom Resource를 code generate할 때 만들어진 scheme 패키지를 아래와 같이 사용한다.
	runtime.Must(klusterscheme.AddToScheme(scheme.Scheme)) // operator type을 k8s standard scheme에 추가함.
	log.Println("creating event broadcaster...")
	eventBroadCaster := record.NewBroadcaster()
	eventBroadCaster.StartStructuredLogging(0) // 로깅관련 세팅 추가
	eventBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: client.CoreV1().Events(""),
	}) // event interface 추가
	recorder := eventBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "Kluster"})

	c := &Controller{
		client:        client,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
		recorder:      recorder,
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
	// 클러스터에서 이미 삭제된 경우
	if apierrors.IsNotFound(err) {
		log.Printf("handle delete event for kluster '%s'", name)
		// todo: digitalocean api에서 클러스터 삭제 요청.
		return true
	}
	if err != nil {
		log.Printf("error during get Kluster resouces from lister: %s", err.Error())
		return false
	}

	log.Printf("Kluster spec from Resource : %+v", kluster.Spec)

	// digital ocean api 호출
	clusterID, err := digitalocean.Create(c.client, kluster.Spec)
	if err != nil {
		// do something (retry or return)
	}
	// 성공적으로 클러스터가 생성될 경우 이벤트 생성
	c.recorder.Event(kluster, corev1.EventTypeNormal, "ClusterCreation", "Digital Ocean Creation API was called to create the cluster.")

	log.Printf("cluster created; cluster id: %s", clusterID)
	err = c.updateStatus(clusterID, "creating", kluster)
	if err != nil {
		log.Printf("error: %s,  during update status of the cluster '%s'\n", err.Error(), kluster.Name)
	}

	// query DigitalOcean API to make sure cluster's status is Running.
	// kubectl create -f 명령어로 실행해도 알아서 백그라운드로 동작함.
	err = c.waitForCluster(kluster.Spec, clusterID)
	if err != nil {
		log.Printf("error %s, waiting for cluster to be running", err.Error())
	}

	// status 변경. production의 경우 retry 로직이 추가되어야 함.
	err = c.updateStatus(clusterID, "running", kluster)
	if err != nil {
		log.Printf("error %s, updating cluster status after waiting for cluster")
	}
	c.recorder.Event(kluster, corev1.EventTypeNormal, "ClusterCreationCompleted", "Digital Ocean Creation API was completed.")
	return true
}

func (c *Controller) waitForCluster(spec v1alpha1.KlusterSpec, clusterId string) error {
	// context에 timeout 붙여서 '특정 시간을 초과하면 더 이상 poll하지 않도록' 설정
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// poll.Wait() 파라미터로 들어간 func가 true를 리턴할 때까지 주기적으로 실행함.
	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := digitalocean.ClusterState(c.client, spec, clusterId)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}
		return false, nil
	})
}

// subresource인 Status를 업데이트하는 로직
func (c *Controller) updateStatus(id, progress string, kluster *v1alpha1.Kluster) error {
	// update를 실행할 때, kluster struct가 이미 modified된 상태면 에러가 발생함
	// i.e. error Operation cannot be fulfilled on kluster.inspirit941.dev "<cr name>" : the object has been modified; please apply your changes to the latest version and try again..
	// 따라서 latest kluster struct를 받을 수 있도록 수정. (get the latest version of kluster)
	k, err := c.klient.Inspirit941V1alpha1().Klusters(kluster.Namespace).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	k.Status.KlusterID = id
	k.Status.Progress = progress
	// subresource 정의한 다음 code-generate하면 새로 생성되는 메소드.
	_, err = c.klient.Inspirit941V1alpha1().Klusters(kluster.Namespace).UpdateStatus(context.Background(), kluster, metav1.UpdateOptions{})
	return err
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
