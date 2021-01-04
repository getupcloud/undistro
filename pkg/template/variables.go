/*
Copyright 2020 The UnDistro authors

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
package template

import (
	"context"
	"encoding/base64"
	"sync"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log = ctrl.Log.WithName("variables")
)

type VariablesInput struct {
	ClientSet      client.Client
	NamespacedName types.NamespacedName
	EnvVars        []corev1.EnvVar
	Variables      map[string]interface{}
	mtx            *sync.RWMutex
}

func SetVariablesFromEnvVar(ctx context.Context, input VariablesInput) error {
	if input.mtx == nil {
		input.mtx = &sync.RWMutex{}
	}
	for _, envVar := range input.EnvVars {
		input.mtx.Lock()
		input.Variables[envVar.Name] = envVar.Value
		input.mtx.Unlock()
		if envVar.ValueFrom != nil {
			if envVar.ValueFrom.FieldRef != nil || envVar.ValueFrom.ResourceFieldRef != nil {
				return errors.New("fieldRef and resourceFieldRef are not supported as provider variables")
			}
			if envVar.ValueFrom.SecretKeyRef != nil {
				v := valueFromSecret(ctx, input.ClientSet, envVar.ValueFrom.SecretKeyRef, input.NamespacedName)
				input.mtx.Lock()
				input.Variables[envVar.Name] = v
				input.mtx.Unlock()
				continue
			}
			if envVar.ValueFrom.ConfigMapKeyRef != nil {
				v := valueFromConfigMap(ctx, input.ClientSet, envVar.ValueFrom.ConfigMapKeyRef, input.NamespacedName)
				input.mtx.Lock()
				input.Variables[envVar.Name] = v
				input.mtx.Unlock()
			}
		}
	}
	return nil
}

func valueFromSecret(ctx context.Context, client client.Client, selector *corev1.SecretKeySelector, n types.NamespacedName) string {
	n.Name = selector.Name
	var secret corev1.Secret
	if err := client.Get(ctx, n, &secret); err != nil {
		log.Error(err, "couldn't get secret", "name", n)
		return ""
	}
	b64, ok := secret.Data[selector.Key]
	if !ok {
		err := errors.Errorf("key %s not found", selector.Key)
		log.Error(err, "couldn't get secret key", "name", n)
		return ""
	}
	s, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		return string(b64)
	}
	return string(s)
}

func valueFromConfigMap(ctx context.Context, client client.Client, selector *corev1.ConfigMapKeySelector, n types.NamespacedName) string {
	n.Name = selector.Name
	var cfgMap corev1.ConfigMap
	if err := client.Get(ctx, n, &cfgMap); err != nil {
		log.Error(err, "couldn't get configMap", "name", n)
		return ""
	}
	value, ok := cfgMap.Data[selector.Key]
	if !ok {
		err := errors.Errorf("key %s not found", selector.Key)
		log.Error(err, "couldn't get configMap key", "name", n)
		return ""
	}
	return value
}
