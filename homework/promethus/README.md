#  使用Promethus 监控 httpserver

##  install Promethus 


yaml files in  prometheus/kubernetes-prometheus

- create namespace monitoring

```
$ kubectl create namespace monitoring
```

-  create Role 

```
$ kubectl create -f clusterRole.yaml
```

- create Config Map

```
$ kubectl create -f config-map.yaml
```

- create Prometheus Deployment

```
$ kubectl create  -f prometheus-deployment.yaml 
```

- create Service of Prometheus

```
$ kubectl create -f prometheus-service.yaml --namespace=monitoring
```

prometheus-service.yaml

```
...
spec:
  selector:
    app: prometheus-server
  type: NodePort
  ports:
    - port: 8080
      targetPort: 9090
      nodePort: 30000
```


- access  and  test

```
http://192.168.37.202:30000/
```


##  install Grafana 

- create configuration  

```
$ kubectl create -f grafana-datasource-config.yaml
```

- create Deployment 

```
$ kubectl create -f deployment.yaml
```

- create service

```
$ kubectl create -f service.yaml
```

service.yaml

```
spec:
  selector:
    app: grafana
  type: NodePort
  ports:
    - port: 3000
      targetPort: 3000
      nodePort: 32000
```

- access and test

```
http://192.168.37.202:32000
```


##  create httpserver with metric 

- create httpserver image

```
# docker build -f Dockerfile -t caoweida2004/httpserver:v1.2 .
```

- upload image

```
# docker login -u caoweida2004
Password:
WARNING! Your password will be stored unencrypted in /root/.docker/config.json.
Configure a credential helper to remove this warning. See
https://docs.docker.com/engine/reference/commandline/login/#credentials-store

Login Succeeded

# docker push caoweida2004/httpserver:v1.2
```

##  httpserver deployment

- create httpserver deployment

```
$kubectl create -f httpserver-deploy.yaml
```



## test the result

- req test 

```
$ k get pods httpserver-metric-d4478c7d5-ttbn5 -oyaml
...
  podIPs:
  - ip: 172.16.9.50

```

```
$ seq 1 1000 | xargs -I % -P 8 curl  "http://172.16.9.50:8090/hello"
```

- result

 Promethus Dashboard
 ![image](https://github.com/weida/cncamp/blob/master/homework/promethus/pic/prometheus.png)

 Grafana Dashboard
 ![image](https://github.com/weida/cncamp/blob/master/homework/promethus/pic/grafana.png)



## referece

 https://devopscube.com/setup-prometheus-monitoring-on-kubernetes/
 https://devopscube.com/setup-grafana-kubernetes/

