package test

import (
	"context"
	"fmt"
	"github.com/chrsalx/k8s-e2e-demos/pkg/argo"
	"github.com/chrsalx/k8s-e2e-demos/pkg/objectfromfile"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

var currentDir, _ = os.Getwd()

var (
	destinationNamespace string
	releaseName          string
)

func TestNginxAppWithHelm(t *testing.T) {
	feature := features.
		New("Testing nginx helm chart no ArgoCD sync").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			err := argo.AddResourcesToScheme(config)
			require.NoError(t, err)

			helmMgr := helm.New(config.KubeconfigFile())

			argoNginxAppSpec, err := os.ReadFile(filepath.Join(currentDir, "..", "app", "nginx", "app.yaml"))
			require.NoError(t, err)

			nginxApp, err := objectfromfile.GetArgoApplicationFromYAML(argoNginxAppSpec)
			require.NoError(t, err)

			helmRepoURL := nginxApp.Spec.Source.RepoURL
			helmRepoName := path.Base(helmRepoURL) // "https://charts.bitnami.com/bitnami" => "bitnami"

			err = helmMgr.RunRepo(helm.WithArgs(
				"add",
				helmRepoName,
				helmRepoURL,
			))
			require.NoError(t, err)

			destinationNamespace = nginxApp.Spec.Destination.Namespace
			helmChartName := nginxApp.Spec.Source.Chart
			releaseName = helmChartName
			helmChartVersion := nginxApp.Spec.Source.TargetRevision
			fullChartName := fmt.Sprintf("%s/%s", helmRepoName, helmChartName)
			err = helmMgr.RunInstall(
				helm.WithName(releaseName),
				helm.WithNamespace(destinationNamespace),
				helm.WithChart(fullChartName),
				helm.WithVersion(helmChartVersion),
			)
			require.NoError(t, err)

			return ctx
		}).
		Assess(
			"Testing the chart is installed correctly",
			func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: releaseName, Namespace: destinationNamespace},
				}

				var isDeploymentFullyRunning = func(object k8s.Object) bool {
					dep := object.(*appsv1.Deployment)

					return dep.Status.AvailableReplicas == dep.Status.ReadyReplicas
				}

				err := wait.For(
					conditions.New(config.Client().Resources()).ResourceMatch(deployment, isDeploymentFullyRunning),
					wait.WithTimeout(time.Minute*5),
				)
				assert.NoError(t, err, "Error waiting nginx-server to start")

				return ctx
			}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
