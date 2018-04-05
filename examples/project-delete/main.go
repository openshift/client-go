package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	projectClient "github.com/openshift/client-go/project/clientset/versioned"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getOpenShiftProjectClient(kubeconfig string) (*projectClient.Clientset, error) {
	// kubeconfig passed as parameter is coming from cmd line

	// if kubeconfig not provided on cmd line then read it from the env
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	// if it is still empty then take the value from the file which acts as default
	if kubeconfig == "" {
		kubeconfig = os.ExpandEnv("$HOME/.kube/config")
	}

	var (
		err    error
		config *rest.Config
	)

	// if kubeconfig is empty then in cluster config generation function will called
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return projectClient.NewForConfig(config)
}

func errorsToError(errs []error) error {
	var errStr string
	for _, err := range errs {
		errStr += fmt.Sprintf("%s\n", err)
	}
	return fmt.Errorf("%s", errStr)
}

func deleteUserProjects(cli *projectClient.Clientset, username string) error {
	projects, err := cli.ProjectV1().Projects().List(meta_v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "could not list projects")
	}
	var errs []error
	for _, project := range projects.Items {
		// check if the annotation "openshift.io/requester", exists on the project
		// if it does then see if the value is username we are looking for, if so
		// we are good to go and delete the project
		if val, ok := project.Annotations["openshift.io/requester"]; ok && val == username {
			// delete this project
			log.Printf("deleting project %s", project.Name)
			err = cli.ProjectV1().Projects().Delete(project.Name, &meta_v1.DeleteOptions{})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errorsToError(errs)
	}
	return nil
}

func main() {
	var username, kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig file")
	flag.StringVar(&username, "username", "", "Provide username whose projects to be deleted")
	flag.Parse()

	if username == "" {
		log.Fatal("Please provide a username whose projects to be deleted")
	}

	projectCli, err := getOpenShiftProjectClient(kubeconfig)
	if err != nil {
		log.Fatalf("could not get the project client: %v", err)
	}

	if err := deleteUserProjects(projectCli, username); err != nil {
		log.Fatalf("error deleting projects: %v", err)
	}
}
