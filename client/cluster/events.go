package cluster

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
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
	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u := unstructured.Unstructured{
		Object: o,
	}
	e.getter = c.CoreV1()
	return e.getter.Events("").Watch(ctx, metav1.ListOptions{
		Watch:         true,
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", u.GetName()),
	})
}
