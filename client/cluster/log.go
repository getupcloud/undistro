/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cluster

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/internal/scheme"
	"github.com/getupcloud/undistro/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	undistroNamespace = "undistro-system"
)

type StopConditionFunc func(context.Context, client.Client, types.NamespacedName) (bool, error)

type LogStreamer interface {
	Stream(context.Context, *rest.Config, io.Writer, types.NamespacedName, StopConditionFunc) error
}

type logStreamer struct {
	client     *kubernetes.Clientset
	kubeClient client.Client
}

func (l *logStreamer) Stream(ctx context.Context, c *rest.Config, w io.Writer, nm types.NamespacedName, fn StopConditionFunc) error {
	var (
		err    error
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithCancel(ctx)
	c.Timeout = 0
	l.kubeClient, err = client.New(c, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		cancel()
		return err
	}
	l.client, err = kubernetes.NewForConfig(c)
	if err != nil {
		cancel()
		return err
	}
	var podList corev1.PodList
	err = l.kubeClient.List(ctx, &podList, client.InNamespace(undistroNamespace))
	if err != nil {
		cancel()
		return err
	}
	logs := l.getLogs(ctx, podList.Items)
	cerr := make(chan error, 1)
	go func(ctx context.Context, cancel context.CancelFunc, fn StopConditionFunc) {
		ready, err := fn(ctx, l.kubeClient, nm)
		if err != nil {
			cerr <- err
			return
		}
		for !ready {
			ready, err = fn(ctx, l.kubeClient, nm)
			if err != nil {
				cerr <- err
				return
			}
			<-time.After(10 * time.Second)
		}
		cancel()
	}(ctx, cancel, fn)
	go func(ctx context.Context) {
		cerr <- l.streamLogs(ctx, logs, w)
	}(ctx)
	select {
	case err = <-cerr:
		return err
	case <-ctx.Done():
		return nil
	}
}

func containReq(reqs []logRequest, pod string) bool {
	for _, req := range reqs {
		if req.podName == pod {
			return true
		}
	}
	return false
}

func (l *logStreamer) streamLogs(ctx context.Context, reqs []logRequest, writer io.Writer) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(reqs))
	r, w := io.Pipe()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(l.client, time.Second*30)
	podsInformer := kubeInformerFactory.Core().V1().Pods().Informer()
	podsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			if pod.Namespace != undistroNamespace || containReq(reqs, pod.Name) {
				return
			}
			var (
				reader io.ReadCloser
				err    error
			)
			err = retryWithExponentialBackoff(newConnectBackoff(), func() error {
				reader, err = l.client.CoreV1().Pods(undistroNamespace).
					GetLogs(pod.Name, &corev1.PodLogOptions{
						Container: "manager",
						Follow:    true,
						SinceTime: &metav1.Time{
							Time: time.Now(),
						},
					}).Stream(ctx)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				log.Log.Error(err, "fail to get pod logs", "name", pod.Name)
				return
			}
			wg.Add(1)
			go l.streamLog(ctx, wg, reader, w, pod.Name)
		},
	})
	stop := make(chan struct{})
	kubeInformerFactory.Start(stop)
	for _, req := range reqs {
		reader, err := req.req.Stream(ctx)
		if err != nil {
			return err
		}
		go l.streamLog(ctx, wg, reader, w, req.podName)
	}
	go func(ctx context.Context) {
		wg.Wait()
		close(stop)
		w.Close()
	}(ctx)
	_, err := io.Copy(writer, r)
	return err
}

func (l *logStreamer) readLog(ctx context.Context, reader io.ReadCloser, writer io.Writer) error {
	defer reader.Close()
	r := bufio.NewReader(reader)
	for {
		byt, err := r.ReadBytes('\n')
		if _, err := writer.Write(byt); err != nil {
			return err
		}
		if err != nil {
			if err != io.EOF && err != context.Canceled && err != context.DeadlineExceeded {
				return err
			}
			return nil
		}
	}
}

func IsReady(ctx context.Context, c client.Client, nm types.NamespacedName) (bool, error) {
	var cl undistrov1.Cluster
	err := c.Get(ctx, nm, &cl)
	if err != nil {
		return false, err
	}
	return cl.Status.Ready, nil
}

func IsDeleted(ctx context.Context, c client.Client, nm types.NamespacedName) (bool, error) {
	var cl undistrov1.Cluster
	err := c.Get(ctx, nm, &cl)
	return apierrors.IsNotFound(err), nil
}

func (l logStreamer) streamLog(ctx context.Context, wg *sync.WaitGroup, reader io.ReadCloser, w io.Writer, pod string) {
	defer wg.Done()
	pw := &prefixingWriter{
		writer: w,
		prefix: []byte(fmt.Sprintf("[%s] ", pod)),
	}
	err := l.readLog(ctx, reader, pw)
	if err != nil {
		log.Log.Error(err, "fail read log", "name", pod)
	}
}

func (l *logStreamer) getLogs(ctx context.Context, pods []corev1.Pod) []logRequest {
	reqs := make([]logRequest, len(pods))
	for i, pod := range pods {
		req := l.client.
			CoreV1().
			Pods(undistroNamespace).
			GetLogs(pod.Name, &corev1.PodLogOptions{
				Container: "manager",
				Follow:    true,
				SinceTime: &metav1.Time{
					Time: time.Now(),
				},
			})
		reqs[i] = logRequest{
			podName: pod.Name,
			req:     req,
		}
	}
	return reqs
}

func NewLogStreamer() *logStreamer {
	return &logStreamer{}
}

type logRequest struct {
	podName string
	req     *rest.Request
}

type prefixingWriter struct {
	prefix []byte
	writer io.Writer
}

func (pw *prefixingWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Perform an "atomic" write of a prefix and p to make sure that it doesn't interleave
	// sub-line when used concurrently with io.PipeWrite.
	n, err := pw.writer.Write(append(pw.prefix, p...))
	if n > len(p) {
		// To comply with the io.Writer interface requirements we must
		// return a number of bytes written from p (0 <= n <= len(p)),
		// so we are ignoring the length of the prefix here.
		return len(p), err
	}
	return n, err
}
