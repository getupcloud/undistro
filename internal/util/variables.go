/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package util

import (
	"github.com/getupcloud/undistro/client/config"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetVariablesFromEnvVar(client client.Client, v config.VariablesClient, envVars []corev1.EnvVar) error {
	for _, envVar := range envVars {
		if envVar.Value != "" {
			v.Set(envVar.Name, envVar.Value)
			continue
		}
		if envVar.ValueFrom != nil {
			if envVar.ValueFrom.FieldRef != nil || envVar.ValueFrom.ResourceFieldRef != nil {
				return errors.New("fieldRef and resourceFieldRef are unsupported as provider variables")
			}
			if envVar.ValueFrom.SecretKeyRef != nil {
				v.Set(envVar.Name, valueFromSecret(client, envVar.ValueFrom.SecretKeyRef))
				continue
			}
			if envVar.ValueFrom.ConfigMapKeyRef != nil {
				v.Set(envVar.Name, valueFromConfigMap(client, envVar.ValueFrom.ConfigMapKeyRef))
			}
		}
	}
	return nil
}

func valueFromSecret(client client.Client, selector *corev1.SecretKeySelector) string {
	return ""
}

func valueFromConfigMap(client client.Client, selector *corev1.ConfigMapKeySelector) string {
	return ""
}
