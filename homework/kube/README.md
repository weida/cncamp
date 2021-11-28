#  httpserver 部署到 kubernetes 集群

## create resource

- create namespace

```
$kubectl create namespace httpserverspace
```

-  create quota

```
$kubectl create -f compute-resources.yaml  -n httpserverspace
resourcequota/compute-resources created

$kubectl get quota -n httpserverspace
NAME                AGE   REQUEST                                         LIMIT
compute-resources   4s    requests.cpu: 0/125m, requests.memory: 0/64Mi   limits.cpu: 0/250m, limits.memory: 0/128Mi
```


## deployment

- configmap

```
$kubectl create -f httpserver-configmap.yaml --namespace=httpserverspace
configmap/httpserver-config created
```


- deployment

```
$kubectl create -f httpserver-deploy.yaml -n httpserverspace
deployment.apps/httpserver-deployment created
```


## service

```
$ kubectl create -f service.yaml -n httpserverspace
service/httpserver-service created
$ kubectl get services -n httpserverspace
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
httpserver-service   ClusterIP   10.104.60.230   <none>        80/TCP    8s
```


## ingress


- gen key

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=httpserver-ingress.com/O=httpserver-ingress.com"
```

- create secret

```
$ kubectl create -f secret.yaml -n httpserverspace
secret/httpserver-tls created
```

- create ingress 

```
$ kubectl  create -f ingress.yaml -n httpserverspace
ingress.networking.k8s.io/httpserver-ingress created

$ kubectl get ingress -n httpserverspace
NAME                 CLASS    HOSTS                    ADDRESS          PORTS     AGE
httpserver-ingress   <none>   httpserver-ingress.com   192.168.37.202   80, 443   17s

```

- modify /etc/hosts

```
192.168.37.202 httpserver-ingress.com
```


## test the result

```
$ curl -k https://httpserver-ingress.com/
hello world
```




## referece
https://segmentfault.com/a/1190000040618813 
