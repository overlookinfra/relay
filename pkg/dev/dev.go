package dev

import (
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"time"

	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/puppetlabs/leg/timeutil/pkg/retry"
	"github.com/puppetlabs/leg/workdir"
	installerv1alpha1 "github.com/puppetlabs/relay-core/pkg/apis/install.relay.sh/v1alpha1"
	"github.com/puppetlabs/relay-core/pkg/operator/dependency"
	v1 "github.com/puppetlabs/relay-core/pkg/workflow/types/v1"
	"github.com/puppetlabs/relay/pkg/cluster"
	"github.com/puppetlabs/relay/pkg/dev/manifests"
	"github.com/puppetlabs/relay/pkg/model"
	helmchartv1 "github.com/rancher/helm-controller/pkg/apis/helm.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage/names"
	kubernetesscheme "k8s.io/client-go/kubernetes/scheme"
	cachingv1alpha1 "knative.dev/caching/pkg/apis/caching/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DefaultScheme = runtime.NewScheme()
	schemeBuilder = runtime.NewSchemeBuilder(
		kubernetesscheme.AddToScheme,
		metav1.AddMetaToScheme,
		rbacv1.AddToScheme,
		apiextensionsv1.AddToScheme,
		apiextensionsv1beta1.AddToScheme,
		dependency.AddToScheme,
		certmanagerv1beta1.AddToScheme,
		helmchartv1.AddToScheme,
		cachingv1alpha1.AddToScheme,
		installerv1alpha1.AddToScheme,
	)
	_ = schemeBuilder.AddToScheme(DefaultScheme)
)

const (
	defaultWorkflowName      = "relay-workflow"
	jwtSigningKeysSecretName = "relay-core-v1-operator-signing-keys"
)

type Config struct {
	WorkDir *workdir.WorkDir
}

type Manager struct {
	cm  cluster.Manager
	cl  *cluster.Client
	cfg Config
}

type InitializeOptions struct {
	ImageRegistryPort int
}

// FIXME Consider a better mechanism for specific service options
type LogServiceOptions struct {
	Enabled               bool
	CredentialsSecretName string
	Project               string
	Dataset               string
	Table                 string
}

func (m *Manager) WriteKubeconfig(ctx context.Context) error {
	return m.cm.WriteKubeconfig(ctx, filepath.Join(m.cfg.WorkDir.Path, "kubeconfig"))
}

func (m *Manager) Delete(ctx context.Context) error {
	// TODO fix hack: deletes the PVCs because dirs inside are often created as root
	// and we don't want relay running like that on the host to rm the data dir.
	nm := newNamespaceManager(m.cl)
	if err := nm.delete(ctx, systemNamespace); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	err := retry.Wait(ctx, func(ctx context.Context) (bool, error) {
		pvcs := &corev1.PersistentVolumeClaimList{}
		if err := m.cl.APIClient.List(ctx, pvcs, client.InNamespace(systemNamespace)); err != nil {
			return retry.Repeat(err)
		}

		if len(pvcs.Items) != 0 {
			return retry.Repeat(fmt.Errorf("waiting for pvcs to be deleted"))
		}

		return retry.Done(nil)
	})
	if err != nil {
		return err
	}

	if err := m.cm.Delete(ctx); err != nil {
		return err
	}

	if err := m.cfg.WorkDir.Cleanup(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) RunWorkflow(ctx context.Context, r io.ReadCloser, params map[string]string) (*model.WorkflowSummary, error) {
	vm := newVaultManager(m.cl, m.cfg)
	am := newAdminManager(m.cl, vm)

	decoder := v1.NewDocumentStreamingDecoder(r, &v1.YAMLDecoder{})

	wd, err := decoder.DecodeStream(ctx)
	if err != nil {
		return nil, err
	}

	name := wd.Name
	if name == "" {
		name = defaultWorkflowName
	}

	runID := names.SimpleNameGenerator.GenerateName(name + "-")
	if err != nil {
		return nil, err
	}

	if err := am.addConnectionForWorkflow(ctx, name); err != nil {
		return nil, err
	}

	runParams := v1.WorkflowRunParameters{}

	for k, v := range params {
		runParams[k] = &v1.WorkflowRunParameter{
			Value: v,
		}
	}

	mapper := v1.NewDefaultRunEngineMapper(
		v1.WithDomainIDRunOption(name),
		v1.WithNamespaceRunOption(name),
		v1.WithWorkflowNameRunOption(name),
		v1.WithWorkflowRunNameRunOption(runID),
		v1.WithVaultEngineMountRunOption("customers"),
		v1.WithRunParametersRunOption(runParams),
	)

	manifest, err := mapper.ToRuntimeObjectsManifest(wd)
	if err != nil {
		return nil, err
	}

	if err := m.cl.APIClient.Create(ctx, manifest.Namespace); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return nil, err
		}
	}

	if err := m.cl.APIClient.Create(ctx, manifest.WorkflowRun); err != nil {
		return nil, err
	}

	ws := &model.WorkflowSummary{
		WorkflowIdentifier: &model.WorkflowIdentifier{
			Name: name,
		},
		Description: wd.Description,
	}

	return ws, nil
}

