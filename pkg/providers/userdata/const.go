package userdata

const (
	installRKE2Cmd    = `curl -sfL https://get.rke2.io | INSTALL_RKE2_VERSION=%s INSTALL_RKE2_TYPE="agent" sh -s`
	InstallRKE2AirGap = `INSTALL_RKE2_ARTIFACT_PATH=/opt/rke2-artifacts INSTALL_RKE2_TYPE="agent" sh /opt/install.sh`
	butaneTemplate    = `
variant: fcos
version: 1.5.0
systemd:
  units:
    - name: node-join.service
      enabled: true
      contents: |
        [Unit]
        Description=kube node join
        Wants=network-online.target
        After=network-online.target network.target
        ConditionPathExists=!/etc/nodejoin-success.complete
        [Service]
        User=root
        # To not restart the unit when it exits, as it is expected.
        Type=oneshot
        ExecStart=/etc/node-join.sh
        [Install]
        WantedBy=multi-user.target
storage:
  files:
    {{- range .Files }}
    - path: {{ .Path }}
      {{- $owner := ParseOwner .Owner }}
      {{ if $owner.User -}}
      user:
        name: {{ $owner.User }}
      {{- end }}
      {{ if $owner.Group -}}
      group:
        name: {{ $owner.Group }}
      {{- end }}
      # Owner
      {{ if ne .Permissions "" -}}
      mode: {{ .Permissions }}
      {{ end -}}
      overwrite: true
      contents:
        inline: |
          {{ .Content | Indent 10 }}
    {{- end }}
    - path: /etc/node-join.sh
      mode: 0700
      overwrite: true
      contents:
        inline: |
          #!/bin/bash
          set -euo pipefail
          {{ range .Commands }}
          {{ . | Indent 10 }}
          {{- end }}
`
)
