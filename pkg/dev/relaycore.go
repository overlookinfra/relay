package dev

import (
	"context"
	"errors"
	"fmt"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/puppetlabs/leg/timeutil/pkg/retry"
	installerv1alpha1 "github.com/puppetlabs/relay-core/pkg/apis/install.relay.sh/v1alpha1"
	"github.com/puppetlabs/relay/pkg/cluster"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	relayCoreName         = "relay-core-v1"
	relayLogServiceImage  = "relaysh/relay-pls:latest"
	relayOperatorImage    = "relaysh/relay-operator:latest"
	relayMetadataAPIImage = "relaysh/relay-metadata-api:latest"

	relayOperatorStorageVolumeSize = "1Gi"
)

type relayCoreObjects struct {
	selfSignedIssuer    certmanagerv1beta1.Issuer
	selfSignedCA        certmanagerv1beta1.Certificate
	issuer              certmanagerv1beta1.Issuer
	operatorWebhookCert certmanagerv1beta1.Certificate
	pvc                 corev1.PersistentVolumeClaim
	relayCore           installerv1alpha1.RelayCore
	rbacDefinition      rbacmanagerv1beta1.RBACDefinition
}

func newRelayCoreObjects() *relayCoreObjects {
	objectMeta := metav1.ObjectMeta{
		Name:      relayCoreName,
		Namespace: systemNamespace,
	}

	selfSignedObjectMeta := objectMeta
	selfSignedObjectMeta.Name = fmt.Sprintf("%s-self-signed", objectMeta.Name)

	operatorObjectMeta := objectMeta
	operatorObjectMeta.Name = fmt.Sprintf("%s-operator", objectMeta.Name)

	return &relayCoreObjects{
		selfSignedIssuer:    certmanagerv1beta1.Issuer{ObjectMeta: selfSignedObjectMeta},
		selfSignedCA:        certmanagerv1beta1.Certificate{ObjectMeta: selfSignedObjectMeta},
		issuer:              certmanagerv1beta1.Issuer{ObjectMeta: objectMeta},
		operatorWebhookCert: certmanagerv1beta1.Certificate{ObjectMeta: operatorObjectMeta},
		pvc:                 corev1.PersistentVolumeClaim{ObjectMeta: operatorObjectMeta},
		relayCore:           installerv1alpha1.RelayCore{ObjectMeta: objectMeta},
		rbacDefinition:      rbacmanagerv1beta1.RBACDefinition{ObjectMeta: operatorObjectMeta},
	}
}

type relayCoreManager struct {
	cl             *cluster.Client
	objects        *relayCoreObjects
	logServiceOpts LogServiceOptions
}

