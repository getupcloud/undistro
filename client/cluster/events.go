package cluster

import (
	"context"

	"github.com/getupio-undistro/undistro/internal/scheme"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/reference"
)

type EventListener interface {
	Listen(context.Context, *rest.Config, runtime.Object) (watch.Interface, error)
}

type eventListener struct {
	getter corev1client.EventsGetter
}

func NewEventListener() *eventListener {
	return &eventListener{}
}

func (e *eventListener) Listen(ctx context.Context, cfg *rest.Config, obj runtime.Object) (watch.Interface, error) {
	cfg.Timeout = 0
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	e.getter = c.CoreV1()
	ref, err := reference.GetReference(scheme.Scheme, obj)
	if err != nil {
		return nil, errors.Errorf("can't get object reference: %v", err)
	}
	uid := string(ref.UID)
	sec := e.getter.Events(undistroNamespace).GetFieldSelector(&ref.Name, &ref.Namespace, &ref.Kind, &uid)
	return e.getter.Events(undistroNamespace).Watch(ctx, metav1.ListOptions{
		Watch:         true,
		FieldSelector: sec.String(),
	})
}
