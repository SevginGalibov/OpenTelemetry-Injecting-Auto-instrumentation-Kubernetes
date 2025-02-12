# OpenTelemetry Auto-Instrumentation on Kubernetes

## 📌 Overview
This project demonstrates how to set up OpenTelemetry auto-instrumentation in a Kubernetes cluster using **Kind**, **OpenTelemetry Operator**, **Prometheus**, and **Grafana**.

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
Verify the cluster is created:
```sh
kubectl cluster-info --context kind-kind
```

### 2️⃣ Install Cert-Manager
```sh
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.7.1/cert-manager.yaml
kubectl get pods -n cert-manager
```

### 3️⃣ Install OpenTelemetry Operator
```sh
helm install opentelemetry-operator open-telemetry/opentelemetry-operator \
--set "manager.collectorImage.repository=otel/opentelemetry-collector-k8s"
kubectl get pods -n default
```

### 4️⃣ Deploy Prometheus
```sh
kubectl apply -f prometheus.yaml -n default
kubectl get pods -n default
```

### 5️⃣ Deploy OpenTelemetry Collector
```sh
kubectl apply -f collector.yaml -n default
kubectl get pods -n default
```

### 6️⃣ Deploy Auto-Instrumentation
```sh
kubectl apply -f instrumentation.yaml -n default
kubectl get Instrumentation -n default
```

### 7️⃣ Deploy Sample Applications (Java & .NET)
```sh
kubectl apply -f sevgingalibov-java-sample.yaml -n default
kubectl apply -f sevgingalibov-dotnet-sample.yaml -n default
kubectl get pods -n default
```
**Note:** Auto-instrumentation is activated during the **Init** phase.

### 8️⃣ Install Grafana
```sh
kubectl apply -f grafana.yaml -n default
kubectl get pods -n default
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

## 📜 License
This project is licensed under the MIT License.

## 💡 Contributing
Feel free to submit issues and pull requests to improve this project!

---
Happy monitoring! 🚀
