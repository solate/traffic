# traffic

流量统计系统学习示例


## 系统

### 结构

* web : 需要收集的模拟网站 8080
* 打点服务器: nginx 8000 , 记录日志
* run 生成日志程序，模拟大量用户访问
* 分析服务：

### 流量统计系统

* 用户流量
* 流量漏斗
* 用户增长和Growth Hacking



## 笔记

### 统计js

在`/views/js/statistics.js`里写需要统计的js代码，如客户端时间，url等

上报用户访问信息，将访问数据上传打点服务器。

后续查资料js锁如何使用，放置上报数量出现错误。防止上传了很多倍，数量级上的上传错误


### 打点服务器(nginx)

使用nginx作为打点服务器

* nginx 高性能webserver服务器 : 模块 ngx_http_empty_git_module
* nginx 借助 access.log 记录打点请求: 性能开销最小，最佳方案


打点服务，接受请求，返回json。如果使用通常接口，返回一串字符串或状态码，会很长，在流量高峰时压力会很大。
在使用ngx_http_empty_git_module返回这个小的图片只有236B, 非常小。

打点服务器只需要上报，所以返回需要尽量的小。这种方式用在高并发，并不需要什么返回的场景。

nginx 配置

```
    server {
        //statistics.js 中写的 http://localhost:8000/dig
        listen       8000;

        //使用模块的配置
    	location = /dig {
            empty_gif;
            error_page 405 =200 $request_uri; //405不允许访问改成允许
    	}

```

该模块使用c拼出一个1*1的gif.

这里有个跨域的问题，暂时没管。


access_log打开，将main的格式化打开，指定生成日志到dig.log

```
 log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';


 access_log  logs/dig.log  main;
```