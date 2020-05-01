---
title: Bare metal Kubernetes at home
published: 
intro: 
---

After a bit deliberating, I decided to setup a small Kubernetes (k8s) cluster at home. There are a few reasons for doing this:

* Reduce my AWS bill (move test environments on AWS to k8s cluster)
* Be able to run prototypes at home without impacting how I use my home workstation (gaming PC) 
* I've been wanting to learn k8s

This post will cover what is Kubernetes, how I built and configured my cluster, and how to deploy pods and services.

## What is Kubernetes

Kubernetes are an orchastration platform that supports managing containers cross multiple nodes (servers). The platform comes in many flavors and is interacted with via a CLI tool called `kubectl`. Once the K8s cluster is setup and running, `kubectl` is used to deploy new pods, deployments, services, volumes, and more. It's also used to get cluster status information and drain/flag nodes for maintence. 

### Major objects in the K8s ecosystem

* Pods
* Nodes
* Deployments
* Services
* Volumes

#### Pods

#### Nodes

#### Deployments

#### Services

#### Volumes

### Major services in the K8s ecosystem

#### Container hosting nodes

* `kubelet` - Receives pod specifications and ensures the node 
* `kube-proxy` -
* Container Runtime - Docker, containerd, rktlet, etc...

#### Master Nodes

* `etcd` - A distributed key value store that is used to store cluster data. It's the primary datastore. 
* `kube-apiserver`
* `kube-scheduler`

## My Setup

### Parts

### Network configuration

### Kubeadm

## Managing the cluster

## Conclusion
