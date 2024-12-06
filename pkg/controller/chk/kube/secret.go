// Copyright 2019 Altinity Ltd and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	"github.com/altinity/clickhouse-operator/pkg/interfaces"
)

type Secret struct {
	kubeClient client.Client
	namer      interfaces.INameManager
}

func NewSecret(kubeClient client.Client, namer interfaces.INameManager) *Secret {
	return &Secret{
		kubeClient: kubeClient,
		namer:      namer,
	}
}

// Get gets Service. Accepted types:
//  1. *core.Service
//  2. *chop.Host
func (c *Secret) Get(ctx context.Context, params ...any) (*core.Secret, error) {
	var name, namespace string
	switch len(params) {
	case 2:
		// Expecting namespace name
		namespace = params[0].(string)
		name = params[1].(string)
	case 1:
		// Expecting obj
		obj := params[0]
		switch typedObj := obj.(type) {
		case *core.Secret:
			name = typedObj.Name
			namespace = typedObj.Namespace
		case *api.Host:
			name = c.namer.Name(interfaces.NameStatefulSetService, typedObj)
			namespace = typedObj.Runtime.Address.Namespace
		}
	}
	service := &core.Secret{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, service)
	if err == nil {
		return service, nil
	} else {
		return nil, err
	}
}

func (c *Secret) Create(ctx context.Context, svc *core.Secret) (*core.Secret, error) {
	err := c.kubeClient.Create(ctx, svc)
	return svc, err
}

func (c *Secret) Update(ctx context.Context, svc *core.Secret) (*core.Secret, error) {
	err := c.kubeClient.Update(ctx, svc)
	return svc, err
}

func (c *Secret) Delete(ctx context.Context, namespace, name string) error {
	svc := &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return c.kubeClient.Delete(ctx, svc)
}

func (c *Secret) List(ctx context.Context, namespace string, opts meta.ListOptions) ([]core.Secret, error) {
	list := &core.SecretList{}
	selector, err := labels.Parse(opts.LabelSelector)
	if err != nil {
		return nil, err
	}
	err = c.kubeClient.List(ctx, list, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, err
	}
	return list.Items, nil
}