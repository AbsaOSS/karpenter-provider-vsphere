package termination

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/karpenter/pkg/events"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
)

func PrettySlice[T any](s []T, maxItems int) string {
	var sb strings.Builder
	for i, elem := range s {
		if i > maxItems-1 {
			fmt.Fprintf(&sb, " and %d other(s)", len(s)-i)
			break
		} else if i > 0 {
			fmt.Fprint(&sb, ", ")
		}
		fmt.Fprint(&sb, elem)
	}
	return sb.String()
}

func WaitingOnNodeClaimTerminationEvent(nodeClass *v1alpha1.VsphereNodeClass, names []string) events.Event {
	return events.Event{
		InvolvedObject: nodeClass,
		Type:           corev1.EventTypeNormal,
		Reason:         "WaitingOnNodeClaimTermination",
		Message:        fmt.Sprintf("Waiting on NodeClaim termination for %s", PrettySlice(names, 5)),
		DedupeValues:   []string{string(nodeClass.UID)},
	}
}