func (m *Manager) SetWorkflowSecret(ctx context.Context, workflow, key, value string) error {
	vm := newVaultManager(m.cl, m.cfg)
	secret := map[string]string{
		path.Join("customers", "workflows", workflow, key): value,
	}

	return vm.writeSecrets(ctx, secret)
}

func (m *Manager) Initialize(ctx context.Context, opts InitializeOptions) error {
	// I introduced some race condition where the cluster hasn't fully setup
	// the object APIs or something, so when we try to create objects here, it
	// will blow up saying the API for that object type doesn't exist. If we
	// sleep for just a second, then we give it enough time to fully warm up or
	// something. I dunno...
	//
	// There's an option in k3d's cluster create that I set to wait for the
	// server, but I think there's something deeper happening inside kubernetes
	// (probably in the API server).
	<-time.After(time.Second * 5)

	nm := newNamespaceManager(m.cl)
	vm := newVaultManager(m.cl, m.cfg)
	am := newAdminManager(m.cl, vm)
	rm := newRegistryManager(m.cl)

	if err := nm.reconcile(ctx); err != nil {
		return err
	}

	if err := am.reconcile(ctx); err != nil {
		return err
	}

	if err := rm.reconcile(ctx); err != nil {
		return err
	}

	patchers := []objectPatcherFunc{
		nm.objectNamespacePatcher(systemNamespace),
		missingProtocolPatcher,
		registryLoadBalancerPortPatcher(opts.ImageRegistryPort),
	}

	// Apply manifests in ordered phases. Note that some managers
	// have weird dependencies on running services. For instance, you cannot
	// create or apply a ClusterIssuer unless the cert-manager webhook service
	// is Ready. This means we will just wait for all services across all created
	// namespaces to be ready before moving to the next phase of applying manifests.
	// TODO: dynamically generate the list as we process the manifests

	if err := m.processManifests(ctx, "/01-init", patchers, []string{"cert-manager", "relay-system"}); err != nil {
		return err
	}

	if err := vm.reconcile(ctx); err != nil {
		return err
	}

	return nil
}

func (m *Manager) InitializeRelayCore(ctx context.Context, lsOpts LogServiceOptions) error {
	// I introduced some race condition where the cluster hasn't fully setup
	// the object APIs or something, so when we try to create objects here, it
	// will blow up saying the API for that object type doesn't exist. If we
	// sleep for just a second, then we give it enough time to fully warm up or
	// something. I dunno...
	//
	// There's an option in k3d's cluster create that I set to wait for the
	// server, but I think there's something deeper happening inside kubernetes
	// (probably in the API server).
	<-time.After(time.Second * 5)

	// log := m.cfg.Dialog

	nm := newNamespaceManager(m.cl)
	vm := newVaultManager(m.cl, m.cfg)
	rim := newRelayInstallerManager(m.cl)
	rcm := newRelayCoreManager(m.cl, lsOpts)

	// Apply manifests in ordered phases. Note that some managers
	// have weird dependencies on running services. For instance, you cannot
	// create or apply a ClusterIssuer unless the cert-manager webhook service
	// is Ready. This means we will just wait for all services across all created
	// namespaces to be ready before moving to the next phase of applying manifests.
	// TODO: dynamically generate the list as we process the manifests

	patchers := []objectPatcherFunc{
		nm.objectNamespacePatcher(tektonPipelinesNamespace),
		missingProtocolPatcher,
	}

	if err := m.processManifests(ctx, "/03-tekton", patchers, nil); err != nil {
		return err
	}

	patchers = []objectPatcherFunc{
		nm.objectNamespacePatcher(knativeServingNamespace),
		missingProtocolPatcher,
	}

	if err := m.processManifests(ctx, "/04-knative", patchers, nil); err != nil {
		return err
	}

	if err := m.processManifests(ctx, "/05-relay", nil, nil); err != nil {
		return err
	}

	if err := rim.reconcile(ctx); err != nil {
		return err
	}

	if err := rcm.reconcile(ctx); err != nil {
		return err
	}

	if err := vm.addRelayCoreAccess(ctx, &rcm.objects.relayCore); err != nil {
		return err
	}

	patchers = []objectPatcherFunc{
		nm.objectNamespacePatcher(ambassadorNamespace),
		missingProtocolPatcher,
		ambassadorPatcher,
	}

	if err := m.processManifests(ctx, "/06-ambassador", patchers, nil); err != nil {
		return err
	}

	patchers = []objectPatcherFunc{
		nm.objectNamespacePatcher("default"),
	}

	if err := m.processManifests(ctx, "/07-hostpath", patchers, nil); err != nil {
		return err
	}

	return nil
}

