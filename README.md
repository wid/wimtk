# Wid's missing tools for Kubernetes (wimtk)

# pucon

Publish files as ConfigMap within a Pod


```
# In Pod
echo "a content" > a.txt
pucon a.txt
```

```
# From the same namespace as the Pod
$ kubectl get configmap/pucon
apiVersion: v1
data:
  a.txt: |
    a content
kind: ConfigMap
metadata:
...

```

