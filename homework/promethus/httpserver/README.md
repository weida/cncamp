# simple httpserver

一个简单的 HTTP 服务器:

1. 接收客户端 request，并将 request 中带的 header 写入 response header
2. 读取当前系统的环境变量中的 VERSION 配置，并写入 response header
3. Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
4. 当访问 localhost/healthz 时，应返回 200


## 使用

```
go run httpserver
```

默认端口8090,  通过 --listenAddr 可以设置参数

```
go run httpserver --listenAddr 8099
```


客户端可以使用curl验证
```
curl -v http://localhost:8090/ -H "X-FirstName:garlic"
```

并发8个进程发送

```
seq 1 1000 | xargs -I % -P 8 curl  "http://localhost:8090/healthz"
```



## 说明


### 服务部分

使用net/http实现http服务， 使用http.Handle指定文根对应的处理函数，设置http.Server，通过ListenAndServe完成启动
使用flag实现启动端口可参数化。

### 心跳部分 
通过"/healthz"对应处理函数实现, 此处增加了`UP`, `DOWN`两个状态, 由于通过signal判断服务运行状态使用另外一个携程关闭服务，
通过atomic设置其状态，保证原子性。


### 日志部分
通过http.Server的Handler设置http日志处理，在日志中增加了请求的ID, 用于区分不同的请求， 同时通过context将值发送给
响应处理 index 和 healthz


### 服务关闭
设置signal对应的处理函数, 与linux中直接定义处理函数不同， golang实现中通过chan os.Signal进行传递, 打开框架如下

```
quit := make(chan os.Signal, 1)
done := make(bool, 1)

//注册信号
signal.Notify(sigs, syscall.SIGTERM, ....)

go func () {
   <-quit 
   // 退出处理
   ...
   done<-true
}()

//正常服务功能
...
<-done

```
### 其他

可以通过以下命令设置环境变量
```
export VERSION=1.0 
```


