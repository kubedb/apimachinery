/*
Copyright AppsCode Inc. and Contributors

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

package client

import (
	"context"
	"reflect"
	"strings"

	"kmodules.xyz/client-go/meta"

	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func NewUncachedClient(cfg *rest.Config, funcs ...func(*runtime.Scheme) error) (client.Client, error) {
	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, hc)
	if err != nil {
		return nil, err
	}

	builder := runtime.NewSchemeBuilder(funcs...)
	builder.Register(clientgoscheme.AddToScheme)
	scheme := runtime.NewScheme()
	err = builder.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

type (
	TransformFunc  func(obj client.Object, createOp bool) client.Object
	TransformFuncE func(obj client.Object, createOp bool) (client.Object, error)
	PatchFunc      func(obj client.Object) client.Object
	PatchFuncE     func(obj client.Object) (client.Object, error)
)

func CreateOrPatchE(ctx context.Context, c client.Client, obj client.Object, transform TransformFuncE, opts ...client.PatchOption) (kutil.VerbType, error) {
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return kutil.VerbUnchanged, errors.Wrapf(err, "failed to get GVK for object %T", obj)
	}

	cur := obj.DeepCopyObject().(client.Object)
	key := types.NamespacedName{
		Namespace: cur.GetNamespace(),
		Name:      cur.GetName(),
	}
	err = c.Get(ctx, key, cur)
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating %+v %s/%s.", gvk, key.Namespace, key.Name)

		createOpts := make([]client.CreateOption, 0, len(opts))
		for i := range opts {
			if opt, ok := opts[i].(client.CreateOption); ok {
				createOpts = append(createOpts, opt)
			}
		}
		mod, err := transform(obj.DeepCopyObject().(client.Object), true)
		if err != nil {
			return kutil.VerbUnchanged, err
		}
		mod.SetResourceVersion("")
		err = c.Create(ctx, mod, createOpts...)
		if err != nil {
			return kutil.VerbUnchanged, err
		}

		assign(obj, mod)
		return kutil.VerbCreated, err
	} else if err != nil {
		return kutil.VerbUnchanged, err
	}

	_, unstructuredObj := obj.(*unstructured.Unstructured)

	var patch client.Patch
	if isOfficialTypes(gvk.Group) && !unstructuredObj {
		patch = client.StrategicMergeFrom(cur)
	} else {
		patch = client.MergeFrom(cur)
	}
	mod, err := transform(cur.DeepCopyObject().(client.Object), false)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	err = c.Patch(ctx, mod, patch, opts...)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	vt := kutil.VerbUnchanged
	if mod.GetGeneration() > 0 {
		if cur.GetGeneration() != mod.GetGeneration() {
			vt = kutil.VerbPatched
		}
	} else {
		// Secret, ServiceAccount etc resources do not use metadata.generation
		if meta.ObjectHash(cur) != meta.ObjectHash(mod) {
			vt = kutil.VerbPatched
		}
	}
	assign(obj, mod)
	return vt, nil
}

func CreateOrPatch(ctx context.Context, c client.Client, obj client.Object, transform TransformFunc, opts ...client.PatchOption) (kutil.VerbType, error) {
	return CreateOrPatchE(ctx, c, obj, func(obj client.Object, createOp bool) (client.Object, error) {
		return transform(obj, createOp), nil
	}, opts...)
}

func PatchE(ctx context.Context, c client.Client, obj client.Object, transform PatchFuncE, opts ...client.PatchOption) (kutil.VerbType, error) {
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return kutil.VerbUnchanged, errors.Wrapf(err, "failed to get GVK for object %T", obj)
	}

	_, unstructuredObj := obj.(*unstructured.Unstructured)

	var patch client.Patch
	if isOfficialTypes(gvk.Group) && !unstructuredObj {
		patch = client.StrategicMergeFrom(obj)
	} else {
		patch = client.MergeFrom(obj)
	}
	mod, err := transform(obj.DeepCopyObject().(client.Object))
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	err = c.Patch(ctx, mod, patch, opts...)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	vt := kutil.VerbUnchanged
	if mod.GetGeneration() > 0 {
		if obj.GetGeneration() != mod.GetGeneration() {
			vt = kutil.VerbPatched
		}
	} else {
		// Secret, ServiceAccount etc resources do not use metadata.generation
		if meta.ObjectHash(obj) != meta.ObjectHash(mod) {
			vt = kutil.VerbPatched
		}
	}
	assign(obj, mod)
	return vt, nil
}

func Patch(ctx context.Context, c client.Client, obj client.Object, transform PatchFunc, opts ...client.PatchOption) (kutil.VerbType, error) {
	return PatchE(ctx, c, obj, func(obj client.Object) (client.Object, error) {
		return transform(obj), nil
	}, opts...)
}

func assign(target, src any) {
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Pointer {
		srcValue = srcValue.Elem()
	}
	reflect.ValueOf(target).Elem().Set(srcValue)
}

func PatchStatusE(ctx context.Context, c client.Client, obj client.Object, transform PatchFuncE, opts ...client.SubResourcePatchOption) (kutil.VerbType, error) {
	cur := obj.DeepCopyObject().(client.Object)
	key := types.NamespacedName{
		Namespace: cur.GetNamespace(),
		Name:      cur.GetName(),
	}
	err := c.Get(ctx, key, cur)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// The body of the request was in an unknown format -
	// accepted media types include:
	//   - application/json-patch+json,
	//   - application/merge-patch+json,
	//   - application/apply-patch+yaml
	patch := client.MergeFrom(cur)
	mod, err := transform(cur.DeepCopyObject().(client.Object))
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	err = c.Status().Patch(ctx, mod, patch, opts...)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	assign(obj, mod)
	return kutil.VerbPatched, nil
}

func PatchStatus(ctx context.Context, c client.Client, obj client.Object, transform PatchFunc, opts ...client.SubResourcePatchOption) (kutil.VerbType, error) {
	return PatchStatusE(ctx, c, obj, func(obj client.Object) (client.Object, error) {
		return transform(obj), nil
	}, opts...)
}

func isOfficialTypes(group string) bool {
	return !strings.ContainsRune(group, '.')
}

func GetForGVR(ctx context.Context, c client.Client, gvr schema.GroupVersionResource, ref types.NamespacedName) (client.Object, error) {
	gvk, err := c.RESTMapper().KindFor(gvr)
	if err != nil {
		return nil, err
	}
	o, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}
	obj := o.(client.Object)
	err = c.Get(ctx, ref, obj)
	return obj, err
}

func GetForGVK(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, ref types.NamespacedName) (client.Object, error) {
	if gvk.Version == "" {
		mapping, err := c.RESTMapper().RESTMapping(gvk.GroupKind())
		if err != nil {
			return nil, err
		}
		gvk = mapping.GroupVersionKind
	}
	o, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}
	obj := o.(client.Object)
	err = c.Get(ctx, ref, obj)
	return obj, err
}
