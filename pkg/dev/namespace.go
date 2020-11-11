package dev

import (
	"context"

	"github.com/puppetlabs/relay/pkg/cluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemNamespace   = "relay-system"
	registryNamespace = "docker-registry"
)

type namespaceManager struct {
	cl *cluster.Client
}

func (m *namespaceManager) create(ctx context.Context, ns string) error {
	sn := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}

	if ns == systemNamespace {
		sn.Labels = map[string]string{
			"nebula.puppet.com/network-policy.tasks": "true",
		}
	}

	if err := m.cl.APIClient.Create(ctx, sn); err != nil {
		return err
	}

	return nil
}

func (m *namespaceManager) objectNamespacePatcher(name string) objectPatcherFunc {
	return func(obj runtime.Object) {
		var gvk schema.GroupVersionKind

		gvks, _, err := DefaultScheme.ObjectKinds(obj)
		if err != nil {
			return
		}

		if len(gvks) > 1 {
			return
		}

		gvk = gvks[0]

		mapping, err := m.cl.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return
		}

		a, err := meta.Accessor(obj)
		if err != nil {
			return
		}

		if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
			return
		}

		if a.GetNamespace() == "" {
			if name == "" {
				a.SetNamespace("default")
			} else {
				a.SetNamespace(name)
			}
		}
	}
}

func (m *namespaceManager) delete(ctx context.Context, ns string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}

	return m.cl.APIClient.Delete(ctx, namespace, client.PropagationPolicy(metav1.DeletePropagationBackground))
}

func newNamespaceManager(cl *cluster.Client) *namespaceManager {
	return &namespaceManager{cl: cl}
}
