// Package instancefixture provides test builders for instance.Instance.
//
// It is kept separate from internal/testutil because it depends on
// pkg/providers/instance; merging it into testutil would create an import
// cycle for any test file inside that package (e.g. instance_test.go).
package instancefixture

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// StandardLaunchTime returns a standard test launch time.
func StandardLaunchTime() time.Time {
	return time.Date(2026, time.January, 10, 12, 30, 0, 0, time.UTC)
}

// Builder builds instance.Instance test fixtures with fluent API.
type Builder struct {
	obj *instance.Instance
}

// New creates a new Instance builder with sensible defaults.
func New() *Builder {
	return &Builder{
		obj: &instance.Instance{
			ID:         "vm-1234",
			Name:       "worker-1234",
			Image:      "ubuntu-2404",
			State:      "poweredOn",
			LaunchTime: StandardLaunchTime(),
			Tags: map[string]string{
				v1alpha1.ClusterNameTagKey: "test-cluster",
				corev1.LabelTopologyZone:   "zone-a",
				karpv1.NodePoolLabelKey:    "default-nodepool",
			},
		},
	}
}

// WithID sets the instance ID.
func (b *Builder) WithID(id string) *Builder {
	b.obj.ID = id
	return b
}

// WithName sets the instance name.
func (b *Builder) WithName(name string) *Builder {
	b.obj.Name = name
	return b
}

// WithImage sets the instance image.
func (b *Builder) WithImage(image string) *Builder {
	b.obj.Image = image
	return b
}

// WithState sets the instance state (poweredOn, powerOff, etc.).
func (b *Builder) WithState(state string) *Builder {
	b.obj.State = state
	return b
}

// WithLaunchTime sets the instance launch time.
func (b *Builder) WithLaunchTime(launchTime time.Time) *Builder {
	b.obj.LaunchTime = launchTime
	return b
}

// WithTag adds a tag to the instance.
func (b *Builder) WithTag(key, value string) *Builder {
	if b.obj.Tags == nil {
		b.obj.Tags = make(map[string]string)
	}
	b.obj.Tags[key] = value
	return b
}

// WithClusterName sets the cluster name tag.
func (b *Builder) WithClusterName(clusterName string) *Builder {
	return b.WithTag(v1alpha1.ClusterNameTagKey, clusterName)
}

// WithZone sets the topology zone tag.
func (b *Builder) WithZone(zone string) *Builder {
	return b.WithTag(corev1.LabelTopologyZone, zone)
}

// WithNodePool sets the node pool label tag.
func (b *Builder) WithNodePool(nodePool string) *Builder {
	return b.WithTag(karpv1.NodePoolLabelKey, nodePool)
}

// Build returns the constructed Instance.
func (b *Builder) Build() *instance.Instance {
	return b.obj
}

// PoweredOnInstance creates a default powered-on instance for testing.
func PoweredOnInstance() *instance.Instance {
	return New().Build()
}

// PoweredOffInstance creates a default powered-off instance for testing.
func PoweredOffInstance() *instance.Instance {
	return New().
		WithID("vm-powered-off").
		WithName("worker-powered-off").
		WithState("powerOff").
		Build()
}
