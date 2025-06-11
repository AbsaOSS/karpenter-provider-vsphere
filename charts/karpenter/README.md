# karpenter

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

A Helm chart for Karpenter, an open-source node provisioning project built for Kubernetes.

**Homepage:** <https://karpenter.sh/>

## Source Code

* <https://github.com/absaoss/karpenter-provider-vsphere/>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| additionalAnnotations | object | `{}` | Additional annotations to add into metadata. |
| additionalClusterRoleRules | list | `[]` | Specifies additional rules for the core ClusterRole. |
| additionalLabels | object | `{}` | Additional labels to add into metadata. |
| affinity | object | `{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"karpenter.sh/nodepool","operator":"DoesNotExist"}]}]}},"podAntiAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":[{"topologyKey":"kubernetes.io/hostname"}]}}` | Affinity rules for scheduling the pod. If an explicit label selector is not provided for pod affinity or pod anti-affinity one will be created from the pod selector labels. |
| controller.containerName | string | `"controller"` | Distinguishing container name (containerName: karpenter-controller). |
| controller.env | list | `[]` | Additional environment variables for the controller pod. |
| controller.envFrom | list | `[]` |  |
| controller.extraVolumeMounts | list | `[]` | Additional volumeMounts for the controller container. |
| controller.healthProbe.port | int | `8081` | The container port to use for http health probe. |
| controller.image.repository | string | `"ghcr.io/absaoss/karpenter-provider-vsphere"` | Repository path to the controller image. |
| controller.image.tag | string | `"0.0.1"` | Tag of the controller image. |
| controller.metrics.port | int | `8080` | The container port to use for metrics. |
| controller.resources | object | `{}` | Resources for the controller container. |
| controller.securityContext.appArmorProfile | object | `{}` | AppArmor profile for the controller container. |
| controller.securityContext.seLinuxOptions | object | `{}` | SELinux options for the controller container. |
| controller.securityContext.seccompProfile | object | `{}` | Seccomp profile for the controller container. |
| controller.sidecarContainer | list | `[]` | Additional sidecarContainer config |
| controller.sidecarVolumeMounts | list | `[]` | Additional volumeMounts for the sidecar - this will be added to the volume mounts on top of extraVolumeMounts |
| dnsConfig | object | `{}` | Configure DNS Config for the pod |
| dnsPolicy | string | `"ClusterFirst"` | Configure the DNS Policy for the pod |
| extraVolumes | list | `[]` | Additional volumes for the pod. |
| fullnameOverride | string | `""` | Overrides the chart's computed fullname. |
| hostNetwork | bool | `false` | Bind the pod to the host network. This is required when using a custom CNI. |
| imagePullPolicy | string | `"IfNotPresent"` | Image pull policy for Docker images. |
| imagePullSecrets | list | `[]` | Image pull secrets for Docker images. |
| initContainers | object | `{}` | add additional initContainers to run before karpenter container starts |
| logErrorOutputPaths | list | `["stderr"]` | Log errorOutputPaths - defaults to stderr only |
| logLevel | string | `"info"` | Global log level, defaults to 'info' |
| logOutputPaths | list | `["stdout"]` | Log outputPaths - defaults to stdout only |
| nameOverride | string | `""` | Overrides the chart's name. |
| nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node selectors to schedule the pod to nodes with labels. |
| podAnnotations | object | `{}` | Additional annotations for the pod. |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podDisruptionBudget.name | string | `"karpenter"` |  |
| podLabels | object | `{}` | Additional labels for the pod. |
| podSecurityContext | object | `{"fsGroup":65532,"runAsNonRoot":false,"seccompProfile":{"type":"RuntimeDefault"}}` | SecurityContext for the pod. |
| priorityClassName | string | `"system-cluster-critical"` | PriorityClass name for the pod. |
| replicas | int | `2` | Number of replicas. |
| revisionHistoryLimit | int | `10` | The number of old ReplicaSets to retain to allow rollback. |
| schedulerName | string | `"default-scheduler"` | Specify which Kubernetes scheduler should dispatch the pod. |
| service.annotations | object | `{}` | Additional annotations for the Service. |
| serviceAccount.annotations | object | `{}` | Additional annotations for the ServiceAccount. |
| serviceAccount.create | bool | `true` | Specifies if a ServiceAccount should be created. |
| serviceAccount.name | string | `""` | The name of the ServiceAccount to use. If not set and create is true, a name is generated using the fullname template. |
| serviceMonitor.additionalLabels | object | `{}` | Additional labels for the ServiceMonitor. |
| serviceMonitor.enabled | bool | `false` | Specifies whether a ServiceMonitor should be created. |
| serviceMonitor.endpointConfig | object | `{}` | Configuration on `http-metrics` endpoint for the ServiceMonitor. Not to be used to add additional endpoints. See the Prometheus operator documentation for configurable fields https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md#endpoint |
| serviceMonitor.metricRelabelings | list | `[]` | Metric relabelings for the `http-metrics` endpoint on the ServiceMonitor. For more details on metric relabelings, see: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs |
| serviceMonitor.relabelings | list | `[]` | Relabelings for the `http-metrics` endpoint on the ServiceMonitor. For more details on relabelings, see: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |
| settings | object | `{"clusterName":"","preferencePolicy":"Respect","vmMemoryOverheadPercent":0.075}` | Global Settings to configure Karpenter |
| settings.clusterName | string | `""` | Cluster name. |
| settings.preferencePolicy | string | `"Respect"` | How the Karpenter scheduler should treat preferences. Preferences include preferredDuringSchedulingIgnoreDuringExecution node and pod affinities/anti-affinities and ScheduleAnyways topologySpreadConstraints. Can be one of 'Ignore' and 'Respect' |
| settings.vmMemoryOverheadPercent | float | `0.075` | The VM memory overhead as a percent that will be subtracted from the total memory for all instance types. The value of `0.075` equals to 7.5%. |
| strategy | object | `{"rollingUpdate":{"maxUnavailable":1}}` | Strategy for updating the pod. |
| terminationGracePeriodSeconds | string | `nil` | Override the default termination grace period for the pod. |
| tolerations | list | `[{"key":"CriticalAddonsOnly","operator":"Exists"}]` | Tolerations to allow the pod to be scheduled to nodes with taints. |
| topologySpreadConstraints | list | `[{"maxSkew":1,"topologyKey":"topology.kubernetes.io/zone","whenUnsatisfiable":"DoNotSchedule"}]` | Topology spread constraints to increase the controller resilience by distributing pods across the cluster zones. If an explicit label selector is not provided one will be created from the pod selector labels. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
