# Azure Service Bus Prometheus exporter

Azure Service Bus metrics Prometheus exporter. Does not use Azure Monitor,
connects to the Service Bus directly and scrapes all queues, topics and
subscriptions.

## TL;DR;

```console
$ helm install giggio/servicebusprometheusexporter
```

## Introduction

This chart bootstraps a [Azure Service Bus Prometheus exporter](https://github.com/marcinbudny/servicebus_exporter) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.12+
- Helm 2.11+ or Helm 3.0-beta3+

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release giggio/servicebusprometheusexporter
```

The command deploys Azure Service Bus Prometheus exporter on the Kubernetes cluster in the default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the Azure Service Bus Prometheus exporter chart and their default values.

|            Parameter                      |                                  Description                                  |                           Default                            |
| ----------------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------ |
| `connectionString` | Connection string for Azure Service Bus. The exporter requires `Manage` SAS policy. | `nil` (required) |
| `timeout` | Timeout when scraping the Service Bus | `30s` |
| `verbose` | Enable verbose logging | `false` |
| `addPromAnnotations` | Adds Prometheus annotations so that scrapping is automatic. | `true` |
| `rewriteAppHTTPProbers` | Adds `sidecar.istio.io/rewriteAppHTTPProbers` annotation to pods so service health checks work with MTLS when using Istio. | `false` |
| `imagePullSecrets` | Global Docker registry secret names as an array | `[]` (does not add image pull secrets to deployed pods)      |
| `image.repository` | Image name | `marcinbudny/servicebus_exporter` |
| `image.tag` | Image tag | `{TAG_NAME}` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.pullSecrets` | Specify docker-registry secret names as an array | `[]` (does not add image pull secrets to deployed pods)      |
| `nameOverride` | String to partially override servicebusprometheusexporter.fullname template with a string (will prepend the release name) | `nil`                               |
| `fullnameOverride` | String to fully override servicebusprometheusexporter.fullname template with a string | `nil` |
| `service.type` | Kubernetes Service type | `ClusterIP` |
| `service.port` | Service HTTP port where the metrics will listen on. | `9580` |
| `ingress.enabled` | Enable ingress controller resource | `false` |
| `ingress.annotations` | Ingress annotations | `[]` |
| `ingress.hosts[0].host` | Hostname to service installation | `nil` |
| `ingress.hosts[0].paths[0]` | Paths within the url structure | `[]` |
| `ingress.tls[0].hosts[0]` | TLS hosts | `nil` |
| `ingress.tls[0].secretName` | TLS Secret (certificates) | `nil` |
| `virtualservice.enabled` | Enable Istio Virtual Service | `false` |
| `virtualservice.gateways` | Istio gateways for the Virtual Service | `[]` |
| `virtualservice.hosts` | Istio hosts for the Virtual Service | `[]` |
| `virtualservice.matches` | Http matches for the  Virtual Service | `[ { uri: { prefix: '/metrics' } } ]` |
| `resources` | Resource for the pods (limits, requests etc.) | `{}` |
| `podSecurityContext` | Pod security context | `{}` |
| `securityContext` | Security context | `{}` |
| `nodeSelector` | Node labels for pod assignment | `{}` |
| `tolerations` | List of node taints to tolerate | `[]` |
| `affinity` | Map of node/pod affinities | `{}`                                                         |

The first three parameters and `service.port` map to the command line arguments for the binary.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install --name my-release \
  --set 'connectionString=<your connection string>' \
    giggio/servicebusprometheusexporter
```

The above command sets the connection string to `<your connection string>`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install --name my-release -f values.yaml giggio/servicebusprometheusexporter
```

> **Tip**: You can use the default [values.yaml](values.yaml).

> **Tip**: Only set the connection string. You don't need to enable ingress or
> an Istio Virtual Service. The scrapper will be found because of the annotations.