func (m *Manager) processManifests(ctx context.Context, path string, patchers []objectPatcherFunc, initNamespaces []string) error {
	objects, err := m.parseAndLoadManifests(manifests.MustAssetListDir(path)...)
	if err != nil {
		return err
	}

	if err := m.applyAllWithPatchers(ctx, patchers, objects); err != nil {
		return err
	}

	for _, ns := range initNamespaces {
		if err := m.waitForServices(ctx, ns); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) StartRelayCore(ctx context.Context) error {
	// same issue where as above in the initialization.
	<-time.After(time.Second * 5)

	vm := newVaultManager(m.cl, m.cfg)
	rm := newRegistryManager(m.cl)

	if err := vm.reconcile(ctx); err != nil {
		return err
	}

	if err := rm.reconcile(ctx); err != nil {
		return err
	}

	return m.waitForServices(ctx, systemNamespace)
}

func (m *Manager) parseAndLoadManifests(files ...string) ([]runtime.Object, error) {
	objects := []runtime.Object{}

	for _, f := range files {
		manifest := manifests.MustAsset(f)

		manifestObjects, err := parseManifest(manifest)
		if err != nil {
			return nil, err
		}

		objects = append(objects, manifestObjects...)
	}

	return objects, nil
}

func (m *Manager) waitForServices(ctx context.Context, namespace string) error {
	err := retry.Wait(ctx, func(ctx context.Context) (bool, error) {
		eps := &corev1.EndpointsList{}
		if err := m.cl.APIClient.List(ctx, eps, client.InNamespace(namespace)); err != nil {
			return retry.Repeat(err)
		}

		if len(eps.Items) == 0 {
			return retry.Repeat(fmt.Errorf("waiting for endpoints"))
		}

		for _, ep := range eps.Items {
			if len(ep.Subsets) == 0 {
				return retry.Repeat(fmt.Errorf("waiting for subsets"))
			}

			for _, subset := range ep.Subsets {
				if len(subset.Addresses) == 0 {
					return retry.Repeat(fmt.Errorf("waiting for pod assignment"))
				}
			}
		}

		return retry.Done(nil)
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) waitForCertificates(ctx context.Context, namespace string) error {
	err := retry.Wait(ctx, func(ctx context.Context) (bool, error) {
		certs := &certmanagerv1beta1.CertificateList{}
		if err := m.cl.APIClient.List(ctx, certs, client.InNamespace(namespace)); err != nil {
			return retry.Repeat(err)
		}

		if len(certs.Items) == 0 {
			return retry.Repeat(fmt.Errorf("waiting for certificates"))
		}

		for _, cert := range certs.Items {
			for _, cond := range cert.Status.Conditions {
				if cond.Type == certmanagerv1beta1.CertificateConditionReady {
					if cond.Status != certmanagermetav1.ConditionTrue {
						return retry.Repeat(fmt.Errorf("waiting for certificates to be ready"))
					}
				}
			}
		}

		return retry.Done(nil)
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) apply(ctx context.Context, obj runtime.Object) error {
	if err := m.cl.APIClient.Patch(ctx, obj, client.Apply, client.ForceOwnership, client.FieldOwner("relay-e2e")); err != nil {
		return fmt.Errorf("failed to apply object '%s': %w", obj.GetObjectKind().GroupVersionKind().String(), err)
	}

	return nil
}

func (m *Manager) applyAllWithPatchers(ctx context.Context, patchers []objectPatcherFunc, objs []runtime.Object) error {
	for _, obj := range objs {
		for _, patcher := range patchers {
			patcher(obj)
		}

		if err := m.apply(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func NewManager(cm cluster.Manager, cl *cluster.Client, cfg Config) *Manager {
	return &Manager{
		cm:  cm,
		cl:  cl,
		cfg: cfg,
	}
}
