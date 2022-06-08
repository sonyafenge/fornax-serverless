/*
Copyright 2022 The fornax-serverless Authors.

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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	corev1 "centaurusinfra.io/fornax-serverless/pkg/apis/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClientSessions implements ClientSessionInterface
type FakeClientSessions struct {
	Fake *FakeCoreV1
	ns   string
}

var clientsessionsResource = schema.GroupVersionResource{Group: "core.fornax-serverless.centaurusinfra.io", Version: "v1", Resource: "clientsessions"}

var clientsessionsKind = schema.GroupVersionKind{Group: "core.fornax-serverless.centaurusinfra.io", Version: "v1", Kind: "ClientSession"}

// Get takes name of the clientSession, and returns the corresponding clientSession object, and an error if there is any.
func (c *FakeClientSessions) Get(ctx context.Context, name string, options v1.GetOptions) (result *corev1.ClientSession, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clientsessionsResource, c.ns, name), &corev1.ClientSession{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ClientSession), err
}

// List takes label and field selectors, and returns the list of ClientSessions that match those selectors.
func (c *FakeClientSessions) List(ctx context.Context, opts v1.ListOptions) (result *corev1.ClientSessionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clientsessionsResource, clientsessionsKind, c.ns, opts), &corev1.ClientSessionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &corev1.ClientSessionList{ListMeta: obj.(*corev1.ClientSessionList).ListMeta}
	for _, item := range obj.(*corev1.ClientSessionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clientSessions.
func (c *FakeClientSessions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(clientsessionsResource, c.ns, opts))

}

// Create takes the representation of a clientSession and creates it.  Returns the server's representation of the clientSession, and an error, if there is any.
func (c *FakeClientSessions) Create(ctx context.Context, clientSession *corev1.ClientSession, opts v1.CreateOptions) (result *corev1.ClientSession, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(clientsessionsResource, c.ns, clientSession), &corev1.ClientSession{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ClientSession), err
}

// Update takes the representation of a clientSession and updates it. Returns the server's representation of the clientSession, and an error, if there is any.
func (c *FakeClientSessions) Update(ctx context.Context, clientSession *corev1.ClientSession, opts v1.UpdateOptions) (result *corev1.ClientSession, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(clientsessionsResource, c.ns, clientSession), &corev1.ClientSession{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ClientSession), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeClientSessions) UpdateStatus(ctx context.Context, clientSession *corev1.ClientSession, opts v1.UpdateOptions) (*corev1.ClientSession, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(clientsessionsResource, "status", c.ns, clientSession), &corev1.ClientSession{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ClientSession), err
}

// Delete takes name of the clientSession and deletes it. Returns an error if one occurs.
func (c *FakeClientSessions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(clientsessionsResource, c.ns, name, opts), &corev1.ClientSession{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClientSessions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(clientsessionsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &corev1.ClientSessionList{})
	return err
}

// Patch applies the patch and returns the patched clientSession.
func (c *FakeClientSessions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *corev1.ClientSession, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clientsessionsResource, c.ns, name, pt, data, subresources...), &corev1.ClientSession{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ClientSession), err
}
