package userdata

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
	corev1 "k8s.io/api/core/v1"
)

var (
	testTaint = corev1.Taint{
		Effect: corev1.TaintEffectNoSchedule,
		Key:    "karpenter.sh/controller",
		Value:  "true",
	}
	expectedRKE2DefaultIgnition       = []byte(`{"ignition":{"config":{"replace":{"verification":{}}},"proxy":{},"security":{"tls":{}},"timeouts":{},"version":"3.4.0"},"kernelArguments":{},"passwd":{},"storage":{"files":[{"group":{"name":"root"},"overwrite":true,"path":"/etc/rancher/rke2/config.yaml","user":{"name":"root"},"contents":{"compression":"","source":"data:,server%3A%20%0Akubelet-arg%3A%0A%20%20-%20--cloud-provider%3Dexternal%0Atoken%3A%20foo%0Anode-taint%3A%0A%20%20-%20karpenter.sh%2Fcontroller%3Dtrue%3ANoSchedule%0A","verification":{}},"mode":416},{"group":{},"overwrite":true,"path":"/etc/node-join.sh","user":{},"contents":{"compression":"gzip","source":"data:;base64,H4sIAAAAAAAC/2zLMWuDQBQA4P1+xasdy91V2y4Fhw4OUrFFpdBJTnnqkYse955CID8+mCmB7N/3/KQ7O+vO0CQIGSSuC3jrcTDWCUEO0UP8Kvo1OJA0FDAxe/rUekRW4YCJsgucIS/r5qso2uo7S9q/rKrznzLdYvX2od5fdhbie9P8/2ZpZEacOQKaQJKgEzEee3aAs+kcwv7klSjCsNkebwyxCfyIXAIAAP//XDZ0ZNMAAAA=","verification":{}},"mode":448}]},"systemd":{"units":[{"contents":"[Unit]\nDescription=kube node join\nWants=network-online.target\nAfter=network-online.target network.target\nConditionPathExists=!/etc/nodejoin-success.complete\n[Service]\nUser=root\n# To not restart the unit when it exits, as it is expected.\nType=oneshot\nExecStart=/etc/node-join.sh\n[Install]\nWantedBy=multi-user.target\n","enabled":true,"name":"node-join.service"}]}}`)
	expectedRKE2IgnitionWithExtra     = []byte(`{"ignition":{"config":{"replace":{"verification":{}}},"proxy":{},"security":{"tls":{}},"timeouts":{},"version":"3.4.0"},"kernelArguments":{},"passwd":{"users":[{"name":"root","passwordHash":"$6$123$ezMtMFnbc7OSrvgIi1PP5i/NTW46cmPrNlwfMDv5dQI9RWvpoe4MsrGJmcNAFA4Rh9N1BiXnEG403YlaVeAHD."}]},"storage":{"files":[{"group":{"name":"root"},"overwrite":true,"path":"/etc/rancher/rke2/config.yaml","user":{"name":"root"},"contents":{"compression":"","source":"data:,server%3A%20%0Akubelet-arg%3A%0A%20%20-%20--cloud-provider%3Dexternal%0Atoken%3A%20foo%0Anode-taint%3A%0A%20%20-%20karpenter.sh%2Fcontroller%3Dtrue%3ANoSchedule%0A","verification":{}},"mode":416},{"group":{},"overwrite":true,"path":"/etc/node-join.sh","user":{},"contents":{"compression":"gzip","source":"data:;base64,H4sIAAAAAAAC/2zLMWuDQBQA4P1+xasdy91V2y4Fhw4OUrFFpdBJTnnqkYse955CID8+mCmB7N/3/KQ7O+vO0CQIGSSuC3jrcTDWCUEO0UP8Kvo1OJA0FDAxe/rUekRW4YCJsgucIS/r5qso2uo7S9q/rKrznzLdYvX2od5fdhbie9P8/2ZpZEacOQKaQJKgEzEee3aAs+kcwv7klSjCsNkebwyxCfyIXAIAAP//XDZ0ZNMAAAA=","verification":{}},"mode":448}]},"systemd":{"units":[{"contents":"[Unit]\nDescription=kube node join\nWants=network-online.target\nAfter=network-online.target network.target\nConditionPathExists=!/etc/nodejoin-success.complete\n[Service]\nUser=root\n# To not restart the unit when it exits, as it is expected.\nType=oneshot\nExecStart=/etc/node-join.sh\n[Install]\nWantedBy=multi-user.target\n","enabled":true,"name":"node-join.service"}]}}`)
	expectedRKE2AirGapDefaultIgnition = []byte(`{"ignition":{"config":{"replace":{"verification":{}}},"proxy":{},"security":{"tls":{}},"timeouts":{},"version":"3.4.0"},"kernelArguments":{},"passwd":{},"storage":{"files":[{"group":{"name":"root"},"overwrite":true,"path":"/etc/rancher/rke2/config.yaml","user":{"name":"root"},"contents":{"compression":"","source":"data:,server%3A%20%0Akubelet-arg%3A%0A%20%20-%20--cloud-provider%3Dexternal%0Atoken%3A%20foo%0Anode-taint%3A%0A%20%20-%20karpenter.sh%2Fcontroller%3Dtrue%3ANoSchedule%0A","verification":{}},"mode":416},{"group":{},"overwrite":true,"path":"/etc/node-join.sh","user":{},"contents":{"compression":"gzip","source":"data:;base64,H4sIAAAAAAAC/2zNsW6DMBDG8d1PcaUzuGVnsCqqoqIKUS+Z0IGOYMUxlu8SKW8fkSyJlP33ff/3Nz26oEfkRTEJ5HRaIbpIMzqvFHuiCJ8fqvn7t6Zth/63LgfT2+bbfNmhM/an0msUnQ5U5pjEzTgJwxO3u66uMtxTkAx4gdvABRb0vti6FxY6TuKBAo6e4H62+YIpnd1ED4YFk7wi1wAAAP//pROMBMsAAAA=","verification":{}},"mode":448}]},"systemd":{"units":[{"contents":"[Unit]\nDescription=kube node join\nWants=network-online.target\nAfter=network-online.target network.target\nConditionPathExists=!/etc/nodejoin-success.complete\n[Service]\nUser=root\n# To not restart the unit when it exits, as it is expected.\nType=oneshot\nExecStart=/etc/node-join.sh\n[Install]\nWantedBy=multi-user.target\n","enabled":true,"name":"node-join.service"}]}}`)
	initData                          = &InitData{
		Token:              "foo",
		KubeVersion:        "v1.35.4+rke2r1",
		Taints:             []corev1.Taint{testTaint},
		NodeName:           "testnode",
		AdditionalUserData: "",
	}
)

