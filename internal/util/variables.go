/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package util

import (
	"context"
	"encoding/base64"

	"github.com/getupio-undistro/undistro/client/config"
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
	VariablesClient config.VariablesClient
	ClientSet       client.Client
	NamespacedName  types.NamespacedName
	EnvVars         []corev1.EnvVar
}

func SetVariablesFromEnvVar(ctx context.Context, input VariablesInput) error {
	for _, envVar := range input.EnvVars {
		if envVar.Value != "" {
			input.VariablesClient.Set(envVar.Name, envVar.Value)
			continue
		}
		if envVar.ValueFrom != nil {
			if envVar.ValueFrom.FieldRef != nil || envVar.ValueFrom.ResourceFieldRef != nil {
				return errors.New("fieldRef and resourceFieldRef are not supported as provider variables")
			}
			if envVar.ValueFrom.SecretKeyRef != nil {
				input.VariablesClient.Set(envVar.Name, valueFromSecret(ctx, input.ClientSet, envVar.ValueFrom.SecretKeyRef, input.NamespacedName))
				continue
			}
			if envVar.ValueFrom.ConfigMapKeyRef != nil {
				input.VariablesClient.Set(envVar.Name, valueFromConfigMap(ctx, input.ClientSet, envVar.ValueFrom.ConfigMapKeyRef, input.NamespacedName))
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
