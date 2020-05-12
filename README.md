# Wid's missing tools for Kubernetes (Wim-T-K)

# publish-files

Publish files as ConfigMap within a Pod

```
# In Pod
echo "a content" > a.txt
wimtk -v publish-files a.txt
```

```
# From the same namespace as the Pod
$ kubectl get configmap/wimtk -o yaml
apiVersion: v1
data:
  a.txt: |
    a content
kind: ConfigMap
metadata:
...

```

# wait-pods

Wait untils pod1 or pod2 are in the "Running" Phase

```
wimtk -v wait-pods 'pod1-.*' pod2 
```

# sync-map

Sync configMap from another namespace

```
wimtk -v sync-map -n othername confimap-to-replicate 
```

