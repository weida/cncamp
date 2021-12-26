#  httpserver 部署到 istio 集群

##  build docker image

- create httpserver image

```
# cd httpserver0
# docker build -f Dockerfile -t caoweida2004/httpserver0:v1.3 .

# cd ../httpserver1
# docker build -f Dockerfile -t caoweida2004/httpserver1:v1.3 .

# cd ../httpserver2
# docker build -f Dockerfile -t caoweida2004/httpserver2:v1.3 .
```

- upload image

```
# docker login -u caoweida2004
Password:
WARNING! Your password will be stored unencrypted in /root/.docker/config.json.
Configure a credential helper to remove this warning. See
https://docs.docker.com/engine/reference/commandline/login/#credentials-store

Login Succeeded

# docker push caoweida2004/httpserver0:v1.3
# docker push caoweida2004/httpserver1:v1.3
# docker push caoweida2004/httpserver2:v1.3

```


##  install istio

- download & install 

```
$curl -L https://istio.io/downloadIstio | sh -
```
or download tgz from [Istio release page](https://github.com/istio/istio/releases/tag/1.12.1)

```
$tar zxvf istio-1.12.1-linux-amd64.tar.gz
```

- install

```
$cp bin/istioctl /usr/local/bin
$istioctl install --set profile=demo -y
```


##  create namespace

- Add a namespace label to instruct Istio to automatically inject Envoy sidecar proxies

```
$kubectl create namespace istiospace
$kubectl label ns istiospace istio-injection=enabled
```
注意: istio-injection=enabled 拼写错误不会报错,未设置的话，jaeger仅能看到一个services

## configmap

```
$ kubectl create -f httpserver-configmap.yaml -n istiospace
```

## deployment

- deployment

```
$ kubectl create -f httpserver0-deploy.yaml -n istiospace
$ kubectl create -f httpserver1-deploy.yaml -n istiospace
$ kubectl create -f httpserver2-deploy.yaml -n istiospace

$ kubectl get pods -n istiospace
NAME                                 READY   STATUS    RESTARTS   AGE
httpserver0-istio-d9cccdd58-kt728    2/2     Running   0          103s
httpserver1-istio-6ffd5c5bb9-69pbw   2/2     Running   0          11s
httpserver2-istio-7444f6cc88-f2l6j   2/2     Running   0          6s

                                 
```


## service

```
$ kubectl create -f service0.yaml -n istiospace
$ kubectl create -f service1.yaml -n istiospace
$ kubectl create -f service2.yaml -n istiospace

$ kubectl get service -n istiospace
NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
httpserver-service0   ClusterIP   10.100.239.48    <none>        80/TCP    77m
httpserver-service1   ClusterIP   10.103.120.227   <none>        80/TCP    10s
httpserver-service2   ClusterIP   10.107.144.126   <none>        80/TCP    5s

```


## secret


- gen key

```
$openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=httpserver-istio.com Inc./CN=*.httpserver-istio.com' -keyout httpserver-istio.com.key -out httpserver-istio.com.crt
```

- create secret

```
$kubectl create -n istio-system secret tls httpserver-credential --key=httpserver-istio.com.key --cert=httpserver-istio.com.crt

$kubectl get secret -n istio-system|grep httpserver
httpserver-credential                              kubernetes.io/tls                     2      102
```

## istio virtualservice & gateway

- get pod dns and config the virtualservice

```
$ kubectl get pods -n istiospace
NAME                                 READY   STATUS    RESTARTS   AGE
httpserver0-istio-d9cccdd58-5km82    1/1     Running   0          22m
httpserver1-istio-6ffd5c5bb9-mlds8   1/1     Running   0          22m
httpserver2-istio-7444f6cc88-cfjbr   1/1     Running   0          21m

$ kubectl exec -t httpserver0-istio-d9cccdd58-5km82  -n istiospace -- cat /etc/resolv.conf
nameserver 10.96.0.10
search istiospace.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
```

- create virtualservrice & gateway

istio-virtualservice.yaml
```
  - name : "test-route"
    match:
    - uri:
        prefix: /test
    rewrite:
      uri: /healthz
    route:
      - destination:
          host: httpserver-service0.istiospace.svc.cluster.local
          port:

```

```
$ kubectl  create -f istio-virtualservice.yaml -n istiospace
$ kubectl get virtualservice -n istiospace
```


```
$ kubectl  create -f istio-gateway.yaml -n istiospace
$ kubectl get gateway -n istiospace
```

- get ingressgateway ip

```
$ kubectl  get svc -n istio-system
```

- modify /etc/hosts

```
 10.105.232.156 httpserver-istio.com
```


## install jaeger

```
$ kubectl apply -f jaeger.yaml
$ kubectl edit configmap istio -n istio-system
set tracing.sampling=100
```

```
apiVersion: v1
data:
  mesh: |-
    accessLogFile: /dev/stdout
    defaultConfig:
      discoveryAddress: istiod.istio-system.svc:15012
      proxyMetadata: {}
      tracing:
        sampling: 100
```

- remote terminal access

```
kubectl port-forward -n istio-system  $(kubectl get pod -n istio-system -l app=jaeger -o jsonpath='{.items[0].metadata.name}')   --address 0.0.0.0 8888:16686 &
```


## test the result

```
$ curl -k https://httpserver-istio.com/test
200
$ curl -k https://httpserver-istio.com/trace
```

 Jaeger Dashboard
![avatar](https://github.com/weida/cncamp/blob/master/homework/istio/pic/jaegerboard.png)

 Jaeger item
![avatar](https://github.com/weida/cncamp/blob/master/homework/istio/pic/jaegeritem.png)


## referece
- https://istio.io/latest/docs/setup/getting-started/
- https://github.com/cncamp/101/tree/master/module12/