const (
	expectedRKE2CloudConfig = `#cloud-config
hostname: testnode

write_files:
  - path: /etc/rancher/rke2/config.yaml
    permissions: "0640"
    content: |
      server: 
      kubelet-arg:
        - --cloud-provider=external
      token: foo
      node-taint:
        - karpenter.sh/controller=true:NoSchedule

runcmd:
  - sleep 10
  - curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.35.4+rke2r1 INSTALL_RKE2_TYPE="agent" sh -s
  - systemctl enable rke2-agent.service
  - systemctl start rke2-agent.service`

	extraCloudConfigCmd = `runcmd:
  - foo bar baz`
	extraCloudConfigFile = `write_files:
  - path: /tmp/foo
    permissions: "0555"
    content: |
      test file`
	expectedRKE2CloudConfigWithExtraCmd = `#cloud-config
hostname: testnode

write_files:
  - path: /etc/rancher/rke2/config.yaml
    permissions: "0640"
    content: |
      server: 
      kubelet-arg:
        - --cloud-provider=external
      token: foo
      node-taint:
        - karpenter.sh/controller=true:NoSchedule

runcmd:
  - sleep 10
  - curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.35.4+rke2r1 INSTALL_RKE2_TYPE="agent" sh -s
  - systemctl enable rke2-agent.service
  - systemctl start rke2-agent.service
  - foo bar baz`
	expectedRKE2CloudConfigWithExtraFiles = `#cloud-config
hostname: testnode

write_files:
  - path: /etc/rancher/rke2/config.yaml
    permissions: "0640"
    content: |
      server: 
      kubelet-arg:
        - --cloud-provider=external
      token: foo
      node-taint:
        - karpenter.sh/controller=true:NoSchedule

  - path: /tmp/foo
    permissions: "0555"
    content: |
      test file

runcmd:
  - sleep 10
  - curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=v1.35.4+rke2r1 INSTALL_RKE2_TYPE="agent" sh -s
  - systemctl enable rke2-agent.service
  - systemctl start rke2-agent.service`

	extraIgnition = `variant: "fcos"
version: "1.5.0"
passwd:
  users:
  - name: root
    password_hash: $6$123$ezMtMFnbc7OSrvgIi1PP5i/NTW46cmPrNlwfMDv5dQI9RWvpoe4MsrGJmcNAFA4Rh9N1BiXnEG403YlaVeAHD.`
)

