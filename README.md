# Wid's missing tools for Kubernetes (Wim-T-K)

NOTE: all those commands needs special RBAC rights

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

# wait-phase

Wait untils anything that starts with pod1- and pod2 are in the "Running" Phase

```
wimtk -v wait-phase --phase-watched Running 'pod1-.*' pod2 
```

# wait-condition

Wait untils anything that starts with pod1- and pod2 have condition Ready=True

```
wimtk -v wait-condition  --condition Ready=True 'pod1-.*' pod2 
```

# sync-map

Sync confimap-to-replicate from othernamespace into current namespace

```
wimtk -v sync-map -n othernamespace confimap-to-replicate 
  ```

