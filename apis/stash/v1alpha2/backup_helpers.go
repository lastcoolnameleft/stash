package v1alpha2

import (
	"hash/fnv"
	"strconv"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/stash/apis"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
)

func (b Backup) GetSpecHash() string {
	hash := fnv.New64a()
	hashutil.DeepHashObject(hash, b.Spec)
	return strconv.FormatUint(hash.Sum64(), 10)
}

func (bkp Backup) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralBackup,
		Singular:      ResourceSingularBackup,
		Kind:          ResourceKindBackup,
		ShortNames:    []string{"bkp"},
		Categories:    []string{"storage", "appscode"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "stash"},
		},
		SpecDefinitionName:      "github.com/appscode/stash/apis/stash/v1alpha2.Backup",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: apis.EnableStatusSubresource,
		AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
			{
				Name:     "Agent",
				Type:     "string",
				JSONPath: ".spec.backupAgent",
			},
			{
				Name:     "Schedule",
				Type:     "string",
				JSONPath: ".spec.schedule",
			},
			{
				Name:     "Backup-Type",
				Type:     "string",
				JSONPath: ".spec.type",
				Priority: 10,
			},
			{
				Name:     "Paused",
				Type:     "boolean",
				JSONPath: ".spec.paused",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
		},
	})
}
