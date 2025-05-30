package v1alpha1

import (
	"fmt"

	"github.com/awslabs/operatorpkg/status"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeKubernetesVersionReady = "KubernetesVersionReady"
)

// VsphereNodeClass is the Schema for the VsphereNodeClass API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vspherenodeclasses,scope=Cluster,categories=karpenter,shortName={vspherenc,vspherencs}
// +kubebuilder:subresource:status
type VsphereNodeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VsphereNodeClassSpec   `json:"spec,omitempty"`
	Status            VsphereNodeClassStatus `json:"status,omitempty"`
}

// VsphereNodeClassList contains a list of VsphereNodeClasses
// +kubebuilder:object:root=true
type VsphereNodeClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VsphereNodeClass `json:"items"`
}

type VsphereNodeClassSpec struct {
	ComputeCluster string                  `json:"computeCluster,omitempty"`
	Datastore      string                  `json:"datastore,omitempty"`
	Path           string                  `json:"path,omitempty"`
	Image          string                  `json:"image,omitempty"`
	Network        string                  `json:"network,omitempty"`
	InstanceTypes  map[string]InstanceType `json:"instanceTypes,omitempty"`
}

type InstanceType struct {
	CPU     string `json:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty"`
	MaxPods string `json:"maxPods,omitempty"`
	Storage string `json:"storage,omitempty"`
	OS      string `json:"os,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

type VsphereNodeClassStatus struct {
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	// +optional
	Conditions []status.Condition `json:"conditions,omitempty"`
}

func (nc *VsphereNodeClass) StatusConditions() status.ConditionSet {
	conds := []string{
		ConditionTypeKubernetesVersionReady,
	}
	return status.NewReadyConditions(conds...).For(nc)
}

func (nc *VsphereNodeClass) GetConditions() []status.Condition {
	return nc.Status.Conditions
}

func (nc *VsphereNodeClass) SetConditions(conditions []status.Condition) {
	nc.Status.Conditions = conditions
}

const VsphereNodeClassHashVersion = "v1"

func (nc *VsphereNodeClass) Hash() string {
	return fmt.Sprint(lo.Must(hashstructure.Hash(nc.Spec, hashstructure.FormatV2, &hashstructure.HashOptions{
		SlicesAsSets:    true,
		IgnoreZeroValue: true,
		ZeroNil:         true,
	})))
}

func (nc *VsphereNodeClass) GetKubernetesVersion() (string, error) {
	err := nc.validateKubernetesVersionReadiness()
	if err != nil {
		return "", err
	}
	return nc.Status.KubernetesVersion, nil
}

func (nc *VsphereNodeClass) validateKubernetesVersionReadiness() error {
	if nc == nil {
		return fmt.Errorf("NodeClass is nil, condition %s is not true", ConditionTypeKubernetesVersionReady)
	}
	kubernetesVersionCondition := nc.StatusConditions().Get(ConditionTypeKubernetesVersionReady)
	if kubernetesVersionCondition.IsFalse() || kubernetesVersionCondition.IsUnknown() {
		return fmt.Errorf("NodeClass condition %s, is in Ready=%s, %s", ConditionTypeKubernetesVersionReady, kubernetesVersionCondition.GetStatus(), kubernetesVersionCondition.Message)
	} else if kubernetesVersionCondition.ObservedGeneration != nc.GetGeneration() {
		return fmt.Errorf("NodeClass condition %s ObservedGeneration %d does not match the NodeClass Generation %d", ConditionTypeKubernetesVersionReady, kubernetesVersionCondition.ObservedGeneration, nc.GetGeneration())
	} else if nc.Status.KubernetesVersion == "" {
		return fmt.Errorf("NodeClass KubernetesVersion is uninitialized")
	}
	return nil
}
