# lain-health

一个监控 LAIN 集群状态的 LAIN 应用。

## TODO

- [ ] tinydns

## 架构

在 [lain.yaml](lain.yaml) 里定义了 2 个 proc：server 和 web。server 是一个永不退出
的容器；web 提供 HTTP 服务，当调用 `/api/v1/tinydns_status` 时，web proc 会解析
`server-1` 的 IP 地址，如果可以解析，则返回 `OK`，否则返回 `Down`。然后在
[hagrid](https://github.com/laincloud/hagrid) 里配置 HTTP 报警即可监控 tinydns
状态。

### TinyDNS

- 检测 proc A 是否能解析同一个 App 内的另一个 proc B 的域名
- 验证域名解析是否正确
