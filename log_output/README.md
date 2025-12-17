# Log Output Application

The Log output application periodically generates a random string and timestamp
and exposes them via an HTTP endpoint.

It also:
- Fetches the current ping / pong count from the Ping-pong application via HTTP
- Reads configuration from a ConfigMap
- Demonstrates multi-container Pods with a shared volume

This application runs in the `exercises` namespace.

---

## Architecture

The application consists of **two containers in a single Pod**:

1. **log-writer**
    - Generates a random string on startup
    - Writes a timestamp and the random string to a shared file every 5 seconds

2. **log-reader**
    - Serves HTTP requests
    - Reads the shared file
    - Fetches ping / pong count from the Ping-pong service
    - Reads configuration from a ConfigMap

---

## Prerequisites (local k3d example)

This application was tested using k3d with an Ingress controller:

```
k3d cluster create \
  --port 8082:30080@agent:0 \
  -p 8081:80@loadbalancer \
  --agents 2
```

## Namespace 

```
kubectl apply -f ../manifests/exercises-ns.yaml
```

## Deploy Log Output 

```
kubectl apply -f manifests/
```

## Accessing the application 

Log output is exposed via Ingress.

If using the k3d setup above, open in browser:
```
http://localhost:8081/status
```

Example (GKE):
- http://<INGRESS_IP>/
- http://<INGRESS_IP>/status

Example output:

```
file content: this text is from file
env variable: MESSAGE=hello world
2024-03-30T12:15:17.705Z: 8523ecb1-c716-4cb6-a044-b9e83bb98e43
Ping / Pongs: 3
```

