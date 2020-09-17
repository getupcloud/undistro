package kubeconfig

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Purpose is the name to append to the secret generated for a cluster.
type Purpose string

const (
	// KubeconfigDataName is the key used to store a Kubeconfig in the secret's data field.
	KubeconfigDataName = "value"
	// Kubeconfig is the secret name suffix storing the Cluster Kubeconfig.
	Kubeconfig = Purpose("kubeconfig")
	// UserKubeconfig is the secret name suffix storing the Cluster Kubeconfig for user usage.
	UserKubeconfig = Purpose("user-kubeconfig")
)

func FromSecret(ctx context.Context, c client.Reader, cluster client.ObjectKey) ([]byte, error) {
	out, err := Get(ctx, c, cluster, UserKubeconfig)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		out, err = Get(ctx, c, cluster, Kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return toKubeconfigBytes(out)
}

// Get retrieves the specified Secret (if any) from the given
// cluster name and namespace.
func Get(ctx context.Context, c client.Reader, cluster client.ObjectKey, purpose Purpose) (*corev1.Secret, error) {
	return GetFromNamespacedName(ctx, c, cluster, purpose)
}

// GetFromNamespacedName retrieves the specified Secret (if any) from the given
// cluster name and namespace.
func GetFromNamespacedName(ctx context.Context, c client.Reader, clusterName client.ObjectKey, purpose Purpose) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: clusterName.Namespace,
		Name:      Name(clusterName.Name, purpose),
	}

	if err := c.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}

	return secret, nil
}

// Name returns the name of the secret for a cluster.
func Name(cluster string, suffix Purpose) string {
	return fmt.Sprintf("%s-%s", cluster, suffix)
}

func toKubeconfigBytes(out *corev1.Secret) ([]byte, error) {
	data, ok := out.Data[KubeconfigDataName]
	if !ok {
		return nil, errors.Errorf("missing key %q in secret data", KubeconfigDataName)
	}
	return data, nil
}
