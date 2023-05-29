package test

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envfuncs"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	testEnv         env.Environment
	kindClusterName string
	argocdNamespace = "argocd"
	nginxNamespace  = "nginx"
)

func TestMain(m *testing.M) {
	config, err := envconf.NewFromFlags()

	if err != nil {
		fmt.Println("Could not create config from env", err)
	}

	testEnv = env.NewWithConfig(config)
	kindClusterName = envconf.RandomName("argo-helm-nginx", 20)

	testEnv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.CreateNamespace(argocdNamespace),
		envfuncs.CreateNamespace(nginxNamespace),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(argocdNamespace),
		envfuncs.DeleteNamespace(nginxNamespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)
	os.Exit(testEnv.Run(m))
}
