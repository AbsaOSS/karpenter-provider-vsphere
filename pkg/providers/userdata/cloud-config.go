package userdata

import (
	"bytes"
	"text/template"

	"go.yaml.in/yaml/v3"
)

type CloudConfigRenderer struct{}

const cloudConfigTpl = `#cloud-config
hostname: {{ .NodeName }}

write_files:
{{- range .Files }}
  - path: {{ .Path }}
    permissions: "{{ .Permissions }}"
    content: |
{{ .Content | indent 6 }}
{{- end }}
runcmd:
{{- range .Commands }}
  - {{ . }}
{{- end }}`

func (r *CloudConfigRenderer) Render(data *DistroConfig, additional string) ([]byte, error) {
	var additionalCfg DistroConfig
	tpl := template.Must(template.New("cloud").Funcs(template.FuncMap{
		"indent": func(spaces int, v string) string {
			pad := bytes.Repeat([]byte(" "), spaces)
			lines := bytes.Split([]byte(v), []byte("\n"))

			var out bytes.Buffer
			for _, l := range lines {
				out.Write(pad)
				out.Write(l)
				out.WriteByte('\n')
			}

			return out.String()
		},
	}).Parse(cloudConfigTpl))

	if additional != "" {
		if err := yaml.Unmarshal([]byte(additional), &additionalCfg); err != nil {
			return nil, err
		}
		data.Files = append(data.Files, additionalCfg.Files...)
		data.Commands = append(data.Commands, additionalCfg.Commands...)
	}
	var out bytes.Buffer
	err := tpl.Execute(&out, map[string]any{
		"NodeName": data.NodeName,
		"Files":    data.Files,
		"Commands": data.Commands,
	})

	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
