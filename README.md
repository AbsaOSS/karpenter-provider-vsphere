# karpenter-provider-vsphere

Karpenter provider for VMWare Vsphere

# !!!Early alpha - NOT for Production use!!!

# Required flags

| Flag             | Environment variable | Required |
|------------------|----------------------|----------|
| cluster-name     | CLUSTER_NAME         | true     |
| vsphere-endpoint | GOVC_URL             | true     |
| vsphere-username | GOVC_USERNAME        | true     |
| vsphere-password | GOVC_PASSWORD        | true     |
| vsphere-path     | VSPHERE_FOLDER       | true     |
| vsphere-insecure | GOVC_INSECURE        | false    |

# VsphereNodeClass API
Besides `VSPHERE_FOLDER` (vsphere folder to place virtulal machines on), all placement settings are defined in `VsphereNodeClass` resource. This is done via selectors:
* `.spec.computeSelector` - defines how to search for desired resourcePool
* `.spec.datastoreSelector` - defines how to search for desired datastore
* `.spec.networkSelector` - difines how to discover network
* `.spec.imageSelector` - VM Template to use for VM Clone

All selectors have `tag` and `name` properties, those are mutually exclusive. Karpenter will find a resource either by Tag or Name.

* `.spec.instanceTypes` - a list of desired instance types:
  - `os`: linux
  - `cpu`: number of CPUS
  - `memory`: amount of memory in gigabytes
  - `region`: region topology
  - `zone`: zone topology
  - `maxPods`: maxPods to pass to kubelet (not implemented)
* `.spec.diskSize` - a desired root volume size in Gigabytes

* `.spec.tags` - a list of tags to apply to Karpenter managed virtual machines
  [!NOTE]
  At least two tags must be specified explicitly:
  * `topology.kubernetes.io/zone` and `k8s-zone` to satisfy Vsphere Cloud controller manager which bootstraps Kubernetes node and removes `unitialized` Taint.


* `.spec.userdata`:
  - `type` - Either `ignition` or `cloud-config`
  - `additionalUserdata` - extra init data to be merged with distribution specific

[!NOTE] user-data should hanndle `karpenter.sh/unregistered` taint to the node
