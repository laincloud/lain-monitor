# lain-monitor

一个监控 LAIN 集群状态的 LAIN 应用。

## 监控 TinyDNS

- 检测 proc A 是否能解析同一个 App 内另一个 proc B 的域名
- 验证域名解析是否正确

## 收集集群的已分配内存信息

- 访问 `http://swarm.lain:2376/info` 得到信息

> 在 lain-monitor 的 calico profile 配置中，需要去掉访问 `${LAIN-Nodes-IP-Range}/24:2376` 的限制。

## 编译部署

```
lain reposit ${LAIN-cluster}

# 填写 client-example.json 里的配置信息，然后复制为 client-prod.json
lain secret add -f client-prod.json ${LAIN-cluster} web /lain/app/client-prod.json

lain build
lain tag ${LAIN-cluster}
lain push ${LAIN-cluster}
lain deploy ${LAIN-cluster}
```
