/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sync

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/api"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"google.golang.org/api/googleapi"
	"k8s.io/klog/v2"
)

// genericOps are a typed dispatch for (API version, scope) for CRUD verbs.
type genericOps[GA any, Alpha any, Beta any] interface {
	getFuncs(gcp cloud.Cloud) *getFuncs[GA, Alpha, Beta]
	createFuncs(gcp cloud.Cloud) *createFuncs[GA, Alpha, Beta]
	updateFuncs(gcp cloud.Cloud) *updateFuncs[GA, Alpha, Beta]
	deleteFuncs(gcp cloud.Cloud) *deleteFuncs[GA, Alpha, Beta]
}

type getFuncsByScope[T any] struct {
	global   func(context.Context, *meta.Key) (*T, error)
	regional func(context.Context, *meta.Key) (*T, error)
	zonal    func(context.Context, *meta.Key) (*T, error)
}

func (s *getFuncsByScope[T]) do(ctx context.Context, key *meta.Key) (*T, error) {
	klog.Infof("get %s", key)
	switch {
	case key.Type() == meta.Global && s.global != nil:
		return s.global(ctx, key)
	case key.Type() == meta.Regional && s.regional != nil:
		return s.regional(ctx, key)
	case key.Type() == meta.Zonal && s.zonal != nil:
		return s.zonal(ctx, key)
	}
	return nil, fmt.Errorf("unsupported scope (key=%s)", key)
}

type getFuncs[GA any, Alpha any, Beta any] struct {
	ga    getFuncsByScope[GA]
	alpha getFuncsByScope[Alpha]
	beta  getFuncsByScope[Beta]
}

func (f *getFuncs[GA, Alpha, Beta]) do(
	ctx context.Context,
	ver meta.Version,
	id *cloud.ResourceID,
	tt api.TypeTrait[GA, Alpha, Beta],
) (api.FrozenResource[GA, Alpha, Beta], error) {
	current := api.NewResource(id, tt)
	switch ver {
	case meta.VersionGA:
		raw, err := f.ga.do(ctx, id.Key)
		if err != nil {
			return nil, err
		}
		if err := current.Set(raw); err != nil {
			return nil, err
		}
	case meta.VersionAlpha:
		raw, err := f.alpha.do(ctx, id.Key)
		if err != nil {
			return nil, err
		}
		if err := current.SetAlpha(raw); err != nil {
			return nil, err
		}
	case meta.VersionBeta:
		raw, err := f.beta.do(ctx, id.Key)
		if err != nil {
			return nil, err
		}
		if err := current.SetBeta(raw); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("getFuncs.do unsupported version %s", ver)
	}
	return current.Freeze()
}

type createFuncsByScope[T any] struct {
	global   func(context.Context, *meta.Key, *T) error
	regional func(context.Context, *meta.Key, *T) error
	zonal    func(context.Context, *meta.Key, *T) error
}

func (s *createFuncsByScope[T]) do(ctx context.Context, key *meta.Key, x *T) error {
	// TODO: handle project routing.
	// TODO: Context logging
	// TODO: span
	switch {
	case key.Type() == meta.Global && s.global != nil:
		return s.global(ctx, key, x)
	case key.Type() == meta.Regional && s.regional != nil:
		return s.regional(ctx, key, x)
	case key.Type() == meta.Zonal && s.zonal != nil:
		return s.zonal(ctx, key, x)
	}
	return fmt.Errorf("unsupported scope (key = %s)", key)
}

type createFuncs[GA any, Alpha any, Beta any] struct {
	ga    createFuncsByScope[GA]
	alpha createFuncsByScope[Alpha]
	beta  createFuncsByScope[Beta]
}

