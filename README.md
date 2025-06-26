# Pedantic Orderliness Blog

Markdown and Go Template driven blog.

## Setup

```
make install
```

## Running

```
./pedantic_orderliness
```

## Deploying

### Kubernetes

GitHub Actions is configured to use Helm to deploy to a Kubernetes cluster.

#### Manual Steps (break glass)

Apply Helm template:
```
docker build .
export TAG_NAME=$(docker images --format='{{.ID}}' | head -1)
docker tag $TAG_NAME docker.pedanticorderliness.com/blog:$TAG_NAME
docker push docker.pedanticorderliness.com/blog:$TAG_NAME
helm template test chart --set image.tag=$TAG_NAME --set image.repository=docker.pedanticorderliness.com/blog > kubectl apply -f -
```

### Production

GitHub Actions is configured to use Helm to deploy to ECS.

#### Manual steps

Build and deploy to AWS (production):
```
make login_aws
make push_aws
make deploy_aws
```

> AWS to uses `latest` tag for images which is not desirable. Need to fix this sometime.

## Notes

* .md files need to be `\n` not `\r\n` otherwise Blackfriday will not render code blocks correctly
