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

## Notes

* .md files need to be `\n` not `\r\n` otherwise Blackfriday will not render code blocks correctly

## Deploying

Build the image and push to K8s and update K8s:
```
make push_k8s
make deploy_k8s
```

Build and deploy to AWS (production):
```
make login_aws
make push_aws
make deploy_aws
```

> AWS to uses `latest` tag for images which is not desirable. Need to fix this sometime.
