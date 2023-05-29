package argo

import (
	"context"

	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type ResourceManager struct {
	k8sConfig *envconf.Config
}

func (r *ResourceManager) GetApplicationWithContext(
	ctx context.Context,
	name string,
	namespace string,
) (*applicationV1Alpha1.Application, error) {
	app := &applicationV1Alpha1.Application{}
	err := r.k8sConfig.Client().Resources().Get(ctx, name, namespace, app)
	if err != nil {
		return &applicationV1Alpha1.Application{}, err
	}
	return app, nil
}

func (r *ResourceManager) CreateApplicationWithContext(
	ctx context.Context,
	obj *applicationV1Alpha1.Application,
) error {
	return r.k8sConfig.Client().Resources().Create(ctx, obj)
}

func NewResourceManager(config *envconf.Config) *ResourceManager {
	return &ResourceManager{k8sConfig: config}
}

func AddResourcesToScheme(config *envconf.Config) error {
	scheme := config.Client().Resources().GetScheme()
	return applicationV1Alpha1.AddToScheme(scheme)
}
