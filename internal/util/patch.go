package util

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrPatch(ctx context.Context, c client.Client, o unstructured.Unstructured) error {
	nm := types.NamespacedName{
		Name:      o.GetName(),
		Namespace: o.GetNamespace(),
	}
	oc := o.DeepCopy()
	err := c.Get(ctx, nm, &o)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if err != nil {
		return c.Create(ctx, oc)
	}
	vo, ok, err := unstructured.NestedString(o.Object, "metadata", "resourceVersion")
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("old object resourceVersion not found")
	}
	v, ok, _ := unstructured.NestedString(oc.Object, "metadata", "resourceVersion")
	if v == "" && !ok {
		err = unstructured.SetNestedField(oc.Object, vo, "metadata", "resourceVersion")
		if err != nil {
			return err
		}
	}
	specGet, ok, err := unstructured.NestedFieldNoCopy(o.Object, "spec")
	if err != nil {
		return err
	}
	if !ok {
		err = c.Patch(ctx, oc, client.MergeFrom(&o))
		return err
	}
	specNew, ok, err := unstructured.NestedFieldNoCopy(oc.Object, "spec")
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("new object spec not found")
	}
	diff := cmp.Diff(specGet, specNew)
	if diff != "" {
		return c.Patch(ctx, oc, client.MergeFrom(&o))
	}
	return nil
}
