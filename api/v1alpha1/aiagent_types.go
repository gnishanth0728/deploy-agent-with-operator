package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AgentModel struct {
	Provider     string `json:"provider"`
	Name         string `json:"name"`
	APIKeySecret string `json:"apiKeySecret,omitempty"`
}

type AgentConfig struct {
	Role         string   `json:"role,omitempty"`
	SystemPrompt string   `json:"systemPrompt,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type AgentMemory struct {
	Enabled bool   `json:"enabled,omitempty"`
	Backend string `json:"backend,omitempty"`
}

type AIAgentSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`

	Image string `json:"image"`

	Model AgentModel `json:"model"`

	Agent AgentConfig `json:"agent,omitempty"`

	Memory AgentMemory `json:"memory,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type AIAgentStatus struct {
	Phase string `json:"phase,omitempty"`

	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	Message string `json:"message,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas"
type AIAgent struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AIAgentSpec `json:"spec"`

	Status AIAgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type AIAgentList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AIAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(addKnownTypes)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		GroupVersion,
		&AIAgent{},
		&AIAgentList{},
	)

	metav1.AddToGroupVersion(scheme, GroupVersion)

	return nil
}
