# karpenter-provider-vsphere

Karpenter provider for VMWare Vsphere

# Required flags

| Flag             | Environment variable | Required |
|------------------|----------------------|----------|
|cluster-name      | CLUSTER_NAME         | true     |
| vsphere-endpoint | GOVC_URL             | true     |
| vsphere-username | GOVC_USERNAME        | true     |
| vsphere-password | GOVC_PASSWORD        | true     |
| vsphere-path     | VSPHERE_PATH         | true     |
| vsphere-insecure | VSPHERE_INSECURE     | false    |

# VsphereNodeClass API
Besides `VSPHERE_PATH` (vsphere folder to place virtulal machines on), all placement settings are defined in `VsphereNodeClass` resource. This is done via selectors:
* `.spec.computeSelector` - defines how to search for desired resourcePool
* `.spec.datastoreSelector` - defines how to search for desired datastore
* `.spec.networkSelector` - difines how to discover network
* `.spec.imageSelector` - VM Template to use for VM Clone

All selectors have `tag` and `name` properties, those are mutually exclusive. Karpenter will find a resource either by Tag or Name.
[!NOTE]
Due to nature of Vsphere tags. Only 1 uniquie tag could be selected to identify desired resource


* `.spec.instanceTypes` - a list of desired instance types, in a following format: `%cpu-%mem-%os`, e.g. 8-16-linux
* `.spec.diskSize` - a desired root volume size in Gigabytes

* `.spec.tags` - a list of tags to apply to Karpenter managed virtual machines
  [!NOTE]
  At least two tags must be specified explicitly:
  * `topology.kubernetes.io/zone` and `k8s-zone` to satisfy Vsphere Cloud controller manager which bootstraps Kubernetes node and removes `unitialized` Taint.


* `.spec.userdata`:
  - `type` - Either `ignition` or `cloud-init`
  - `templateBase64` - A base64 encoded template
  - `values` - a v1.Secret reference (name/namespace) to key/values used in a template