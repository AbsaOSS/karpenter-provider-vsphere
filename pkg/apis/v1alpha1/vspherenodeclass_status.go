package v1alpha1

import (
	"fmt"

	"github.com/awslabs/operatorpkg/status"
	corev1 "k8s.io/api/core/v1"
)

type VsphereNodeClassStatus struct {
	// ImageTemplateID is the resolved vSphere inventory path used to clone virtual machines.
	// +optional
	ImageTemplateID string `json:"imageTemplateID,omitempty"`
	// ResourcePool is the resolved vSphere ResourcePool. Kind holds the managed object type and Name the managed object ID.
	// +optional
	ResourcePool *corev1.ObjectReference `json:"resourcePool,omitempty"`
	// Datastore is the resolved vSphere Datastore. Kind holds the managed object type and Name the managed object ID.
	// +optional
	Datastore *corev1.ObjectReference `json:"datastore,omitempty"`
	// Network is the resolved vSphere Network. Kind holds the managed object type and Name the managed object ID.
	// +optional
	Network *corev1.ObjectReference `json:"network,omitempty"`
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	// +optional
	Conditions []status.Condition `json:"conditions,omitempty"`
}

func (nc *VsphereNodeClass) StatusConditions(...status.ForOption) status.ConditionSet {
	conds := []string{
		ConditionTypeKubernetesVersionReady,
		ConditionTypeImageReady,
		ConditionTypeResourcePoolReady,
		ConditionTypeDatastoreReady,
		ConditionTypeNetworkReady,
	}
	return status.NewReadyConditions(conds...).For(nc)
}

func (nc *VsphereNodeClass) GetConditions() []status.Condition {
	return nc.Status.Conditions
}

func (nc *VsphereNodeClass) SetConditions(conditions []status.Condition) {
	nc.Status.Conditions = conditions
}

func (nc *VsphereNodeClass) GetImageTemplateID() string {
	return nc.Status.ImageTemplateID
}

func (nc *VsphereNodeClass) GetResourcePool() (*corev1.ObjectReference, error) {
	if nc.Status.ResourcePool == nil || nc.Status.ResourcePool.Name == "" {
		return nil, fmt.Errorf("NodeClass ResourcePool is not resolved")
	}
	return nc.Status.ResourcePool, nil
}

func (nc *VsphereNodeClass) GetDatastore() (*corev1.ObjectReference, error) {
	if nc.Status.Datastore == nil || nc.Status.Datastore.Name == "" {
		return nil, fmt.Errorf("NodeClass Datastore is not resolved")
	}
	return nc.Status.Datastore, nil
}

func (nc *VsphereNodeClass) GetNetwork() (*corev1.ObjectReference, error) {
	if nc.Status.Network == nil || nc.Status.Network.Name == "" {
		return nil, fmt.Errorf("NodeClass Network is not resolved")
	}
	return nc.Status.Network, nil
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
