package cluster

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type EventListener interface {
	Listen(context.Context, *rest.Config, ...string) (watch.Interface, error)
}

type MultipleEventsWatcher struct {
	Watchers []watch.Interface
	ch       chan watch.Event
}

func (w *MultipleEventsWatcher) Stop() {
	for i := range w.Watchers {
		w.Watchers[i].Stop()
	}
	close(w.ch)
}

func (w *MultipleEventsWatcher) ResultChan() <-chan watch.Event {
	ctx := context.Background()
	if w.ch == nil {
		w.ch = make(chan watch.Event)
	}
	for index := range w.Watchers {
		go func(ctx context.Context, wi watch.Interface, eventCh chan watch.Event) {
			for e := range wi.ResultChan() {
				eventCh <- e
			}
		}(ctx, w.Watchers[index], w.ch)
	}
	return w.ch
}

type eventListener struct {
	getter corev1client.EventsGetter
}

func NewEventListener() *eventListener {
	return &eventListener{}
}

func (e *eventListener) Listen(ctx context.Context, cfg *rest.Config, objNames ...string) (watch.Interface, error) {
	cfg.Timeout = 0
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	namesSet := sets.NewString(objNames...)
	e.getter = c.CoreV1()
	ws := make([]watch.Interface, namesSet.Len())
	for i, n := range namesSet.List() {
		w, err := e.getter.Events("").Watch(ctx, metav1.ListOptions{
			Watch:         true,
			FieldSelector: fmt.Sprintf("involvedObject.name=%s", n),
		})
		if err != nil {
			return nil, err
		}
		ws[i] = w
	}
	return &MultipleEventsWatcher{
		Watchers: ws,
	}, nil
}
