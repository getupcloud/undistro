/*
Copyright 2018 The Kubernetes Authors.

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

package record

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
)

var (
	initOnce        sync.Once
	defaultRecorder record.EventRecorder
)

func init() {
	defaultRecorder = new(record.FakeRecorder)
}

// InitFromRecorder initializes the global default recorder. It can only be called once.
// Subsequent calls are considered noops.
func InitFromRecorder(recorder record.EventRecorder) {
	initOnce.Do(func() {
		defaultRecorder = recorder
	})
}

type WatcherResource struct {
	Cluster *undistrov1.Cluster
	Obj     runtime.Object
}

func EventWatcher(ctx context.Context, cfg *rest.Config) (chan WatcherResource, error) {
	c := make(chan WatcherResource, 1)
	cerr := make(chan error, 1)
	uc, err := client.New("")
	if err != nil {
		return nil, err
	}
	listener, err := uc.GetEventListener(client.Kubeconfig{
		RestConfig: cfg,
	})
	if err != nil {
		return nil, err
	}
	go func(ctx context.Context) {
		for o := range c {
			w, err := listener.Listen(ctx, cfg, o.Obj)
			if err != nil {
				cerr <- err
				return
			}
			go watchEvents(w, o.Cluster)
		}
	}(ctx)
	return c, nil
}

func watchEvents(w watch.Interface, o runtime.Object) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	select {
	case e := <-w.ResultChan():
		ev, ok := e.Object.(*corev1.Event)
		if !ok {
			return
		}
		switch ev.Type {
		case corev1.EventTypeNormal:
			Event(o, ev.Reason, ev.Message)
		case corev1.EventTypeWarning:
			Warn(o, ev.Reason, ev.Message)
		}
	case <-ch:
		w.Stop()
	}
}

// Event constructs an event from the given information and puts it in the queue for sending.
func Event(object runtime.Object, reason, message string) {
	defaultRecorder.Event(object, corev1.EventTypeNormal, strings.Title(reason), message)
}

// Eventf is just like Event, but with Sprintf for the message field.
func Eventf(object runtime.Object, reason, message string, args ...interface{}) {
	defaultRecorder.Eventf(object, corev1.EventTypeNormal, strings.Title(reason), message, args...)
}

// Event constructs a warning event from the given information and puts it in the queue for sending.
func Warn(object runtime.Object, reason, message string) {
	defaultRecorder.Event(object, corev1.EventTypeWarning, strings.Title(reason), message)
}

// Eventf is just like Event, but with Sprintf for the message field.
func Warnf(object runtime.Object, reason, message string, args ...interface{}) {
	defaultRecorder.Eventf(object, corev1.EventTypeWarning, strings.Title(reason), message, args...)
}
