# The Project – Todo Application

This project implements a simple Todo application using Kubernetes.
It consists of three components:

- **todo-app** – serves HTML, handles image caching and frontend routing
- **todo-backend** – REST API for managing todos
- **PostgreSQL** – database for storing todos (StatefulSet)

All components run in the `project` namespace.

---

## Architecture Overview

Browser -> Ingress  
-> **todo-app** (HTML + image caching)  
-> **todo-backend** (`GET /todos`, `POST /todos`)  
-> **PostgreSQL** (StatefulSet, persistent storage)

---

## Deploy with Kustomize

```
kubectl apply -k .
```

---

## Namespace

Apply namespace:
```
kubectl apply -f the_project/manifests/project-ns.yaml
```


## PostgreSQL (StatefulSet)
PostgreSQL is deployed as a StatefulSet with one replica and persistent storage.

### Secrets
The repository contains an encrypted Secret manifest for PostgreSQL
(using SOPS + age). The private key is intentionally not included.

To run the project locally, create your own Secret matching the expected keys:
```
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: project
type: Opaque
stringData:
  POSTGRES_DB: todo
  POSTGRES_USER: todo
  POSTGRES_PASSWORD: todopassword
  DATABASE_URL: "postgres://todo:todopassword@postgres:5432/todo?sslmode=disable"
```

Apply it:
```
kubectl apply -f the_project/postgres/postgres-secret.yaml
```

### Deploy PostgreSQL

```
kubectl apply -f the_project/postgres/manifests/
```

## Deploy todo-backend

todo-backend provides:
- GET /todos
- POST /todos

It reads the database connection string from the DATABASE_URL
environment variable provided by the Postgres Secret.

```
kubectl apply -f todo-backend/manifests/
kubectl rollout status deployment/todo-backend -n project
```

## Deploy todo-app

todo-app serves the frontend, caches a random image, and routes
todo requests to todo-backend via Ingress.

```
kubectl apply -f todo-app/manifests/
kubectl rollout status deployment/todo-app -n project
```

## Accessing the application
If using the k3d setup above, open in browser:

```
http://localhost:8081/
```

## CronJob: hourly “Read <URL>” todo

A Kubernetes CronJob creates a new todo every hour with text:

`Read <random Wikipedia article URL>`

Apply it:

```
kubectl apply -f the_project/manifests/cronjob-wiki-todo.yaml
kubectl get cronjobs -n project
```

## Logging and Monitoring

The project includes centralized logging using **Grafana Loki** and
**Grafana Alloy**.

- **todo-backend** emits structured request logs (including rejected
  todos longer than 140 characters) to stdout.
- **Grafana Alloy** runs as a DaemonSet and collects Kubernetes pod logs.
- Logs are forwarded to **Loki** and can be explored in **Grafana**.

Monitoring components run in separate namespaces and are installed
via Helm:
- `prometheus` – Prometheus + Grafana
- `loki-stack` – Loki
- `alloy` – Grafana Alloy (log collector)

Configuration files are located at `./manifests/loki` (accessed from the root of the repository).
