# GoEdge 调度代码阅读清单

用法：按清单从上到下读。每读完一个条目，在旁边写下“输入、输出、状态变化、失败路径”四件事。

## 0. 开源版边界

- [ ] `EdgeAPI/internal/db/models/node_model_ext_schdule.go`
  - 关注：`HasScheduleSettings()` 在开源版返回 `false`。
  - 结论：更细的节点智能调度不在这份开源代码里。
- [ ] `EdgeNode/internal/nodes/http_request_ln.go`
  - 关注：`!plus` 构建下 LN 二级节点方法都是 stub。
  - 结论：LN/二级节点调度只能从调用点理解边界，不能在开源版读完整实现。

## 1. 架构入口

- [ ] `EdgeAdmin/cmd/edge-admin/main.go`
  - 看：管理台启动、配置、服务安装入口。
- [ ] `EdgeAPI/cmd/edge-api/main.go`
  - 看：API 节点启动、服务注册、后台任务加载。
- [ ] `EdgeNode/cmd/edge-node/main.go`
  - 看：边缘节点启动、daemon、命令参数。
- [ ] `EdgeCommon/pkg/rpc/protos/*.proto`
  - 看：Admin/API/Node 之间的服务契约。

阅读问题：

- 哪些代码是进程入口，哪些代码靠 `init()` 注册？
- EdgeCommon 为什么不能直接依赖 EdgeAPI 或 EdgeNode？
- 控制面和执行面之间传递的是命令、配置、状态还是日志？

## 2. 控制面数据模型

- [ ] `EdgeAPI/internal/db/models/node_cluster_dao.go`
  - 看：`CreateCluster()`、`UpdateCluster()`、`NotifyDNSUpdate()`、`Notify*Update()`。
- [ ] `EdgeAPI/internal/db/models/node_dao.go`
  - 看：节点查询、状态更新、节点配置生成入口。
- [ ] `EdgeAPI/internal/db/models/node_ip_address_dao.go`
  - 看：节点访问 IP、健康状态、上下线计数。
- [ ] `EdgeAPI/internal/db/models/node_region_dao.go`
  - 看：区域启停、排序、可用区域查询。
- [ ] `EdgeAPI/internal/db/models/node_task_dao.go`
  - 看：任务创建、通知、完成状态。

阅读问题：

- 哪些字段影响“能不能被调度”？
- 集群、节点、IP、区域之间是强关联还是配置关联？
- 控制面什么时候创建 node task，什么时候创建 DNS task？

## 3. 配置下发与共享类型

- [ ] `EdgeCommon/pkg/nodeconfigs/node_config.go`
  - 看：边缘节点完整配置结构、缓存读取、clone。
- [ ] `EdgeCommon/pkg/nodeconfigs/node_ip_addr.go`
  - 看：IP 地址可用性结构。
- [ ] `EdgeCommon/pkg/serverconfigs/server_config.go`
  - 看：站点配置和请求处理引用的配置根。
- [ ] `EdgeCommon/pkg/serverconfigs/origin_config.go`
  - 看：源站地址、权重、失败阈值、运行时 `IsOk`。
- [ ] `EdgeCommon/pkg/serverconfigs/network_address_config.go`
  - 看：多地址选择 `PickAddress()`。

阅读问题：

- 哪些配置是持久化配置，哪些字段只用于运行时状态？
- 为什么 `OriginConfig` 可以作为调度候选对象？
- `NodeConfig` 里哪些字段和海外调度直接相关？

## 4. DNS 记录同步

- [ ] `EdgeCommon/pkg/dnsconfigs/cluster_dns_config.go`
  - 看：集群 DNS 同步开关和默认值。
- [ ] `EdgeCommon/pkg/dnsconfigs/record_types.go`
  - 看：支持的 DNS 记录类型。
- [ ] `EdgeAPI/internal/dnsclients/provider_interface.go`
  - 看：DNS provider 的最小能力集合。
- [ ] `EdgeAPI/internal/dnsclients/provider_dnspod.go`
  - 看：腾讯 DNSPod provider 实现方式。
- [ ] `EdgeAPI/internal/dnsclients/provider_tencent_dns.go`
  - 看：腾讯云 DNS provider 实现方式。
- [ ] `EdgeAPI/internal/tasks/dns_task_executor.go`
  - 看：`Start()`、`Loop()`、`doCluster()`、`doDomainWithTask()`、`findDNSManagerWithClusterId()`。

