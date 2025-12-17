# Ping-pong (Exercise: Stateful applications)

Ping-pong responds on `GET /pingpong` and increments a counter.  
The counter is stored in PostgreSQL (StatefulSet, 1 replica).

Ping-pong responds to:
- `GET /pingpong` -> returns `pong N` and increments counter
- `GET /pings` -> returns current counter value

## Prerequisites (local k3d example)
This repository was tested using k3d with an Ingress/LoadBalancer mapping:

```
k3d cluster create \
  --port 8082:30080@agent:0 \
  -p 8081:80@loadbalancer \
  --agents 2
```  

## Namespace
This application runs in the `exercises` namespace.

```
kubectl apply -f ../manifests/exercises-ns.yaml
```


## Ping-pong responds to:
- `GET /pingpong` -> returns `pong N` and increments counter
- `GET /pings` -> returns current counter value

## PostgreSQL (StatefulSet)
PostgreSQL is deployed as a StatefulSet in `exercises`.

The repository contains an encrypted secret manifest for Postgres configuration (SOPS + age).
Because the private key is not in the repository.

- Create and apply a Secret that matches the credentials used by ping-pong’s DATABASE_URL
in `ping_pong/manifests/deployment.yaml`.
- Apply it:
```
kubectl apply -f postgres-secret.yaml
```

- Then deploy Postgres manifests:
``` 
kubectl apply -f postgres/
```

### Initialize DB Schema

```
CREATE TABLE pingpong_counter (
id integer primary key,
value integer not null
);

INSERT INTO pingpong_counter (id, value) VALUES (1, 0);
```

## Deploy Ping-Pong 

Ping-pong reads `DATABASE_URL` from its Deployment manifest.
**The credentials in the DB secret must match this DSN.**

```
kubectl apply -f manifests/
 ```

## Access

Ping-pong is exposed via the shared Ingress (configured with log-output: `kubectl apply -f /log_output/manifests/ingress.yaml`) at:
- http://localhost:8081/pingpong
- http://localhost:8081/pings

Example (GKE):
- http://<INGRESS_IP>/pinpong

## Notes on Kubernetes configuration

### PostgreSQL data directory
The PostgreSQL StatefulSet mounts a PersistentVolume.
Because Kubernetes volumes may contain a `lost+found` directory,
PostgreSQL is configured to use a subdirectory via `subPath`
to allow proper initialization of the data directory.

### Service exposure
For this exercise, the ping-pong Service is exposed using
`type: LoadBalancer` instead of Ingress to allow direct access
to the application via an external IP.
