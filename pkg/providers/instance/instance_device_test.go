package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/mock/gomock"
)

func TestGetDiskConfigSpecResize(t *testing.T) {
	tests := []struct {
		name              string
		templateDiskSize  int64
		requestedDiskSize int64
		expectError       bool
		errorContains     string
		expectedDiskSize  int64
	}{
		{
			name:              "requested size less than template size should error",
			templateDiskSize:  30,
			requestedDiskSize: 20,
			expectError:       true,
			errorContains:     "can't resize template disk down",
		},
		{
			name:              "requested size equal to template size",
			templateDiskSize:  20,
			requestedDiskSize: 20,
			expectError:       false,
			expectedDiskSize:  20,
		},
		{
			name:              "requested size greater than template size",
			templateDiskSize:  20,
			requestedDiskSize: 40,
			expectError:       false,
			expectedDiskSize:  40,
		},
		{
			name:              "zero requested size defaults to template size",
			templateDiskSize:  25,
			requestedDiskSize: 0,
			expectError:       false,
			expectedDiskSize:  25,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			disk := &types.VirtualDisk{
				VirtualDevice: types.VirtualDevice{
					Key: 2000,
				},
				CapacityInKB: test.templateDiskSize,
			}

			result, err := getDiskConfigSpec(disk, test.requestedDiskSize)

			if test.expectError {
				assert.Error(t, err)
				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)

			spec := result.(*types.VirtualDeviceConfigSpec)
			assert.Equal(t, types.VirtualDeviceConfigSpecOperationEdit, spec.Operation)

			resultDisk := spec.Device.(*types.VirtualDisk)
			assert.Equal(t, test.expectedDiskSize, resultDisk.CapacityInKB)
		})
	}
}
