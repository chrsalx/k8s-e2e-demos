package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/chrsalx/k8s-e2e-demos/pkg/argo"
	"github.com/chrsalx/k8s-e2e-demos/pkg/objectfromfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

var currentDir, _ = os.Getwd()

func TestNginxAppWithArgo(t *testing.T) {
	feature := features.
		New("Nginx server").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			err := argo.AddResourcesToScheme(config)
			require.NoError(t, err)

			helmMgr := helm.New(config.KubeconfigFile())
			argoClient := argo.NewResourceManager(config)

			err = helmMgr.RunRepo(helm.WithArgs(
				"add",
				"argo",
				"https://argoproj.github.io/argo-helm",
			))

			err = helmMgr.RunInstall(
				helm.WithName("argo-cd"),
				helm.WithNamespace(argocdNamespace),
				helm.WithReleaseName("argo/argo-cd"),
				helm.WithVersion("5.34.1"),
			)
			require.NoError(t, err)

			argoNginxAppSpec, err := os.ReadFile(filepath.Join(currentDir, "..", "argo-apps", "nginx", "app.yaml"))
			require.NoError(t, err)

			nginxApp, err := objectfromfile.GetArgoApplicationFromYAML(argoNginxAppSpec)
			require.NoError(t, err)

			err = argoClient.CreateApplicationWithContext(ctx, nginxApp)
			require.NoError(t, err)

			return ctx
		}).
		Assess(
			"Testing the app syncs",
			func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
				app := &applicationV1Alpha1.Application{ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx-server",
					Namespace: argocdNamespace,
				}}

				var isAppHealthyAndSynced = func(object k8s.Object) bool {
					argoApp := object.(*applicationV1Alpha1.Application)

					return string(argoApp.Status.Health.Status) == "Healthy" &&
						string(argoApp.Status.Sync.Status) == "Synced"
				}

				err := wait.For(
					conditions.New(config.Client().Resources()).ResourceMatch(app, isAppHealthyAndSynced),
					wait.WithTimeout(time.Minute*5),
				)
				assert.NoError(t, err, "Error waiting for ArgoCD app to sync")

				return ctx
			}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			helmClient := helm.New(config.KubeconfigFile())
			err := helmClient.RunRepo(helm.WithArgs("remove", "argo"))
			require.NoError(t, err)

			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
