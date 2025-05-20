# `/healthz`

The `/healthz` endpoint provides a simple health check for `go-dims`. It is intended for use with
load balancers, orchestration systems (like Kubernetes), or uptime monitoring tools.

---

## ✅ Purpose

This endpoint is designed to verify that:

- The HTTP server is up and accepting connections 
- The service can respond in a timely manner

It does **not** validate image backends, S3 access, or disk I/O — it only confirms the process is
running and ready to serve requests.

---

## 🚦 Usage with Kubernetes

You can use it as a `readinessProbe` or `livenessProbe` in Kubernetes:

``` 
readinessProbe: 
  httpGet: 
    path: /healthz 
    port: 8080 
  periodSeconds: 5 
  initialDelaySeconds: 3 
```