func (m *relayCoreManager) reconcile(ctx context.Context) error {
	cl := m.cl.APIClient

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.selfSignedIssuer, func() error {
		m.selfSignedIssuer(&m.objects.selfSignedIssuer)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.selfSignedCA, func() error {
		m.selfSignedCA(&m.objects.selfSignedCA)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.issuer, func() error {
		m.issuer(&m.objects.issuer)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.operatorWebhookCert, func() error {
		m.operatorWebhookCert(&m.objects.operatorWebhookCert)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.pvc, func() error {
		m.operatorStoragePVC(&m.objects.pvc)

		return nil
	}); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.relayCore, func() error {
		m.relayCore(&m.objects.relayCore)

		return nil
	}); err != nil {
		return err
	}

	rcKey, err := client.ObjectKeyFromObject(&m.objects.relayCore)
	if err != nil {
		return err
	}

	if err := cl.Get(ctx, rcKey, &m.objects.relayCore); err != nil {
		return err
	}

	if err := m.wait(ctx); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, cl, &m.objects.rbacDefinition, func() error {
		m.rbacDefinition(&m.objects.rbacDefinition)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *relayCoreManager) selfSignedIssuer(issuer *certmanagerv1beta1.Issuer) {
	issuer.Spec.SelfSigned = &certmanagerv1beta1.SelfSignedIssuer{}
}

func (m *relayCoreManager) selfSignedCA(cert *certmanagerv1beta1.Certificate) {
	cert.Spec.SecretName = fmt.Sprintf("%s-ca-tls", cert.Name)
	cert.Spec.CommonName = fmt.Sprintf("%s.svc.cluster.local", cert.Namespace)
	cert.Spec.DNSNames = append(cert.Spec.DNSNames,
		fmt.Sprintf("%s.svc", cert.Namespace),
		fmt.Sprintf("%s.local", cert.Namespace),
	)
	cert.Spec.IsCA = true
	cert.Spec.IssuerRef = certmanagermetav1.ObjectReference{
		Name: m.objects.selfSignedIssuer.Name,
	}
}

func (m *relayCoreManager) issuer(issuer *certmanagerv1beta1.Issuer) {
	issuer.Spec.CA = &certmanagerv1beta1.CAIssuer{
		SecretName: m.objects.selfSignedCA.Spec.SecretName,
	}
}

func (m *relayCoreManager) operatorWebhookCert(cert *certmanagerv1beta1.Certificate) {
	operatorServiceName := fmt.Sprintf("%s-operator", m.objects.relayCore.Name)

	cert.Spec.SecretName = fmt.Sprintf("%s-tls", cert.Name)
	cert.Spec.CommonName = fmt.Sprintf("%s.%s.svc", operatorServiceName, cert.Namespace)
	cert.Spec.DNSNames = append(cert.Spec.DNSNames,
		fmt.Sprintf("%s.%s.svc", operatorServiceName, cert.Namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", operatorServiceName, cert.Namespace),
		operatorServiceName,
	)
	cert.Spec.IssuerRef = certmanagermetav1.ObjectReference{
		Name: m.objects.issuer.Name,
	}
}

func (m *relayCoreManager) operatorStoragePVC(pvc *corev1.PersistentVolumeClaim) {
	pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}

	pvc.Spec.Resources = corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceStorage: resource.MustParse(relayOperatorStorageVolumeSize),
		},
	}
}

func (m *relayCoreManager) relayCore(rc *installerv1alpha1.RelayCore) {
	if m.logServiceOpts.Enabled {
		rc.Spec.LogService = &installerv1alpha1.LogServiceConfig{
			Image:           relayLogServiceImage,
			ImagePullPolicy: corev1.PullAlways,

			CredentialsSecretName: m.logServiceOpts.CredentialsSecretName,
			Project:               m.logServiceOpts.Project,
			Dataset:               m.logServiceOpts.Dataset,
			Table:                 m.logServiceOpts.Table,
		}
	}

	rc.Spec.Operator = &installerv1alpha1.OperatorConfig{
		Image:             relayOperatorImage,
		ImagePullPolicy:   corev1.PullAlways,
		Standalone:        true,
		LogStoragePVCName: &m.objects.pvc.Name,
		AdmissionWebhookServer: &installerv1alpha1.AdmissionWebhookServerConfig{
			TLSSecretName:      m.objects.operatorWebhookCert.Spec.SecretName,
			CABundleSecretName: &m.objects.selfSignedCA.Spec.SecretName,
		},
		GenerateJWTSigningKey: true,
	}

	rc.Spec.MetadataAPI = &installerv1alpha1.MetadataAPIConfig{
		Image:           relayMetadataAPIImage,
		ImagePullPolicy: corev1.PullAlways,
		VaultAuthRole:   "tenant",
		VaultAuthPath:   "auth/jwt-tenants",
	}
}

func (m *relayCoreManager) rbacDefinition(rd *rbacmanagerv1beta1.RBACDefinition) {
	delegateName := fmt.Sprintf("%s-delegate", rd.Name)

	rd.RBACBindings = []rbacmanagerv1beta1.RBACBinding{
		{
			Name: delegateName,
			Subjects: []rbacmanagerv1beta1.Subject{
				{
					Subject: rbacv1.Subject{
						Kind:      "ServiceAccount",
						Name:      m.objects.relayCore.Status.OperatorServiceAccount,
						Namespace: m.objects.relayCore.Namespace,
					},
				},
			},
			ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{},
			RoleBindings: []rbacmanagerv1beta1.RoleBinding{
				{
					ClusterRole: delegateName,
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"controller.relay.sh/tenant-workload": "true",
						},
					},
				},
			},
		},
	}
}

func (m *relayCoreManager) wait(ctx context.Context) error {
	err := retry.Wait(ctx, func(ctx context.Context) (bool, error) {
		key, err := client.ObjectKeyFromObject(&m.objects.relayCore)
		if err != nil {
			return retry.Done(err)
		}

		if err := m.cl.APIClient.Get(ctx, key, &m.objects.relayCore); err != nil {
			return retry.Repeat(err)
		}

		if m.objects.relayCore.Status.Status != installerv1alpha1.StatusCreated {
			return retry.Repeat(errors.New("waiting for relaycore to be created"))
		}

		return retry.Done(nil)
	})
	if err != nil {
		return err
	}

	return nil
}

func newRelayCoreManager(cl *cluster.Client, logServiceOpts LogServiceOptions) *relayCoreManager {
	return &relayCoreManager{
		cl:             cl,
		objects:        newRelayCoreObjects(),
		logServiceOpts: logServiceOpts,
	}
}
