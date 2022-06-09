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
// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "centaurusinfra.io/fornax-serverless/pkg/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClientSessionLister helps list ClientSessions.
// All objects returned here must be treated as read-only.
type ClientSessionLister interface {
	// List lists all ClientSessions in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.ClientSession, err error)
	// ClientSessions returns an object that can list and get ClientSessions.
	ClientSessions(namespace string) ClientSessionNamespaceLister
	ClientSessionListerExpansion
}

// clientSessionLister implements the ClientSessionLister interface.
type clientSessionLister struct {
	indexer cache.Indexer
}

// NewClientSessionLister returns a new ClientSessionLister.
func NewClientSessionLister(indexer cache.Indexer) ClientSessionLister {
	return &clientSessionLister{indexer: indexer}
}

// List lists all ClientSessions in the indexer.
func (s *clientSessionLister) List(selector labels.Selector) (ret []*v1.ClientSession, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ClientSession))
	})
	return ret, err
}

// ClientSessions returns an object that can list and get ClientSessions.
func (s *clientSessionLister) ClientSessions(namespace string) ClientSessionNamespaceLister {
	return clientSessionNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ClientSessionNamespaceLister helps list and get ClientSessions.
// All objects returned here must be treated as read-only.
type ClientSessionNamespaceLister interface {
	// List lists all ClientSessions in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.ClientSession, err error)
	// Get retrieves the ClientSession from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.ClientSession, error)
	ClientSessionNamespaceListerExpansion
}

// clientSessionNamespaceLister implements the ClientSessionNamespaceLister
// interface.
type clientSessionNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ClientSessions in the indexer for a given namespace.
func (s clientSessionNamespaceLister) List(selector labels.Selector) (ret []*v1.ClientSession, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ClientSession))
	})
	return ret, err
}

// Get retrieves the ClientSession from the indexer for a given namespace and name.
func (s clientSessionNamespaceLister) Get(name string) (*v1.ClientSession, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("clientsession"), name)
	}
	return obj.(*v1.ClientSession), nil
}
