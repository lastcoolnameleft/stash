package v1alpha2

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindBackup     = "Backup"
	ResourceSingularBackup = "backup"
	ResourcePluralBackup   = "backups"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Backup struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupSpec `json:"spec,omitempty"`
}

type BackupSpec struct {
	Schedule string `json:"schedule,omitempty"`
	// BackupAgent specify the AgentTemplate that will be used for backup sidecar or job
	BackupAgent core.LocalObjectReference `json:"backupAgent,omitempty"`
	// Repository refer to the Repository crd that hold backend information
	Repository core.LocalObjectReference `json:"repository"`
	// TargetRef specify the backup target
	TargetRef WorkloadRef `json:"targetRef"`
	// TargetDirectories specify the directories to backup when the target is a volume
	//+optional
	TargetDirectories []string `json:"targetDirectories,omitempty"`
	// RetentionPolicy indicates the policy to follow to clean old backup snapshots
	RetentionPolicy `json:"retentionPolicy,omitempty"`
	// Indicates that the Backup is paused from taking backup. Default value is 'false'
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

type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items,omitempty"`
}

type BackupType string

const (
	ScheduledBackup BackupType = "Scheduled" // default, backup using sidecar or cron job
	OneTimeBackup   BackupType = "OneTime"   // backup using init container or job
)

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

type PodAttributes struct {
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements.
	// +optional
	NodeName string `json:"nodeName,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *core.PodSecurityContext `json:"securityContext,omitempty"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *core.Affinity `json:"affinity,omitempty"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []core.Toleration `json:"tolerations,omitempty"`
	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// The priority value. Various system components use this field to find the
	// priority of the pod. When Priority Admission Controller is enabled, it
	// prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// +optional
	Priority *int32 `json:"priority,omitempty"`
}

type ContainerAttributes struct {
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	// Security options the pod should run with.
	// More info: https://kubernetes.io/docs/concepts/policy/security-context/
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	// +optional
	SecurityContext *core.SecurityContext `json:"securityContext,omitempty"`
	// Periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	LivenessProbe *core.Probe `json:"livenessProbe,omitempty"`
	// Periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	ReadinessProbe *core.Probe `json:"readinessProbe,omitempty"`
	// Actions that the management system should take in response to container lifecycle events.
	// Cannot be updated.
	// +optional
	Lifecycle *core.Lifecycle `json:"lifecycle,omitempty"`
}

type WorkloadRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
}
