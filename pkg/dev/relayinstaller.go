package dev

import (
	"context"

	"github.com/puppetlabs/relay/pkg/cluster"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	relayInstallerImage = "relaysh/relay-install-operator:latest"
)

type relayInstallerObjects struct {
	serviceAccount     corev1.ServiceAccount
	clusterRole        rbacv1.ClusterRole
	clusterRoleBinding rbacv1.ClusterRoleBinding
	deployment         appsv1.Deployment
}

func newRelayInstallerObjects() *relayInstallerObjects {
	objectMeta := metav1.ObjectMeta{
		Name:      "relay-install-operator",
		Namespace: systemNamespace,
	}

	return &relayInstallerObjects{
		serviceAccount:     corev1.ServiceAccount{ObjectMeta: objectMeta},
		clusterRole:        rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: objectMeta.Name}},
		clusterRoleBinding: rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: objectMeta.Name}},
		deployment:         appsv1.Deployment{ObjectMeta: objectMeta},
	}
}

type relayInstallerManager struct {
	cl      *cluster.Client
	objects *relayInstallerObjects
}

func (m *relayInstallerManager) reconcile(ctx context.Context) error {
	cl := m.cl.APIClient

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.serviceAccount, func() error {
		m.serviceAccount(&m.objects.serviceAccount)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.clusterRole, func() error {
		m.clusterRole(&m.objects.clusterRole)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.clusterRoleBinding, func() error {
		m.clusterRoleBinding(&m.objects.clusterRoleBinding)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.deployment, func() error {
		m.deployment(&m.objects.deployment)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *relayInstallerManager) serviceAccount(sa *corev1.ServiceAccount) {
	sa.Labels = m.labels()
}

// clusterRole configures the roles requires for the installer to run. It has
// to delegate roles and bindings to relay-operator and relay-metadata-api via
// the creation of clusterroles/roles and clusterrolebindings/rolebindings, so
// in order to do that, it itself needs bindings to a large amount of resources
// and verbs. It would be nice if this was autogenerated somehow, because as we
// change relay-core controllers, we are going to need to reflect those rbac
// policies here as well.
func (m *relayInstallerManager) clusterRole(cr *rbacv1.ClusterRole) {
	cr.Labels = m.labels()

	verbAllGroups := []string{
		"",
		"networking.k8s.io",
		"rbac.authorization.k8s.io",
		"install.relay.sh",
		"tekton.dev",
		"serving.knative.dev",
	}

	verbAllResources := []string{
		"configmaps",
		"limitranges",
		"namespaces",
		"secrets",
		"serviceaccounts",
		"revisions",
		"services",
		"networkpolicies",
		"roles",
		"rolebindings",
		"relaycores",
		"conditions",
		"pipelineruns",
		"pipelines",
		"taskruns",
		"tasks",
	}

	cr.Rules = []rbacv1.PolicyRule{
		{APIGroups: verbAllGroups, Resources: verbAllResources, Verbs: []string{rbacv1.VerbAll}},
		{
			APIGroups: []string{"apps", "rbac.authorization.k8s.io", "admissionregistration.k8s.io"},
			Resources: []string{"deployments", "clusterroles", "clusterrolebindings", "mutatingwebhookconfigurations"},
			Verbs:     []string{"create", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "pods/log"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"install.relay.sh"},
			Resources: []string{"relaycores/status"},
			Verbs:     []string{"get", "patch", "update"},
		},
		{
			APIGroups: []string{"nebula.puppet.com"},
			Resources: []string{"workflowruns", "workflowruns/status"},
			Verbs:     []string{"get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"pvpool.puppet.com"},
			Resources: []string{"checkouts", "checkouts/status"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"relay.sh"},
			Resources: []string{"tenants", "tenants/status", "webhooktriggers", "webhooktriggers/status"},
			Verbs:     []string{"get", "list", "patch", "update", "watch"},
		},
	}
}

func (m *relayInstallerManager) clusterRoleBinding(crb *rbacv1.ClusterRoleBinding) {
	crb.Labels = m.labels()

	crb.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     m.objects.clusterRole.Name,
	}

	crb.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      m.objects.serviceAccount.Name,
			Namespace: m.objects.serviceAccount.Namespace,
		},
	}
}

func (m *relayInstallerManager) deployment(deployment *appsv1.Deployment) {
	deployment.Labels = m.labels()
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: m.labels(),
	}

	template := &deployment.Spec.Template
	template.Labels = m.labels()

	template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	template.Spec.ServiceAccountName = m.objects.serviceAccount.Name

	container := corev1.Container{
		Name:            "controller",
		Image:           relayInstallerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
	}

	template.Spec.Containers = []corev1.Container{container}
}

func (m *relayInstallerManager) labels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "relay-install-operator",
		"app.kubernetes.io/component": "controller",
	}
}

func newRelayInstallerManager(cl *cluster.Client) *relayInstallerManager {
	return &relayInstallerManager{
		cl:      cl,
		objects: newRelayInstallerObjects(),
	}
}