func TestRKE2Ignition(t *testing.T) {
	initType := &InitType{
		Distro: v1alpha1.Distro("rke2"),
		Format: v1alpha1.UserDataTypeIgnition,
	}
	factory := &Factory{}
	gen, par, err := factory.Build(initType)
	// it should not err
	assert.Nil(t, err)

	data, err := gen.Generate(initData)
	assert.Nil(t, err)

	res, err := par.Render(data, initData.AdditionalUserData)
	assert.Nil(t, err)
	assert.Equal(t, string(expectedRKE2DefaultIgnition), string(res))

}
func TestRKE2AirGapIgnition(t *testing.T) {
	initType := &InitType{
		Distro: v1alpha1.Distro("rke2airgapped"),
		Format: v1alpha1.UserDataTypeIgnition,
	}
	factory := &Factory{}
	gen, par, err := factory.Build(initType)
	// it should not err
	assert.Nil(t, err)

	data, err := gen.Generate(initData)
	assert.Nil(t, err)

	res, err := par.Render(data, initData.AdditionalUserData)
	assert.Nil(t, err)
	assert.Equal(t, string(expectedRKE2AirGapDefaultIgnition), string(res))
}

func TestRKE2CloudConfig(t *testing.T) {
	initType := &InitType{
		Distro: v1alpha1.Distro("rke2"),
		Format: v1alpha1.UserDataTypeCloudConfig,
	}
	factory := &Factory{}
	gen, par, err := factory.Build(initType)
	// it should not err
	assert.Nil(t, err)
	data, err := gen.Generate(initData)
	assert.Nil(t, err)
	res, err := par.Render(data, initData.AdditionalUserData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRKE2CloudConfig, string(res))
}

func TestRKE2CloudConfigWithExtraCmds(t *testing.T) {
	initType := &InitType{
		Distro: v1alpha1.Distro("rke2"),
		Format: v1alpha1.UserDataTypeCloudConfig,
	}
	factory := &Factory{}
	gen, par, err := factory.Build(initType)
	// it should not err
	assert.Nil(t, err)
	data, err := gen.Generate(initData)
	assert.Nil(t, err)
	res, err := par.Render(data, extraCloudConfigCmd)
	assert.Nil(t, err)
	assert.Equal(t, expectedRKE2CloudConfigWithExtraCmd, string(res))
}

func TestRKE2CloudConfigWithExtraFiles(t *testing.T) {
	initType := &InitType{
		Distro: v1alpha1.Distro("rke2"),
		Format: v1alpha1.UserDataTypeCloudConfig,
	}
	factory := &Factory{}
	gen, par, err := factory.Build(initType)
	// it should not err
	assert.Nil(t, err)
	data, err := gen.Generate(initData)
	assert.Nil(t, err)
	res, err := par.Render(data, extraCloudConfigFile)
	assert.Nil(t, err)
	assert.Equal(t, expectedRKE2CloudConfigWithExtraFiles, string(res))
}

func TestRKE2IgnitionWithExtraButane(t *testing.T) {
	initType := &InitType{
		Distro: v1alpha1.Distro("rke2"),
		Format: v1alpha1.UserDataTypeIgnition,
	}
	factory := &Factory{}
	gen, par, err := factory.Build(initType)
	// it should not err
	assert.Nil(t, err)
	data, err := gen.Generate(initData)
	assert.Nil(t, err)
	res, err := par.Render(data, extraIgnition)
	assert.Nil(t, err)
	assert.Equal(t, string(expectedRKE2IgnitionWithExtra), string(res))
}

// RKE2 expects the singular "node-taint" key (see
// https://docs.rke2.io/install/configuration); "node-taints" is silently
// ignored by the agent (see commit b7bfcfd).
func TestGetCommon_UsesSingularNodeTaintKey(t *testing.T) {
	config, err := getCommon(initData, "install-cmd")
	assert.Nil(t, err)
	require.Len(t, config.Files, 1)

	content := config.Files[0].Content
	assert.Contains(t, content, "node-taint:")
	assert.NotContains(t, content, "node-taints:")
}

func TestGetCommon_OmitsNodeTaintKeyWhenNoTaints(t *testing.T) {
	input := &InitData{
		Token:       "foo",
		KubeVersion: "v1.35.4+rke2r1",
		NodeName:    "testnode",
	}

	config, err := getCommon(input, "install-cmd")
	assert.Nil(t, err)
	require.Len(t, config.Files, 1)

	assert.NotContains(t, config.Files[0].Content, "node-taint")
}

