package userdata

import (
	"bytes"
	"fmt"
	"text/template"

	corev1 "k8s.io/api/core/v1"
)

const (
	RKE2ConfTemplate = `server: {{.APIEndpoint}}
kubelet-arg:
  - --cloud-provider=external
token: {{.Token}}
{{- if .Taints }}
node-taint:
{{- range taints .Taints }}
  - {{ . }}
{{- end }}
{{- end }}`
)

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

func getCommon(input *InitData, installCMD string) (*DistroConfig, error) {
	tmpl := template.Must(template.New("init").Funcs(template.FuncMap{"taints": formatTaints}).Parse(RKE2ConfTemplate))
	configData := &bytes.Buffer{}
	err := tmpl.Execute(configData, input)
	if err != nil {
		return nil, err
	}
	return &DistroConfig{
		NodeName: input.NodeName,
		Files: []File{
			{
				Owner:       "root:root",
				Permissions: "0640",
				Path:        "/etc/rancher/rke2/config.yaml",
				Content:     configData.String()},
		},
		Commands: []string{
			"sleep 10",
			installCMD,
			"systemctl enable rke2-agent.service",
			"systemctl start rke2-agent.service",
		},
	}, nil
}
