package instance

import (
	"bytes"
	"fmt"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
	"text/template"
)

const (
	RKE2ConfTemplate = `server: {{.APIEndpoint}}
kubelet-arg:
  - --cloud-provider=external
token: {{.Token}}
{{- if .Taints }}
node-taints:
{{- range taints .Taints }}
  - {{ . }}
{{- end }}
{{- end }}
`
)

type CloudConfig struct {
	WriteFiles []File   `yaml:"write_files" json:"write_files"`
	RunCmd     []string `yaml:"runcmd" json:"runcmd"`
}
type File struct {
	Path        string `yaml:"path" json:"path"`
	Permissions string `yaml:"permissions" json:"permissions"`
	Owner       string `yaml:"owner" json:"owner"`
	Content     string `yaml:"content" json:"content"`
}

func mergeCloudConfig(base, extra *CloudConfig) {
	base.WriteFiles = append(base.WriteFiles, extra.WriteFiles...)
	base.RunCmd = append(base.RunCmd, extra.RunCmd...)
	// Add more fields if CloudConfig has more
}

func formatTaints(taints []corev1.Taint) []string {
	var out []string
	for _, t := range taints {
		if t.Value != "" {
			out = append(out, fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect))
		} else {
			// Value is optional in taints
			out = append(out, fmt.Sprintf("%s:%s", t.Key, t.Effect))
		}
	}
	return out
}

func GetRKE2CloudConfig(data *InitData) (*CloudConfig, error) {
	tmpl := template.Must(template.New("init").Funcs(template.FuncMap{"taints": formatTaints}).Parse(RKE2ConfTemplate))

	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	rke2InstallCmd := fmt.Sprintf("curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=%s INSTALL_RKE2_TYPE=\"agent\" sh -s", data.KubeVersion)
	if data.Distro == v1alpha1.RKE2AirGapped {
		rke2InstallCmd = "INSTALL_RKE2_ARTIFACT_PATH=/opt/rke2-artifacts INSTALL_RKE2_TYPE=\"agent\" sh /opt/install.sh"
	}
	base := &CloudConfig{
		WriteFiles: []File{
			{
				Path:        "/etc/rancher/rke2/config.yaml",
				Permissions: "0640",
				Owner:       "root:root",
				Content:     buf.String(),
			},
		},
		RunCmd: []string{
			"sleep 10",
			rke2InstallCmd,
			"systemctl enable rke2-agent.service",
			"systemctl start rke2-agent.service",
		},
	}
	// Parse AdditionalUserData if provided
	if data.AdditionalUserData != "" {
		var extra CloudConfig
		if err := yaml.UnmarshalStrict([]byte(data.AdditionalUserData), &extra); err != nil {
			return nil, fmt.Errorf("failed to parse AdditionalUserData: %w", err)
		}
		mergeCloudConfig(base, &extra)
	}
	return base, nil
}
