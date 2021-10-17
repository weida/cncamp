# simple httpserver container

一个简单的 HTTP 服务器镜像:

- 构建本地镜像。
- 编写 Dockerfile 将练习 2.2 编写的 httpserver 容器化（请思考有哪些最佳实践可以引入到 Dockerfile 中来）。
- 将镜像推送至 Docker 官方镜像仓库。
- 通过 Docker 命令本地启动 httpserver。
- 通过 nsenter 进入容器查看 IP 配置。


## 编译

- 生成镜像

```
make release
```


- 上传镜像

```
# docker login -u caoweida2004
Password:
WARNING! Your password will be stored unencrypted in /root/.docker/config.json.
Configure a credential helper to remove this warning. See
https://docs.docker.com/engine/reference/commandline/login/#credentials-store

Login Succeeded

# make push

```

- 下载镜像

```
# docker images
REPOSITORY                                                        TAG       IMAGE ID       CREATED        SIZE
ubuntu                                                            latest    ba6acccedd29   40 hours ago   72.8MB
centos                                                            latest    5d0da3dc9764   4 weeks ago    231MB
...
# docker pull caoweida2004/httpserver:v1.0
v1.0: Pulling from caoweida2004/httpserver
7b1a6ab2e44d: Already exists
cbbc6cf9a918: Already exists
Digest: sha256:6e5e9ec3701f719388f79b855cb3ea48b13ffaf3cf255df7f61417528dd58b8c
Status: Downloaded newer image for caoweida2004/httpserver:v1.0
docker.io/caoweida2004/httpserver:v1.0
# docker images
REPOSITORY                                                        TAG       IMAGE ID       CREATED          SIZE
caoweida2004/httpserver                                           v1.0      050e9a51bc5e   55 minutes ago   79MB
ubuntu                                                            latest    ba6acccedd29   40 hours ago     72.8MB
centos                                                            latest    5d0da3dc9764   4 weeks ago      231MB
...
```


## 验证


- 启动镜像
```
docker run caoweida2004/httpserver:v1.0 -P 8090
```

- 查看地址

```
# docker container ls
CONTAINER ID   IMAGE                          COMMAND                  CREATED          STATUS          PORTS      NAMES
71171e09fc49   caoweida2004/httpserver:v1.0   "/bin/sh -c /httpser…"   28 seconds ago   Up 27 seconds   8090/tcp   elated_swanson

# docker inspect 71171e09fc49|grep Pid
            "Pid": 86380,
            "PidMode": "",
            "PidsLimit": null,

# nsenter -n -t 86380
[root@dev container]# ipaddr
-bash: ipaddr: command not found
[root@dev container]# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
8: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever

```


- 发送请求

```
#curl -v http://172.17.0.2:8090/ -H "X-FirstName:garlic"
[root@dev container]# curl -v http://172.17.0.2:8090/ -H "X-FirstName:garlic"
*   Trying 172.17.0.2...
* TCP_NODELAY set
* Connected to 172.17.0.2 (172.17.0.2) port 8090 (#0)
> GET / HTTP/1.1
> Host: 172.17.0.2:8090
> User-Agent: curl/7.61.1
> Accept: */*
> X-FirstName:garlic
>
< HTTP/1.1 200 OK
< Accept: */*
< User-Agent: curl/7.61.1
< Version: 1.1
< X-Firstname: garlic
< X-Request-Id: 1634488557609513602
< Date: Sun, 17 Oct 2021 16:35:57 GMT
< Content-Length: 12
< Content-Type: text/plain; charset=utf-8
<
hello world
* Connection #0 to host 172.17.0.2 left intact

```

- 服务端日志

```
# docker run caoweida2004/httpserver:v1.0 -P 8090
2021/10/17 16:30:36 Server is starting...
2021/10/17 16:35:57 <- [1634488557609513602] 172.17.0.2:54906 GET /
2021/10/17 16:35:57 -> [1634488557609513602] 172.17.0.2:54906 GET / 200

```



- 关闭

```
# docker exec -it f2af4b9b2777 bash
root@f2af4b9b2777:/# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 16:44 ?        00:00:00 /bin/sh -c /httpserver -P 8090
root           8       1  0 16:44 ?        00:00:00 /httpserver
root          14       0  0 16:45 pts/0    00:00:00 bash
root          24      14  0 16:45 pts/0    00:00:00 ps -ef
root@f2af4b9b2777:/# kill -SIGUSR1 8

```

- 服务端显示
```
[root@dev ~]# docker run caoweida2004/httpserver:v1.0 -P 8090
2021/10/17 16:44:43 Server is starting...
2021/10/17 16:45:32 Server is shutting...
2021/10/17 16:45:32 Server stopped

```
