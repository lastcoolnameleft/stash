package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDefaultBackupConfiguration     = "DefaultBackupConfiguration"
	ResourcePluralDefaultBackupConfiguration   = "defaultBackupConfigurations"
	ResourceSingularDefaultBackupConfiguration = "defaultBackupConfiguration"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DefaultBackupConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DefaultBackupConfigurationSpec `json:"spec,omitempty"`
}

type DefaultBackupConfigurationSpec struct {
	// RepositorySpec is used to create Repository crd for respective workload
	RepositorySpec `json:",inline"`
	Schedule       string `json:"schedule,omitempty"`
	// BackupTemplate specify the BackupTemplate crd that will be used for creating backup job
	// +optional
	BackupTemplate string `json:"backupTemplate,omitempty"`
	// RetentionPolicy indicates the policy to follow to clean old backup snapshots
	RetentionPolicy `json:"retentionPolicy,omitempty"`
	// ContainerAttributes allow to specify Resources, SecurityContext, ReadinessProbe etc. for backup sidecar or job's container
	//+optional
	*ContainerAttributes `json:"containerAttributes,omitempty"`
	// PodAttributes allow to specify NodeSelector, Affinity, Toleration etc. for backup job's pod
	//+optional
	*PodAttributes `json:"podAttributes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DefaultBackupConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DefaultBackupConfiguration `json:"items,omitempty"`
}