func (f *createFuncs[GA, Alpha, Beta]) do(
	ctx context.Context,
	id *cloud.ResourceID,
	r api.FrozenResource[GA, Alpha, Beta],
) error {
	// TODO: handle project routing.
	// TODO: Context logging
	// TODO: span
	switch r.Version() {
	case meta.VersionGA:
		raw, err := r.ToGA()
		if err != nil {
			return err
		}
		err = f.ga.do(ctx, id.Key, raw)
		if err != nil {
			return err
		}
		return nil
	case meta.VersionAlpha:
		raw, err := r.ToAlpha()
		if err != nil {
			return err
		}
		err = f.alpha.do(ctx, id.Key, raw)
		if err != nil {
			return err
		}
		return nil
	case meta.VersionBeta:
		raw, err := r.ToBeta()
		if err != nil {
			return err
		}
		err = f.beta.do(ctx, id.Key, raw)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("createFuncs.do unsupported version %s", r.Version())
}

type updateFuncsByScope[T any] struct {
	global   func(context.Context, *meta.Key, *T) error
	regional func(context.Context, *meta.Key, *T) error
	zonal    func(context.Context, *meta.Key, *T) error
}

func (s *updateFuncsByScope[T]) do(ctx context.Context, key *meta.Key, x *T) error {
	switch {
	case key.Type() == meta.Global && s.global != nil:
		return s.global(ctx, key, x)
	case key.Type() == meta.Regional && s.regional != nil:
		return s.regional(ctx, key, x)
	case key.Type() == meta.Zonal && s.zonal != nil:
		return s.zonal(ctx, key, x)
	}
	return fmt.Errorf("unsupported scope (key = %s)", key)
}

const (
	// Resource does not have a .Fingerprint field. Note: this
	// means that the resource is technically not compliant with
	// API conventions but these exceptions occur throughout the
	// GCE APIs and we have to work around them.
	updateFuncsNoFingerprint = 1 << iota
)

type updateFuncs[GA any, Alpha any, Beta any] struct {
	ga    updateFuncsByScope[GA]
	alpha updateFuncsByScope[Alpha]
	beta  updateFuncsByScope[Beta]

	options int
}

func fingerprintField(v reflect.Value) (reflect.Value, error) {
	typeCheck := func(v reflect.Value) error {
		t := v.Type()
		if !(t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct) {
			return fmt.Errorf("fingerprintField: invalid type %T", v.Interface())
		}
		if fv := v.Elem().FieldByName("Fingerprint"); !fv.IsValid() || fv.Kind() != reflect.String {
			return fmt.Errorf("fingerprintField: no Fingerprint field (%T)", v.Interface())
		}
		return nil
	}
	if err := typeCheck(v); err != nil {
		return reflect.Value{}, err
	}
	return v.Elem().FieldByName("Fingerprint"), nil
}

func (f *updateFuncs[GA, Alpha, Beta]) do(
	ctx context.Context,
	fingerprint string,
	id *cloud.ResourceID,
	desired api.FrozenResource[GA, Alpha, Beta],
) error {
	// TODO: handle project routing.
	// TODO: Context logging
	// TODO: span
	switch desired.Version() {
	case meta.VersionGA:
		raw, err := desired.ToGA()
		if err != nil {
			return err
		}
		if f.options&updateFuncsNoFingerprint == 0 {
			// TODO: we need to make sure this is the right way to do this as it
			// modifies the FrozenResource. Patch fingerprint for the update.
			if fv, err := fingerprintField(reflect.ValueOf(raw)); err != nil {
				return err
			} else {
				fv.Set(reflect.ValueOf(fingerprint))
			}
		}
		err = f.ga.do(ctx, id.Key, raw)
		if err != nil {
			return err
		}
		return nil

	case meta.VersionAlpha:
		raw, err := desired.ToAlpha()
		if err != nil {
			return err
		}
		if f.options&updateFuncsNoFingerprint == 0 {
			if fv, err := fingerprintField(reflect.ValueOf(raw)); err != nil {
				return err
			} else {
				fv.Set(reflect.ValueOf(fingerprint))
			}
		}
		err = f.alpha.do(ctx, id.Key, raw)
		if err != nil {
			return err
		}
		return nil

	case meta.VersionBeta:
		raw, err := desired.ToBeta()
		if err != nil {
			return err
		}
		if f.options&updateFuncsNoFingerprint == 0 {
			if fv, err := fingerprintField(reflect.ValueOf(raw)); err != nil {
				return err
			} else {
				fv.Set(reflect.ValueOf(fingerprint))
			}
		}
		err = f.beta.do(ctx, id.Key, raw)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("updateFuncs.do unsupported version %s", desired.Version())
}

type deleteFuncsByScope[T any] struct {
	global   func(context.Context, *meta.Key) error
	regional func(context.Context, *meta.Key) error
	zonal    func(context.Context, *meta.Key) error
}

func (s *deleteFuncsByScope[T]) do(ctx context.Context, id *cloud.ResourceID) error {
	key := id.Key
	switch {
	case key.Type() == meta.Global && s.global != nil:
		return s.global(ctx, key)
	case key.Type() == meta.Regional && s.regional != nil:
		return s.regional(ctx, key)
	case key.Type() == meta.Zonal && s.zonal != nil:
		return s.zonal(ctx, key)
	}
	return fmt.Errorf("unsupported scope (key = %s)", key)
}

type deleteFuncs[GA any, Alpha any, Beta any] struct {
	ga    deleteFuncsByScope[GA]
	alpha deleteFuncsByScope[Alpha]
	beta  deleteFuncsByScope[Beta]
}

func (f *deleteFuncs[GA, Alpha, Beta]) do(ctx context.Context, id *cloud.ResourceID) error {
	// TODO: handle project routing.
	// TODO: Context logging
	// TODO: span
	return f.ga.do(ctx, id)
}

func isErrorCode(err error, code int) bool {
	var gerr *googleapi.Error
	if !errors.As(err, &gerr) {
		return false
	}
	return gerr.Code == code
}

func isErrorNotFound(err error) bool { return isErrorCode(err, 404) }

func genericGet[GA any, Alpha any, Beta any](
	ctx context.Context,
	gcp cloud.Cloud,
	resourceName string,
	ops genericOps[GA, Alpha, Beta],
	node *nodeBase[GA, Alpha, Beta],
) error {
	r, err := ops.getFuncs(gcp).do(ctx, node.getVer, node.ID(), node.typeTrait)

	switch {
	case isErrorNotFound(err):
		node.state = NodeDoesNotExist
		return nil // Not found is not an error condition.

	case err != nil:
		node.state = NodeStateError
		node.getErr = err
		return fmt.Errorf("genericGet %s: %w", resourceName, err)

	default:
		node.state = NodeExists
		node.resource = r
		return nil
	}
}
