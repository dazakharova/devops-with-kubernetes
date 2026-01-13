# Exercise 4.3. Prometheus

Query that shows the number of pods created by StatefulSets in prometheus namespace:

```
count(kube_pod_info{namespace="prometheus", created_by_kind="StatefulSet"})
```
