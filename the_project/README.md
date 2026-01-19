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

## Database backups

The project includes a Kubernetes CronJob that creates daily backups of the
PostgreSQL database and uploads them to Google Cloud Storage.

- Schedule: once every 24 hours
- Backup tool: `pg_dump`
- Storage: Google Cloud Storage bucket
- Authentication: GCP Service Account mounted as a Secret

The required GCP credentials secret (`gcs-sa`) is created manually in the cluster
and is intentionally not committed to the repository.



## DBaaS vs DIY

## DIY (Self-managed database solution)

### Pros

1. High level of customization: Tailored specifically to organizational needs.
2. Full control of infrastructure: Freedom in choosing hardware, software, and configurations.
3. Own security management: Transparent and direct control over security measures. 
4. Independence from vendor lock-in. 

### Cons

1. Significant initial costs: hardware and setup. 
2. Ongoing operational expenses: Continual administrative and maintenance costs.
3. Requires specialized personnel 
4. Manual updates and patches: Manual patches increase vulnerability. 
5. Backup complexity
6. Scaling challenges: Difficulties in quickly expanding infrastructure. 

## DBaaS

### Pros 

1. Quick setup: Rapid deployment through automated provisioning. 
2. Cost-efficient: Lower initial investment compared to buying hardware.
3. Automatic management: Automized updates, monitoring, and scaling.
4. Scalable: Easy to scale resources based on demand.
5. Modern tech access: Use latest technologies without large investment.
6. No internal staff requirements
7. Enhanced security: Provider takes responsibility for protecting data and ensuring high levels of uptime.
8. Ease of Maintenance: Low operational effort required. 

### Cons 

1. Lock-in risks: Harder to switch vendor later. 
2. Limited configurations: Less flexible when optimizing performance. 
3. Monthly fees: Subscription model could become expensive.
4. Less control: Reduced visibility into inner workings.
5. Third-Party Dependence: Uptime affected by third-party outages.
6. Security Concerns: Data hosted externally raises potential compliance issues.

## Message Queue Integration (NATS)

This project uses **NATS** for asynchronous communication between services.

### Todo Backend
- Publishes events to NATS subject `todos.events` when:
    - a todo is created
    - a todo is marked as done

### Broadcaster Service
- Subscribes to `todos.events` using a **queue subscription**
- Forwards todo status updates to **Telegram**

### Architecture
todo-backend -> NATS -> broadcaster -> Telegram

### Environment Variables (Broadcaster)
- `NATS_URL`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`
- `NATS_SUBJECT` (default: `todos.events`)
- `NATS_QUEUE` (default: `broadcaster`)
