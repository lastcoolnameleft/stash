package v1alpha2

import (
	"github.com/appscode/go/encoding/json/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindBackupInstance     = "BackupInstance"
	ResourcePluralBackupInstance   = "backupinstances"
	ResourceSingularBackupInstance = "backupinstance"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupInstanceSpec   `json:"spec,omitempty"`
	Status            BackupInstanceStatus `json:"status,omitempty"`
}

type BackupInstanceSpec struct {
	// TargetBackupConfiguration indicates the respective BackupConfiguration crd for target backup
	TargetBackupConfiguration string `json:"targetBackupConfiguration"`
}

type BackupInstancePhase string

const (
	BackupInstancePending   BackupInstancePhase = "Pending"
	BackupInstanceRunning   BackupInstancePhase = "Running"
	BackupInstanceSucceeded BackupInstancePhase = "Succeeded"
	BackupInstanceFailed    BackupInstancePhase = "Failed"
	BackupInstanceUnknown   BackupInstancePhase = "Unknown"
)

type BackupInstanceStatus struct {
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *types.IntHash      `json:"observedGeneration,omitempty"`
	Phase              BackupInstancePhase `json:"phase,omitempty"`
	Stats              []BackupStats       `json:"stats,omitempty"`
}

type BackupStats struct {
	Snapshot  string    `json:"snapshot,omitempty"`
	Size      string    `json:"size,omitempty"`
	Uploaded  string    `json:"uploaded,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	FileStats FileStats `json:"fileStats,omitempty"`
}

type FileStats struct {
	New        int `json:"new,omitempty"`
	Changed    int `json:"changed,omitempty"`
	Unmodified int `json:"unmodified,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupInstance `json:"items,omitempty"`
}
