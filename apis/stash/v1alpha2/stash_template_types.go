package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindStashTemplate     = "StashTemplate"
	ResourcePluralStashTemplate   = "stashtemplates"
	ResourceSingularStashTemplate = "stashtemplate"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type StashTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StashTemplateSpec `json:"spec,omitempty"`
}

type StashTemplateSpec struct {
	Actions []Steps `json:"actions,omitempty"`
}

type Steps struct {
	// Name indicates the name of Action crd
	Name string `json:"name,omitempty"`
	// Inputs specifies the inputs of respective Action
	// +optional
	Inputs map[string]string `json:"inputs,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type StashTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StashTemplate `json:"items,omitempty"`
}
