package instance

const (
	RKE2ButaneTemplate = `
version: 3.5.0
systemd:
units:
- name: rke2-install.service
  enabled: true
  contents: |
    [Unit]
    Description=rke2-install
    Wants=network-online.target
    After=network-online.target network.target
    ConditionPathExists=!/etc/cluster-api/bootstrap-success.complete
    [Service]
    User=root
    # To not restart the unit when it exits, as it is expected.
    Type=oneshot
    ExecStart=/etc/rke2-install.sh
    [Install]
    WantedBy=multi-user.target
`
)
