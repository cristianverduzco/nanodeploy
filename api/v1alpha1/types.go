package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceType defines the type of managed service
// +kubebuilder:validation:Enum=postgresql;redis;rabbitmq
type ServiceType string

const (
	ServiceTypePostgresql ServiceType = "postgresql"
	ServiceTypeRedis      ServiceType = "redis"
	ServiceTypeRabbitMQ   ServiceType = "rabbitmq"
)

// ServicePhase represents the current lifecycle phase of a ManagedService
type ServicePhase string

const (
	ServicePhasePending     ServicePhase = "Pending"
	ServicePhaseProvisioning ServicePhase = "Provisioning"
	ServicePhaseReady       ServicePhase = "Ready"
	ServicePhaseFailed      ServicePhase = "Failed"
	ServicePhaseTerminating ServicePhase = "Terminating"
)

// ManagedServiceSpec defines the desired state of a ManagedService
type ManagedServiceSpec struct {
	// Type is the kind of service to provision (e.g. postgresql, redis)
	Type ServiceType `json:"type"`

	// Version is the service version to deploy (e.g. "15" for PostgreSQL 15)
	Version string `json:"version"`

	// Replicas is the number of instances to run
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`

	// StorageGB is the amount of persistent storage to allocate in gigabytes
	// +kubebuilder:default=5
	StorageGB int32 `json:"storageGB,omitempty"`

	// DatabaseName is the initial database to create (postgresql only)
	// +optional
	DatabaseName string `json:"databaseName,omitempty"`
}

// ManagedServiceStatus defines the observed state of a ManagedService
type ManagedServiceStatus struct {
	// Phase is the current lifecycle phase
	Phase ServicePhase `json:"phase,omitempty"`

	// Message is a human-readable status message
	Message string `json:"message,omitempty"`

	// ReadyReplicas is how many replicas are currently ready
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Endpoint is the in-cluster DNS address to reach the service
	Endpoint string `json:"endpoint,omitempty"`

	// LastUpdated is the last time the status was updated
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// ManagedService is the Schema for the managedservices API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.endpoint"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServiceSpec   `json:"spec,omitempty"`
	Status ManagedServiceStatus `json:"status,omitempty"`
}

// ManagedServiceList contains a list of ManagedService
// +kubebuilder:object:root=true
type ManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedService `json:"items"`
}