package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindBackupTemplate     = "BackupTemplate"
	ResourcePluralBackupTemplate   = "backupTemplates"
	ResourceSingularBackupTemplate = "backupTemplate"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupTemplateSpec `json:"spec,omitempty"`
}

type BackupTemplateSpec struct {
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

type BackupTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupTemplate `json:"items,omitempty"`
}
