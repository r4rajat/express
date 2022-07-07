package main

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"
)

func main() {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Printf("Unable to get Home Directory of Current User.\nReason --> %s", err.Error())
	}
	kubeconfigPath := homeDir + "/.kube/config"
	log.Printf("Setting default kubeconfig location to --> %s", kubeconfigPath)
	kubeconfig := flag.String("kubeconfig", kubeconfigPath, "Location to your kubeconfig file")
	if (*kubeconfig != "") && (*kubeconfig != kubeconfigPath) {
		log.Printf("Recieved new kubeconfig location --> %s", *kubeconfig)
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Not able to create kubeconfig object from default location.\nReason --> %s", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("Not able to create kubeconfig object from inside pod.\nReason --> %s", err.Error())
		}
	}
	log.Println("Created config object with provided kubeconfig")

	// Creating Clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error occurred while creating Client Set with provided config.\nReason --> %s", err.Error())
	}
	ch := make(chan struct{})
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	controller := getNewController(clientSet, informer.Apps().V1().Deployments())
	informer.Start(ch)
	controller.run(ch)
}
