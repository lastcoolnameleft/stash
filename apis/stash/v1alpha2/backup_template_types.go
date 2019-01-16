package v1alpha2

import (
	core "k8s.io/api/core/v1"
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
	// RepositorySpec is used to create Repository crd for respective workload
	RepositorySpec  `json:",inline"`
	Schedule        string                    `json:"schedule,omitempty"`
	BackupAgent     core.LocalObjectReference `json:"backupAgent,omitempty"`
	RetentionPolicy `json:"retentionPolicy,omitempty"`
	// ContainerAttributes allow to specify Resources, SecurityContext, ReadinessProbe etc. for backup sidecar or job's container
	//+optional
	*ContainerAttributes `json:"containerAttributes,omitempty"`
	// PodAttributes allow to specify NodeSelector, Affinity, Toleration etc. for backup job's pod
	//+optional
	*PodAttributes `json:"podAttributes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupTemplate `json:"items,omitempty"`
}
