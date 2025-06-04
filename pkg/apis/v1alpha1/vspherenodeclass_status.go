package v1alpha1

import (
	"fmt"

	"github.com/awslabs/operatorpkg/status"
)

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
