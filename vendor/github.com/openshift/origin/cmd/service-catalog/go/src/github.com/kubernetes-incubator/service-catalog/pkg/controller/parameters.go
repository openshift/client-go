/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// buildParameters generates the parameters JSON structure to be passed
// to the broker
func buildParameters(kubeClient kubernetes.Interface, namespace string, parametersFrom []v1alpha1.ParametersFromSource, parameters *runtime.RawExtension) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if parametersFrom != nil {
		for _, p := range parametersFrom {
			fps, err := fetchParametersFromSource(kubeClient, namespace, &p)
			if err != nil {
				return nil, err
			}
			for k, v := range fps {
				if _, ok := params[k]; ok {
					return nil, fmt.Errorf("conflict: duplicate entry for parameter %q", k)
				}
				params[k] = v
			}
		}
	}
	if parameters != nil {
		pp, err := UnmarshalRawParameters(parameters.Raw)
		if err != nil {
			return nil, err
		}
		for k, v := range pp {
			if _, ok := params[k]; ok {
				return nil, fmt.Errorf("conflict: duplicate entry for parameter %q", k)
			}
			params[k] = v
		}
	}
	return params, nil
}

// fetchParametersFromSource fetches data from a specified external source and
// represents it in the parameters map format
func fetchParametersFromSource(kubeClient kubernetes.Interface, namespace string, parametersFrom *v1alpha1.ParametersFromSource) (map[string]interface{}, error) {
	var params map[string]interface{}
	if parametersFrom.SecretKeyRef != nil {
		data, err := fetchSecretKeyValue(kubeClient, namespace, parametersFrom.SecretKeyRef)
		if err != nil {
			return nil, err
		}
		p, err := unmarshalJSON(data)
		if err != nil {
			return nil, err
		}
		params = p

	}
	return params, nil
}

// UnmarshalRawParameters produces a map structure from a given raw YAML/JSON input
func UnmarshalRawParameters(in []byte) (map[string]interface{}, error) {
	parameters := make(map[string]interface{})
	if len(in) > 0 {
		if err := yaml.Unmarshal(in, &parameters); err != nil {
			return parameters, err
		}
	}
	return parameters, nil
}

// unmarshalJSON produces a map structure from a given raw JSON input
func unmarshalJSON(in []byte) (map[string]interface{}, error) {
	parameters := make(map[string]interface{})
	if err := json.Unmarshal(in, &parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters as JSON object: %v", err)
	}
	return parameters, nil
}

// fetchSecretKeyValue requests and returns the contents of the given secret key
func fetchSecretKeyValue(kubeClient kubernetes.Interface, namespace string, secretKeyRef *v1alpha1.SecretKeyReference) ([]byte, error) {
	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(secretKeyRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret.Data[secretKeyRef.Key], nil
}
