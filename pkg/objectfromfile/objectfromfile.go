package objectfromfile

import (
	applicationV1Alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"sigs.k8s.io/yaml"
)

func GetArgoApplicationFromYAML(fileData []byte) (*applicationV1Alpha1.Application, error) {
	app := &applicationV1Alpha1.Application{}
	jsonData, err := yaml.YAMLToJSON(fileData)
	if err != nil {
		return &applicationV1Alpha1.Application{}, err
	}

	err = yaml.Unmarshal(jsonData, app)
	if err != nil {
		return &applicationV1Alpha1.Application{}, err
	}

	return app, nil
}
