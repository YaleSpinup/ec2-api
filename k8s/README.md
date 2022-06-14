# k8s development Readme

The application ships with a basic k8s config in the `k8s/` directory.  There, you will find an `api` helm chart and a `values.yaml` to deploy the pod, service and ingress.  By default, `skaffold` will use the [paketo buildpacks](https://paketo.io/) and will reference the configuration in `config/config.json`.

## pre-requisites

### install docker desktop, minikube and skaffold

* [Install docker desktop](https://www.docker.com/products/docker-desktop)

* [Install minikube](https://minikube.sigs.k8s.io/docs/start/)

* [Install skaffold](https://skaffold.dev/docs/getting-started/#installing-skaffold)

### setup ingress controller

```
minikube addons enable ingress
```

For other ways to deploy the nginx-ingress controller see https://kubernetes.github.io/ingress-nginx/deploy/

## check

Make sure minikube is running and you can connect to the local k8s cluster:

```
minikube status
```

```
minikube kubectl -- get nodes
```

## develop

* run `skaffold dev` in the root of the project to build a docker image and deploy it to the minikube cluster

* in a separate terminal window, run `minikube tunnel` to expose ports 80/443

* use the endpoint `http://localhost/v1/<apiname>`

Saving your code should rebuild and redeploy your project automatically

## [non-]profit
