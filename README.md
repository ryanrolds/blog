# pedantic_orderliness

Markdown and Go Template driven blog.

## Setup

```
make install
make build
```

## Running

```
./pedantic_orderliness
```

## Deploying

### PO Kubernetes

GitHub Actions is configured to use Helm to deploy to a Kubernetes cluster.

#### Manual Steps

Apply Helm template:
```
helm template test chart --set image.tag=latest --set image.repository=pedanticorderliness/pedantic_orderliness > kubectl apply -f -
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
