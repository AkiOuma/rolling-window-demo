# 模仿Hystrix实现一个滑动窗口计数器

## 需求
参考 Hystrix 实现一个滑动窗口计数器

## 实现思路

实习滑动窗口的核心代码位置是 `/pkg/hystrix`

* 定义一个放置请求结果的对象Bucket(`/pkg/hystrix/bucket.go`)
* 定义一个滑动窗口对象RollingWindown(`/pkg/hystrix/rolling-window.go`)
  * 核心属性是一个实现了FIFO方法的Bucket切片
  * Bucket切片的规定大小的上限
  * 当前是否处于熔断状态
  * 熔断触发的请求总数的阈值
  * 熔断触发的请求失败率的阈值
  * 熔断触发持续时间
  * 上次熔断触发的时间戳
* RollingWindow中实现以下方法
  * 每隔一秒输出一次当前是否处于熔断状态
  * 每隔100毫秒在Bucket队列中放入一个新的Bucket，若队列满了则丢去队首的Bucket
  * 统计队列中所有的Bucket的总请求次数与失败率，若达到设定的阈值，则将熔断状态改为true，并记录档次熔断时间
  * 若熔断状态为true，当当前时间与上次熔断记录时间到达记录的最大间隔，则将熔断状态改回false
* 为滑动窗口包装一个gin的中间件(`pkg/hystrix/wrapper.go`)


## 测试用例

实现位置是`testcase/`

实现以下三个端：
* 客户端

  `testcase/client/client.go`
  ```bash
  # 启动命令
  make client
  ```

  请求上游服务

* 上游服务端

  `testcase/server/upstream.go`
  ```bash
  # 启动命令
  make up
  ```

  使用gin实现服务，并使用`pkg/hystrix/wrapper.go`中定义的中间件，对下游服务发起请求

* 下游服务端

  `testcase/server/downstream.go`
  ```bash
  # 启动命令
  make down
  ```

  初始化的时候传入一个成功率，当该服务被请求时，会根据定义的成功率返回请求成功(状态码200)或者失败(状态码500)

## 测试结果
测试使用的参数为：
* 客户端单次发其并发请求100次
* 下游服务成功率为20%
* 熔断阈值，总请求数为50，总失败率为80%
* 熔断生效时长5秒

上游服务端

从控制台信息中可以看到，服务最开始监听的时候，每次打印出来的熔断状态都会是false，直到第一批服务涌入(2021/11/20 - 10:07:46),当服务请求完成后，由于下游的请求结果的失败率超过了阈值，此时熔断状态被打开，控制台打印的熔断状态开始转为true(2021/11/20 10:07:47)，并一直持续了5秒，直到2021/11/20 10:07:52结束，恢复为false状态
```bash
[root@playground hystrix-demo]# make up
go run ./testcase/cmd/upstream
[GIN-debug] GET    /api/up/v1                --> hystrix-demo/testcase/server.upHandler (4 handlers)
[GIN-debug] Listening and serving HTTP on :9000
2021/11/20 10:07:42 false
2021/11/20 10:07:43 false
2021/11/20 10:07:44 false
2021/11/20 10:07:45 false
2021/11/20 10:07:46 false
[GIN] 2021/11/20 - 10:07:46 | 500 |   42.827448ms |             ::1 | GET      "/api/up/v1"
.
.
# 省略中间98次的请求日志....
.
.
[GIN] 2021/11/20 - 10:07:47 | 500 |    116.3309ms |             ::1 | GET      "/api/up/v1"
2021/11/20 10:07:47 true
2021/11/20 10:07:48 true
2021/11/20 10:07:49 true
[GIN] 2021/11/20 - 10:07:49 | 500 |      47.393µs |             ::1 | GET      "/api/up/v1"
.
.
# 省略中间98次的请求日志....
.
.
[GIN] 2021/11/20 - 10:07:49 | 500 |       2.849µs |             ::1 | GET      "/api/up/v1"
2021/11/20 10:07:50 true
2021/11/20 10:07:51 true
2021/11/20 10:07:52 false
2021/11/20 10:07:53 false
2021/11/20 10:07:54 false
[GIN] 2021/11/20 - 10:07:55 | 500 |   10.016411ms |             ::1 | GET      "/api/up/v1"
.
.
# 省略中间98次的请求日志....
.
.
[GIN] 2021/11/20 - 10:07:55 | 500 |  108.886327ms |             ::1 | GET      "/api/up/v1"
2021/11/20 10:07:55 true
2021/11/20 10:07:56 true
^Csignal: interrupt
make: *** [up] Error 1
```

客户端
可以看到，在第一次开始发起请求的时候拒绝的信息是来自下游服务端的拒绝信息(2021/11/20 10:07:46 - 2021/11/20 10:07:47), 此时还没处于熔断状态。当熔断状态处于开启时(2021/11/20 10:07:47 - 2021/11/20 10:07:52)，所有的拒绝信息(2021/11/20 10:07:49)都是来自于上游服务器，并且可以看到下游服务器并没有打印相关请求信息，可以判断为拦截成功。在熔断状态恢复后，之后的请求信息的结果恢复来自下游服务发送的信息(2021/11/20 10:07:55)

```bash
[root@playground hystrix-demo]# make client
go run ./testcase/cmd/client
2021/11/20 10:07:46 500: reject from downstream
.
.
# 省略中间98次的请求日志....
.
.
2021/11/20 10:07:47 500: reject from downstream
[root@playground hystrix-demo]# make client
go run ./testcase/cmd/client
2021/11/20 10:07:49 500: reject by hystrix
.
.
# 省略中间98次的请求日志....
.
.
2021/11/20 10:07:49 500: reject by hystrix
[root@playground hystrix-demo]# make client
go run ./testcase/cmd/client
2021/11/20 10:07:55 500: reject from downstream
.
.
# 省略中间98次的请求日志....
.
.
2021/11/20 10:07:55 500: reject from downstream
```

## TODO
加入熔断后下游服务的探测功能