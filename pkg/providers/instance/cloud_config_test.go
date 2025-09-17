package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRKE2CloudConfig(t *testing.T) {
	tests := []struct {
		name        string
		initData    *InitData
		expectError bool
		validate    func(*testing.T, *CloudConfig)
	}{
		{
			name: "base config without additional user data",
			initData: &InitData{
				Token:       "abc123",
				APIEndpoint: "https://api.example.com",
				NodeName:    "node1",
				Distro:      "rke2",
				KubeVersion: "v1.31.1+rke2r1",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *CloudConfig) {
				assert.NotNil(t, cfg)
				assert.Len(t, cfg.WriteFiles, 1)
				assert.Contains(t, string(cfg.WriteFiles[0].Content), "https://api.example.com")
				assert.Contains(t, cfg.RunCmd, "curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.31.1+rke2r1 INSTALL_RKE2_TYPE=\"agent\" sh -s")
				assert.Contains(t, cfg.RunCmd, "systemctl start rke2-agent.service")
			},
		},
		{name: "base config for rke2airgapped distro",
			initData: &InitData{
				Token:       "abc123",
				APIEndpoint: "https://api.example.com",
				NodeName:    "node1",
				Distro:      "rke2airgapped",
				KubeVersion: "v1.31.1+rke2r1",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *CloudConfig) {
				assert.NotNil(t, cfg)
				assert.Contains(t, cfg.RunCmd, "INSTALL_RKE2_ARTIFACT_PATH=/opt/rke2-artifacts INSTALL_RKE2_TYPE=\"agent\" sh /opt/install.sh")
			},
		},
		{
			name: "merge additional user data",
			initData: &InitData{
				Token:       "def456",
				APIEndpoint: "https://api.merge.com",
				NodeName:    "node2",
				Distro:      "rke2",
				AdditionalUserData: `
write_files:
  - path: /etc/extra.conf
    permissions: "0644"
    owner: root:root
    content: |
      EXTRA=1
runcmd:
  - echo "extra command"
`,
			},
			expectError: false,
			validate: func(t *testing.T, cfg *CloudConfig) {
				assert.NotNil(t, cfg)
				// base + extra
				assert.Len(t, cfg.WriteFiles, 2)
				assert.Equal(t, "/etc/extra.conf", cfg.WriteFiles[1].Path)
				assert.Contains(t, cfg.RunCmd, "echo \"extra command\"")
			},
		},
		{
			name: "invalid additional user data",
			initData: &InitData{
				APIEndpoint:        "https://api.fail.com",
				Distro:             "rke2",
				AdditionalUserData: ":::notyaml:::",
			},
			expectError: true,
			validate:    func(t *testing.T, cfg *CloudConfig) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := GetRKE2CloudConfig(tt.initData)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, cfg)
			}
		})
	}
}
