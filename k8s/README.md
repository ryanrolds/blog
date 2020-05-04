# K8s files

```
export ENV=test
export TAG_NAME=latest
envsubst < deployment.yaml | kubectl apply -f -
envsubst < service.yaml | kubectl apply -f -
kubectl apply -f ingress-test.yaml
```