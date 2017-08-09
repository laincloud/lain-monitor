# tinydns-healthcheck

一个监控 [tinydns](https://github.com/laincloud/tinydns) 状态的 LAIN 应用。

## 背景

在 LAIN 集群里，tinydns 负责应用内部各 proc 之间、service 和 resource 的域名解析，
具有很重要的作用。如果 tinydns 挂掉，需要报警。tinydns-healthcheck 提供了一个
HTTP 服务，返回 tinydns 的工作状态。

## 架构

在 [lain.yaml](lain.yaml) 里定义了 2 个 proc：server 和 web。server 是一个永不退出
的容器；web 提供 HTTP 服务，当调用 `/api/v1/tinydns_status` 时，web proc 会解析
`server-1` 的 IP 地址，如果可以解析，则返回 `OK`，否则返回 `Down`。然后在
[hagrid](https://github.com/laincloud/hagrid) 里配置 HTTP 报警即可监控 tinydns
状态。