阅读问题：

- 集群 DNS 名称如何变成 A/AAAA/CNAME 记录？
- provider API 的差异被封装在哪一层？
- 记录同步失败后任务如何重试或保留？
- DNS 记录同步和健康检查状态之间哪里产生间接联系？

## 5. 健康检查

- [ ] `EdgeCommon/pkg/serverconfigs/health_check_config.go`
  - 看：探测 URL、间隔、超时、重试、上下线计数。
- [ ] `EdgeAPI/internal/tasks/health_check_task.go`
  - 看：所有集群健康检查任务如何管理。
- [ ] `EdgeAPI/internal/tasks/health_check_cluster_task.go`
  - 看：单集群任务的 ticker、reset、消息聚合。
- [ ] `EdgeAPI/internal/tasks/health_check_executor.go`
  - 看：节点/IP 收集、并发探测、状态更新。
- [ ] `EdgeAPI/internal/tasks/health_check_result.go`
  - 看：探测结果字段。

阅读问题：

- 为什么要区分集群级任务和 executor？
- 失败几次后才自动下线？恢复几次后才上线？
- 自动下线作用在 IP 还是节点？Plus 分支和开源分支有什么差异？
- 健康检查如何避免重复消息轰炸？

## 6. 边缘节点请求链路

- [ ] `EdgeNode/internal/nodes/node.go`
  - 看：节点生命周期和配置加载。
- [ ] `EdgeNode/internal/nodes/listener_http.go`
  - 看：HTTP listener 如何接收请求。
- [ ] `EdgeNode/internal/nodes/listener_tcp.go`
  - 看：TCP 转发如何选择源站。
- [ ] `EdgeNode/internal/nodes/listener_udp.go`
  - 看：UDP 转发如何选择源站。
- [ ] `EdgeNode/internal/nodes/http_request.go`
  - 看：HTTP 请求处理状态机。
- [ ] `EdgeNode/internal/nodes/http_request_reverse_proxy.go`
  - 看：反向代理、源站选择、失败重试。
- [ ] `EdgeNode/internal/nodes/http_client_pool.go`
  - 看：连接池和回源连接复用。

阅读问题：

- 首次回源和失败重试分别调用哪个选择函数？
- `failedOriginIds` 和 `failedLnNodeIds` 如何避免重复选择？
- 哪些 HTTP 状态码会触发重试？哪些不会？
- TCP/UDP 的源站选择和 HTTP 的差异在哪里？

## 7. 源站调度算法

- [ ] `EdgeCommon/pkg/serverconfigs/schedulingconfigs/candidate.go`
  - 看：候选对象必须提供哪些信息。
- [ ] `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling.go`
  - 看：调度接口。
- [ ] `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_random.go`
  - 看：权重随机。
- [ ] `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_round_robin.go`
  - 看：权重轮询。
- [ ] `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_hash.go`
  - 看：按 key hash。
- [ ] `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_sticky.go`
  - 看：cookie/header/query 参数粘滞。
- [ ] `EdgeCommon/pkg/serverconfigs/scheduling_group.go`
  - 看：主备源、`IsOk` 过滤、fallback。
- [ ] `EdgeCommon/pkg/serverconfigs/reverse_proxy_config.go`
  - 看：按域名分组源站、调度初始化、重置。

阅读问题：

- 每个算法是否有运行时状态？是否线程安全？
- 权重如何被归一化或限制？
- hash/sticky 的稳定性依赖什么输入？
- 主源全不可用后如何进入备用源？恢复后如何回切？

## 8. 海外调度对标清单

- [ ] 地域维度：国家、城市、POP、区域、可用区。
- [ ] 网络维度：ISP、AS、跨境链路、骨干质量、丢包/RTT。
- [ ] 成本维度：带宽成本、节点容量、合同约束。
- [ ] 健康维度：探测点、协议、端口、业务 URL、误摘除保护。
- [ ] 切流维度：DNS TTL、权重、灰度、回滚、强制迁移。
- [ ] 观测维度：可用性、命中率、回源失败、调度命中、区域 SLA。

每读完一个 GoEdge 子系统，都把它填到这 6 个维度之一。填不进去的，就是 GoEdge 开源版没有覆盖或你还没有读到的部分。
