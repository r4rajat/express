package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"os"
	"time"
)

type controller struct {
	clientSet           kubernetes.Interface
	deploymentLister    appslisters.DeploymentLister
	deploymentCacheSync cache.InformerSynced
	queue               workqueue.RateLimitingInterface
}

func getNewController(clientSet kubernetes.Interface, deploymentInformer appsinformer.DeploymentInformer) *controller {
	queue := os.Getenv("EXPRESS_QUEUE")
	if queue == "" {
		queue = "EXPRESS"
	}
	newController := &controller{
		clientSet:           clientSet,
		deploymentLister:    deploymentInformer.Lister(),
		deploymentCacheSync: deploymentInformer.Informer().HasSynced,
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), queue),
	}
	deploymentInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: newController.addHandler,
		},
	)
	return newController
}

func (c *controller) addHandler(obj interface{}) {
	log.Println("Add Handler Called")
	c.queue.Add(obj)
}

func (c *controller) run(ch <-chan struct{}) {
	if !cache.WaitForCacheSync(ch, c.deploymentCacheSync) {
		log.Println("Waiting for cache to be Synced.")
	}
	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *controller) worker() {
	for c.processItems() {

	}
}

func (c *controller) processItems() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("Error getting key from cache.\nreason --> %s", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("Error splitting key into namespace and name.\nReason --> %s", err.Error())
		return false
	}

	err = c.createService(ns, name)
	if err != nil {
		//retry
		log.Printf("Error Syncing Deployments.\nReason --> %s", err.Error())
		return false
	}
	return true
}

func (c *controller) createService(ns string, name string) error {
	//create service
	deployment, err := c.deploymentLister.Deployments(ns).Get(name)
	if err != nil {
		log.Printf("Error Getting Deployment Name from Lister.\nReason --> %s", err.Error())
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: deployment.Spec.Template.Labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
				{
					Name: "https",
					Port: 443,
				},
			},
		},
	}
	ctx := context.Background()
	s, err := c.clientSet.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error Creating Service.\nReason --> %s", err.Error())
	}
	log.Println("Service Created... ", s)
	//create ingress
	return createIngressForService(ctx, c.clientSet, svc)
}

func createIngressForService(ctx context.Context, client kubernetes.Interface, service corev1.Service) error {
	pathType := "Prefix"
	ingress := netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				{
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     fmt.Sprintf("/%s", service.Name),
									PathType: (*netv1.PathType)(&pathType),
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: service.Name,
											Port: netv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := client.NetworkingV1().Ingresses(service.Namespace).Create(ctx, &ingress, metav1.CreateOptions{})
	return err
}
