# OpenTelemetry Auto-Instrumentation on Kubernetes

![Application architecture collecting telemetry](https://media.licdn.com/dms/image/v2/D4D12AQHjP5mRSN-oYw/article-inline_image-shrink_1500_2232/article-inline_image-shrink_1500_2232/0/1719942922311?e=1744848000&v=beta&t=99h5A6CpHpbSOPqSvae9schxVjiEcciWaTrcenuqfLk)

## 📌 Overview
This project demonstrates how to set up OpenTelemetry auto-instrumentation in a Kubernetes cluster using Kind, OpenTelemetry Operator, Collector, Prometheus, and Grafana.

It provides auto-instrumentation support for applications developed in .NET, Java, Node.js, and Python, enabling tracing, metrics collection, and observability without modifying application code. OpenTelemetry Operator is used to automatically inject the necessary instrumentation into the deployed workloads, while Prometheus and Grafana are utilized for monitoring and visualization.

## 📋 Prerequisites
Ensure you have the following tools installed on your system:
- [Kind](https://kind.sigs.k8s.io/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/)

## 🚀 Deployment Steps

### 1️⃣ Create a Kubernetes Cluster with Kind
```sh
kind create cluster --config kind-ingress-config.yaml
```

### 2️⃣ Install Cert-Manager
```sh
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.7.1/cert-manager.yaml
```
```sh
kubectl get pods -n cert-manager
```
**Expected Output:**
```
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-7bbdf4ddc7-vsnhk              1/1     Running   0          20s
cert-manager-cainjector-68d96bb946-qbp9p   1/1     Running   0          20s
cert-manager-webhook-57647647b5-4vvpj      1/1     Running   0          20s
```

### 3️⃣ Install OpenTelemetry Operator
```sh
helm install opentelemetry-operator open-telemetry/opentelemetry-operator \
--set "manager.collectorImage.repository=otel/opentelemetry-collector-k8s"
```
```sh
kubectl get pods -n default

```
**Expected Output:**
```sh
NAME                                      READY   STATUS    RESTARTS   AGE
opentelemetry-operator-58dd4c7487-npkb8   2/2     Running   0          100s
```

### 4️⃣ Deploy Prometheus
```sh
kubectl apply -f prometheus.yaml -n default
```
```sh
kubectl get pods -n default
```
**Expected Output:**
```
NAME                                      READY   STATUS    RESTARTS   AGE
opentelemetry-operator-58dd4c7487-npkb8   2/2     Running   0          3m11s
prometheus-64c78655f5-k7l6c               1/1     Running   0          53s
```

### 5️⃣ Deploy OpenTelemetry Collector
```
kubectl apply -f collector.yaml -n default
```
```yaml
apiVersion: opentelemetry.io/v1alpha1
kind: OpenTelemetryCollector
metadata:
  name: simplest        #https://github.com/SevginGalibov/
spec:
  image: otel/opentelemetry-collector-contrib:0.103.0
  config: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
    connectors:
      spanmetrics:
        dimensions:
          - name: http.method
          - name: http.status_code
          - name: http.route
    exporters:
      prometheusremotewrite:
        endpoint: "http://prometheus.default.svc.cluster.local:9090/api/v1/write"       # Prometheus Server
    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: []
          exporters: [spanmetrics]
        metrics:
          receivers: [spanmetrics]
          exporters: [prometheusremotewrite]
```

```sh
kubectl get pods -n default
```
**Expected Output:**
```
NAME                                      READY   STATUS    RESTARTS   AGE
opentelemetry-operator-58dd4c7487-npkb8   2/2     Running   0          4m29s
prometheus-64c78655f5-k7l6c               1/1     Running   0          2m11s
simplest-collector-675f995c9b-227l8       1/1     Running   0          31s
```

### 6️⃣ Deploy Auto-Instrumentation
```sh
kubectl apply -f instrumentation.yaml -n default
```
```yaml
apiVersion: opentelemetry.io/v1alpha1
kind: Instrumentation
metadata:
  name: sevgingalibov-instrumentation
spec:
  exporter:
    endpoint: http://simplest-collector.default.svc.cluster.local:4318		# Collector Server and OTLP
  env:
    - name: OTEL_EXPORTER_OTLP_HEADERS
      value: 'Authorization=Basic x'

  python:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://simplest-collector.default.svc.cluster.local:4318		# Collector Server and OTLP
  dotnet:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://simplest-collector.default.svc.cluster.local:4318		# Collector Server and OTLP
  java:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://simplest-collector.default.svc.cluster.local:4317		# Collector Server and OTLP
  nodejs:
    env:
      - name: OTEL_EXPORTER_OTLP_ENDPOINT
        value: http://simplest-collector.default.svc.cluster.local:4317		# Collector Server and OTLP
```
```sh
kubectl get Instrumentation -n default
```
**Expected Output:**
```
NAME                            AGE   ENDPOINT                                                   SAMPLER   SAMPLER ARG
sevgingalibov-instrumentation   15s   http://simplest-collector.default.svc.cluster.local:4318
```

### 7️⃣ Deploy Sample Applications (Java & .NET)
```sh
kubectl apply -f sevgingalibov-java-sample.yaml -n default

```
```yaml
For Java injection to work, the relevant annotation must be added to the yaml.
annotations:
    instrumentation.opentelemetry.io/inject-java: "true"       
```
```sh
kubectl apply -f sevgingalibov-dotnet-sample.yaml -n default
```
```yaml
For .NET injection to work, the relevant annotation must be added to the yaml.
annotations:
    instrumentation.opentelemetry.io/inject-dotnet: "true"          
```
```sh
kubectl get pods -n default
```
**Expected Output:**
```
NAME                                            READY   STATUS     RESTARTS   AGE
opentelemetry-operator-58dd4c7487-npkb8         2/2     Running    0          7m5s
prometheus-64c78655f5-k7l6c                     1/1     Running    0          4m47s
sevgingalibov-sample-dotnet-b7dfb56b6-f2n4s     0/1     Init:0/1   0          12s   # Injection activated
sevgingalibov-sample-java                       0/1     Init:0/1   0          13s   # Injection activated
simplest-collector-675f995c9b-227l8             1/1     Running    0          3m7s
```
### Let's see the environments injected into the Java pod
```yaml
    Environment:
      OTEL_NODE_IP:                         (v1:status.hostIP)
      OTEL_POD_IP:                          (v1:status.podIP)
      OTEL_EXPORTER_OTLP_ENDPOINT:         http://simplest-collector.default.svc.cluster.local:4317
      JAVA_TOOL_OPTIONS:                    -javaagent:/otel-auto-instrumentation-java/javaagent.jar
      OTEL_EXPORTER_OTLP_HEADERS:          Authorization=Basic x
      OTEL_SERVICE_NAME:                   sevgingalibov-sample-java
      OTEL_RESOURCE_ATTRIBUTES_POD_NAME:   sevgingalibov-sample-java (v1:metadata.name)
      OTEL_RESOURCE_ATTRIBUTES_NODE_NAME:   (v1:spec.nodeName)
      OTEL_RESOURCE_ATTRIBUTES:            k8s.container.name=sevgingalibov-sample-java,k8s.namespace.name=default,k8s.node.name=$(OTEL_RESOURCE_ATTRIBUTES_NODE_NAME),k8s.pod.name=$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME),service.instance.id=default.$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME).sevgingalibov-sample-java,service.version=main
    Mounts:
      /otel-auto-instrumentation-java from opentelemetry-auto-instrumentation-java (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-b496g (ro)
```

### Let's see the environments injected into the .NET pod
```yaml
    Environment:
      OTEL_NODE_IP:                         (v1:status.hostIP)
      OTEL_POD_IP:                          (v1:status.podIP)
      OTEL_EXPORTER_OTLP_ENDPOINT:         http://simplest-collector.default.svc.cluster.local:4318
      CORECLR_ENABLE_PROFILING:            1
      CORECLR_PROFILER:                    {918728DD-259F-4A6A-AC2B-B85E1B658318}
      CORECLR_PROFILER_PATH:               /otel-auto-instrumentation-dotnet/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so
      DOTNET_STARTUP_HOOKS:                /otel-auto-instrumentation-dotnet/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll
      DOTNET_ADDITIONAL_DEPS:              /otel-auto-instrumentation-dotnet/AdditionalDeps
      OTEL_DOTNET_AUTO_HOME:               /otel-auto-instrumentation-dotnet
      DOTNET_SHARED_STORE:                 /otel-auto-instrumentation-dotnet/store
      OTEL_EXPORTER_OTLP_HEADERS:          Authorization=Basic x
      OTEL_SERVICE_NAME:                   sevgingalibov-sample-dotnet
      OTEL_RESOURCE_ATTRIBUTES_POD_NAME:   sevgingalibov-sample-dotnet-b7dfb56b6-f2n4s (v1:metadata.name)
      OTEL_RESOURCE_ATTRIBUTES_NODE_NAME:   (v1:spec.nodeName)
      OTEL_RESOURCE_ATTRIBUTES:            k8s.container.name=sevgingalibov-sample-dotnet,k8s.deployment.name=sevgingalibov-sample-dotnet,k8s.namespace.name=default,k8s.node.name=$(OTEL_RESOURCE_ATTRIBUTES_NODE_NAME),k8s.pod.name=$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME),k8s.replicaset.name=sevgingalibov-sample-dotnet-b7dfb56b6,service.instance.id=default.$(OTEL_RESOURCE_ATTRIBUTES_POD_NAME).sevgingalibov-sample-dotnet,service.version=development-r15
    Mounts:
      /otel-auto-instrumentation-dotnet from opentelemetry-auto-instrumentation-dotnet (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-cj8mv (ro)
```

### 8️⃣ Install Grafana
```sh
kubectl apply -f grafana.yaml -n default
```
```sh
kubectl get pods -n default

```
**Expected Output:**
```
grafana-6475cb6779-4v25z                      1/1     Running             0          33s
opentelemetry-operator-58dd4c7487-npkb8       2/2     Running             0          17m
prometheus-64c78655f5-k7l6c                   1/1     Running             0          15m
sevgingalibov-sample-dotnet-b7dfb56b6-f2n4s   1/1     Running             0          8m53s
sevgingalibov-sample-java                     1/1     Running             0          10m
simplest-collector-675f995c9b-227l8           1/1     Running             0          13m
```

### 9️⃣ Generate a Trace
```sh
kubectl exec -it sevgingalibov-sample-java -- curl "http://localhost:8080"
kubectl exec -it sevgingalibov-sample-dotnet-b7dfb56b6-f2n4s -- curl "http://localhost/healtcheck/ping"
```

## 📊 Access Grafana Dashboard
1. Forward Grafana service:
   ```sh
   kubectl port-forward svc/grafana -n default 8080:80
   ```
2. Open your browser and go to: [http://localhost:8080](http://localhost:8080)
3. **Login Credentials:**
   - **User:** `admin`
   - **Password:** `admin`
4. Import the OpenTelemetry Dashboard:
   - Go to **Create > Import**.
   - Enter **Dashboard ID: 19419**.
   - Click **Load** and configure Prometheus as the data source.
   - Click **Import**.

## 🎉 Visualizing Traces
Check out the traces and metrics collected in Grafana:
![Grafana Dashboard](https://i.imgur.com/DND8a2g.png)
![Grafana Dashboard](https://i.imgur.com/6pmOvGT.png)


## 📜 License
This project is licensed under the MIT License.

## 💡 Contributing
Feel free to submit issues and pull requests to improve this project!

---
Happy monitoring! 🚀
