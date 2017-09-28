package main

import (
	"flag"
	"fmt"
	"github.com/openshift/client-go/project/internalclientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	if *kubeconfig == "" {
		panic("-kubeconfig not specified")
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err)
		}
		clientset, err := internalclientset.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		projects, err := clientset.Projects().List(metav1.ListOptions{})
		if err != nil {
			fmt.Println("List projects failed: ", err)
		}
		for c, _ := range projects.Items {
			fmt.Println(projects.Items[c].GetName())
		}
	}
}
