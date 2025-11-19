package assets

import (
	"bytes"
	"embed"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	//go:embed manifests/*
	manifests  embed.FS
	appsScheme = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsScheme)
)

func init() {
	if err := appsv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

type DeploymentParams struct {
	Replicas int32
}

func RenderDeployment(params DeploymentParams) *appsv1.Deployment {
	tpl := template.Must(template.New("nginx").ParseFS(manifests, "manifests/nginx_deployment.yaml"))
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, params); err != nil {
		panic(err)
	}
	obj, err := runtime.Decode(
		appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion),
		buf.Bytes(),
	)
	if err != nil {
		panic(err)
	}
	return obj.(*appsv1.Deployment)
}
func GetDeploymentFromFile(name string) *appsv1.Deployment {
	deploymentBytes, err := manifests.ReadFile(name)
	if err != nil {
		panic(err)
	}
	deploymentObject, err := runtime.Decode(
		appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion),
		deploymentBytes,
	)
	if err != nil {
		panic(err)
	}
	return deploymentObject.(*appsv1.Deployment)
}
