# Setup KubeCI

## Build and Install Engine

```
$ cd /home/dipta/go/src/kube.ci/engine
$ git checkout deploy
$ ./hack/docker/setup.sh
$ docker tag kubeci/kubeci:deploy diptadas/kubeci:deploy
$ docker push diptadas/kubeci:deploy
$ export APPSCODE_ENV=dev; ./hack/deploy/install.sh --docker-registry=diptadas
```

## Expose web-ui

```
$ kubectl port-forward -n kube-system {operator-pod} 9090:9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

- Status: `http://127.0.0.1:9090/namespaces/{namespace}/workplans/{workplan-name}`
- Log: `http://127.0.0.1:9090/namespaces/{namespace}/workplans/{workplan-name}/steps/{step-name}`

## Build and Install GitApiserver

```
$ cd /home/dipta/go/src/kube.ci/git-apiserver
$ git checkout deploy
$ ./hack/docker/setup.sh
$ docker tag kubeci/git-apiserver:deploy diptadas/git-apiserver:deploy
$ docker push diptadas/git-apiserver:deploy
$ export APPSCODE_ENV=dev; ./hack/deploy/install.sh --docker-registry=diptadas
```

## Create Repository CRD

```
$ kubectl apply -f repository.yaml
$ kubectl get branches
$ kubectl get tags
$ kubectl get pullrequests
```
## Create Workflow CRD

```
$ kubectl create secret generic github-credential --from-literal=TOKEN=...
$ kubectl apply -f rbac.yaml
$ kubectl apply -f workflow.yaml
```

## Cleanup

```
$ export APPSCODE_ENV=dev; ./hack/deploy/install.sh --docker-registry=diptadas --uninstall
$ kubectl delete secret github-credential
$ kubectl delete -f repository.yaml
$ kubectl delete -f rbac.yaml
$ kubectl delete -f workflow.yaml
```
