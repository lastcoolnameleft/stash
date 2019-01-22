package v1alpha2

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	store "kmodules.xyz/objectstore-api/api/v1"
)

const (
	ResourceKindBackupConfiguration     = "BackupConfiguration"
	ResourceSingularBackupConfiguration = "backupConfiguration"
	ResourcePluralBackupConfiguration   = "backupConfigurations"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupConfiguration struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupConfigurationSpec `json:"spec,omitempty"`
}

type BackupConfigurationSpec struct {
	Schedule string `json:"schedule,omitempty"`
	// BackupTemplate specify the BackupTemplate crd that will be used for creating backup job
	// +optional
	BackupTemplate string `json:"backupTemplate,omitempty"`
	// Repository refer to the Repository crd that holds backend information
	Repository core.LocalObjectReference `json:"repository"`
	// TargetRef specify the backup target
	// +optional
	*TargetRef `json:"targetRef,omitempty"`
	// TargetDirectories specify the directories to backup when the target is a volume
	//+optional
	TargetDirectories []string `json:"targetDirectories,omitempty"`
	// RetentionPolicy indicates the policy to follow to clean old backup snapshots
	RetentionPolicy `json:"retentionPolicy,omitempty"`
	// Indicates that the BackupConfiguration is paused from taking backup. Default value is 'false'
	// +optional
	Paused bool `json:"paused,omitempty"`
	// ContainerAttributes allow to specify Resources, SecurityContext, ReadinessProbe etc. for backup sidecar or job's container
	//+optional
	*ContainerAttributes `json:"containerAttributes,omitempty"`
	// PodAttributes allow to specify NodeSelector, Affinity, Toleration etc. for backup job's pod
	//+optional
	*PodAttributes `json:"podAttributes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupConfiguration `json:"items,omitempty"`
}

type RetentionStrategy string

const (
	KeepLast    RetentionStrategy = "--keep-last"
	KeepHourly  RetentionStrategy = "--keep-hourly"
	KeepDaily   RetentionStrategy = "--keep-daily"
	KeepWeekly  RetentionStrategy = "--keep-weekly"
	KeepMonthly RetentionStrategy = "--keep-monthly"
	KeepYearly  RetentionStrategy = "--keep-yearly"
	KeepTag     RetentionStrategy = "--keep-tag"
)

type RetentionPolicy struct {
	Name        string   `json:"name,omitempty"`
	KeepLast    int      `json:"keepLast,omitempty"`
	KeepHourly  int      `json:"keepHourly,omitempty"`
	KeepDaily   int      `json:"keepDaily,omitempty"`
	KeepWeekly  int      `json:"keepWeekly,omitempty"`
	KeepMonthly int      `json:"keepMonthly,omitempty"`
	KeepYearly  int      `json:"keepYearly,omitempty"`
	KeepTags    []string `json:"keepTags,omitempty"`
	Prune       bool     `json:"prune,omitempty"`
	DryRun      bool     `json:"dryRun,omitempty"`
}

type TargetRef struct {
	// Volume specifies the target volume to backup
	//+optional
	Volume *store.LocalSpec `json:"volume,omitempty"`
	// WorkloadRef refers to the workload to backup
	// +optional
	Workload *WorkloadRef `json:"workload,omitempty"`
}

type WorkloadRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
}
