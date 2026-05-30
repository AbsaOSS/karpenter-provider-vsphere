package userdata

import (
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
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
