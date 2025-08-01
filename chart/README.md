# 🧠 Auto Namespace Instrumentation Injector

This Helm chart deploys a controller on Kubernetes. Its main purpose is to **automatically apply an OpenTelemetry `Instrumentation` CRD to every newly created namespace.**

---

## 🚀 Features

- Watches the Kubernetes API and detects newly created namespaces.
- Dynamically applies `Instrumentation` custom resources.
- Supports an **ignore list** for namespaces you don’t want to modify.
- Collector endpoint and authentication headers are fully configurable.
- All configurations are managed via `values.yaml`.

---

## 📦 Installation

### 1. Add the Helm repository _(hosted via GitHub Pages)_

```bash
helm repo add namespace-injector https://sevgingalibov.github.io/auto-namespace-instrumentation
helm repo update
```

### 2. Install the chart

```bash
helm install injector namespace-injector/auto-namespace-instrumentation-injector --version 0.0.1
```

---

## ⚙️ Configuration

Customize the behavior using `values.yaml`. Example:

```yaml
ignoreNamespaces:
  - kube-system
  - default
  - monitoring

instrumentation:
  name: otel-auto-instrumentation
  exporterEndpoint: http://simplest-collector.default.svc.cluster.local:4318
  authHeader: 'Authorization=Basic x'
```

---

## 🛡️ License

MIT © 2025 Sevgin Galipoglu
