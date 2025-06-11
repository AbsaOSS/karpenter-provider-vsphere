package utils

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/awslabs/operatorpkg/serrors"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

var (
	instanceIDRegex = regexp.MustCompile(`(?P<Provider>.*)://(?P<InstanceID>.*)`)
)

func GetAllSingleValuedRequirementLabels(instanceType *cloudprovider.InstanceType) map[string]string {
	labels := map[string]string{}
	if instanceType == nil {
		return labels
	}
	for key, req := range instanceType.Requirements {
		if req.Len() == 1 {
			labels[key] = req.Values()[0]
		}
	}
	return labels
}

func ParseInstanceID(providerID string) (string, error) {
	matches := instanceIDRegex.FindStringSubmatch(providerID)
	if matches == nil {
		return "", serrors.Wrap(fmt.Errorf("provider id does not match known format"), "provider-id", providerID)
	}
	for i, name := range instanceIDRegex.SubexpNames() {
		if name == "InstanceID" {
			return matches[i], nil
		}
	}
	return "", serrors.Wrap(fmt.Errorf("provider id does not match known format"), "provider-id", providerID)
}

func GiToByteAsString(size int64) string {
	return strconv.FormatInt(GiToByte(size), 10)
}
func GiToByte(size int64) int64 {
	return size * 1024 * 1024 * 1024
}
func GiToKb(size int64) int64 {
	return size * 1024 * 1024
}
func GiToMb(size int64) int64 {
	return size * 1024
}