const testKubeletConfig = `apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
maxPods: 110
kubeReserved:
    cpu: 80m
    memory: 1465Mi
`

func initDataWithKubelet() *InitData {
	withKubelet := *initData
	withKubelet.KubeletConfig = testKubeletConfig
	return &withKubelet
}

func TestGetCommon_WritesKubeletConfig(t *testing.T) {
	config, err := getCommon(initDataWithKubelet(), "install-cmd")
	require.NoError(t, err)
	require.Len(t, config.Files, 2)

	assert.Equal(t, "/etc/rancher/rke2/config.yaml", config.Files[0].Path)
	assert.Equal(t, RKE2KubeletConfigPath, config.Files[1].Path)
	assert.Equal(t, "0644", config.Files[1].Permissions)
	assert.Equal(t, strings.TrimSuffix(testKubeletConfig, "\n"), config.Files[1].Content)
}

func TestRKE2CloudConfigWithKubeletConfig(t *testing.T) {
	for _, distro := range []string{"rke2", "rke2airgapped"} {
		t.Run(distro, func(t *testing.T) {
			gen, renderer, err := (&Factory{}).Build(&InitType{
				Distro: v1alpha1.Distro(distro),
				Format: v1alpha1.UserDataTypeCloudConfig,
			})
			require.NoError(t, err)
			data, err := gen.Generate(initDataWithKubelet())
			require.NoError(t, err)
			res, err := renderer.Render(data, "")
			require.NoError(t, err)

			assert.Contains(t, string(res), `  - path: /var/lib/rancher/rke2/agent/etc/kubelet.conf.d/50-karpenter.conf
    permissions: "0644"
    content: |
      apiVersion: kubelet.config.k8s.io/v1beta1
      kind: KubeletConfiguration
      maxPods: 110
      kubeReserved:
          cpu: 80m
          memory: 1465Mi
`)

			// The rendered user data must still be valid cloud-config YAML.
			var parsed DistroConfig
			require.NoError(t, yaml.Unmarshal(res, &parsed))
			require.Len(t, parsed.Files, 2)
			assert.Equal(t, testKubeletConfig, parsed.Files[1].Content)
		})
	}
}

func TestRKE2IgnitionWithKubeletConfig(t *testing.T) {
	gen, renderer, err := (&Factory{}).Build(&InitType{
		Distro: v1alpha1.Distro("rke2"),
		Format: v1alpha1.UserDataTypeIgnition,
	})
	require.NoError(t, err)
	data, err := gen.Generate(initDataWithKubelet())
	require.NoError(t, err)
	res, err := renderer.Render(data, "")
	require.NoError(t, err)

	var ignition struct {
		Storage struct {
			Files []struct {
				Path     string `json:"path"`
				Contents struct {
					Compression string `json:"compression"`
					Source      string `json:"source"`
				} `json:"contents"`
			} `json:"files"`
		} `json:"storage"`
	}
	require.NoError(t, json.Unmarshal(res, &ignition))

	found := false
	for _, f := range ignition.Storage.Files {
		if f.Path != RKE2KubeletConfigPath {
			continue
		}
		found = true
		assert.Equal(t, testKubeletConfig, decodeIgnitionSource(t, f.Contents.Source, f.Contents.Compression))
	}
	require.True(t, found, "kubelet config file should be in the ignition config")
}

// decodeIgnitionSource decodes an Ignition data URL. Butane picks whichever encoding is
// shortest (plain, base64 or gzip+base64), so the test must handle all of them.
func decodeIgnitionSource(t *testing.T, source, compression string) string {
	t.Helper()
	meta, payload, ok := strings.Cut(strings.TrimPrefix(source, "data:"), ",")
	require.True(t, ok, "not a data URL: %s", source)

	raw := []byte(payload)
	if strings.HasSuffix(meta, ";base64") {
		decoded, err := base64.StdEncoding.DecodeString(payload)
		require.NoError(t, err)
		raw = decoded
	} else {
		unescaped, err := url.PathUnescape(payload)
		require.NoError(t, err)
		raw = []byte(unescaped)
	}
	if compression == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(raw))
		require.NoError(t, err)
		unzipped, err := io.ReadAll(reader)
		require.NoError(t, err)
		raw = unzipped
	}
	return string(raw)
}
