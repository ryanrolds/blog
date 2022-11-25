# K8s files

> These files were replaced by the Helm chart. See the `chart` directory.

```
export ENV=test
envsubst < k8s/namespace.yaml | kubectl apply -f -

docker build .
export TAG_NAME=$(docker images --format='{{.ID}}' | head -1)
docker tag $TAG_NAME docker.pedanticorderliness.com/pedantic-orderliness:$TAG_NAME
docker push docker.pedanticorderliness.com/pedantic-orderliness:$TAG_NAME
envsubst < k8s/deployment.yaml | kubectl apply -f -

envsubst < k8s/service.yaml | kubectl apply -f -
envsubst < k8s/ingress.yaml | kubectl apply -f -
```
