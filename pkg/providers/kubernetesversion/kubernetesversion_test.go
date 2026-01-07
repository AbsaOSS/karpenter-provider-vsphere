package kubernetesversion

import (
	"context"

	"testing"
	"time"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/mocks"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/version"
	_ "k8s.io/client-go/kubernetes"
)

func TestKubeServerVersions(t *testing.T) {
	var tests = []struct {
		name            string
		expectedError   bool
		version         *version.Info
		expectedVersion string
		discoveryError  error
	}{
		{
			name:            "valid version",
			expectedError:   false,
			version:         &version.Info{GitVersion: "v1.20.4+rke2r1"},
			expectedVersion: "1.20.4",
		},
		{
			name:            "valid version without prefix",
			expectedError:   false,
			version:         &version.Info{GitVersion: "1.21.0"},
			expectedVersion: "1.21.0",
		},
		{
			name:            "valid version with different suffix",
			expectedError:   false,
			version:         &version.Info{GitVersion: "v1.22.1+customsuffix"},
			expectedVersion: "1.22.1+customsuffix",
		},
		{
			name:            "empty version",
			expectedError:   false,
			version:         &version.Info{GitVersion: ""},
			expectedVersion: "",
		},
		{
			name:           "discovery error",
			expectedError:  true,
			version:        nil,
			discoveryError: assert.AnError,
		},
		{
			name:            "funny version",
			expectedError:   false,
			version:         &version.Info{GitVersion: "!!@@##???"},
			expectedVersion: "!!@@##???",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			q := mocks.NewMockInterface(controller)
			d := mocks.NewMockDiscoveryInterface(controller)
			d.EXPECT().ServerVersion().Return(test.version, test.discoveryError).AnyTimes()

			q.EXPECT().Discovery().Return(d).AnyTimes()

			c := cache.New(time.Second*10, time.Second*2)
			provider := NewKubernetesVersionProvider(q, c)

			str, err := provider.KubeServerVersion(context.TODO())

			if test.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedVersion, str)

		})
	}
}
