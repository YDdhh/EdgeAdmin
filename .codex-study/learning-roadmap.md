# GoEdge CDN 调度方向 6 周学习路线

## 目标

以调度为主线学习 GoEdge：从配置入口、控制面任务、DNS 记录同步、健康检查、边缘节点请求调度，一直读到源站选择和故障摘除。Go 语言学习穿插在真实代码里完成，重点关注模块组织、接口、goroutine、ticker、gRPC、错误处理和测试。

这不是完整复刻 EdgeOne 海外调度的路线。GoEdge 开源版更适合学习 CDN 调度骨架：控制面如何管理节点、DNS 如何同步、健康检查如何影响可用性、边缘节点如何做源站级调度。

## 第 1 周：Go 工程和整体架构

阅读入口：

- `EdgeAdmin/go.mod`
- `EdgeAdmin/cmd/edge-admin/main.go`
- `EdgeAPI/go.mod`
- `EdgeAPI/cmd/*/main.go`
- `EdgeNode/go.mod`
- `EdgeNode/cmd/edge-node/main.go`
- `EdgeCommon/pkg/rpc/protos/*.proto`

要回答的问题：

- `EdgeAdmin`、`EdgeAPI`、`EdgeNode`、`EdgeCommon` 分别承担什么职责？
- `go.mod` 里的 `replace github.com/TeaOSLab/EdgeCommon => ../EdgeCommon` 解决了什么问题？
- 后台任务是如何通过 `init()`、事件和 goroutine 启动的？
- gRPC proto 在控制面和节点之间承担什么契约？

Go 语言重点：

- Go module、包路径、`internal` 可见性。
- `init()` 的副作用和注册式初始化。
- `context.Context`、`error` 返回、结构体方法。
- goroutine、ticker、defer 的常见用法。

验收产物：

- 一张三仓关系图。
- 一页“GoEdge 控制面/执行面/共享协议”笔记。
- 一个问题清单：哪些调度能力在开源版可见，哪些是 stub 或 Plus 能力。

## 第 2 周：节点、集群、区域模型

阅读入口：

- `EdgeAPI/internal/db/models/node_cluster_dao.go`
- `EdgeAPI/internal/db/models/node_dao.go`
- `EdgeAPI/internal/db/models/node_region_dao.go`
- `EdgeAPI/internal/db/models/node_ip_address_dao.go`
- `EdgeAPI/internal/db/models/node_task_dao.go`
- `EdgeCommon/pkg/nodeconfigs/node_config.go`
- `EdgeCommon/pkg/nodeconfigs/node_ip_addr.go`

要回答的问题：

- 集群、节点、区域、IP 地址在数据库模型里如何关联？
- 节点的 `isOn`、`isUp`、IP 的 `isOn`/`isUp` 各自表达什么层级的可用性？
- `CreateCluster()` 默认写入哪些 DNS 配置？
- `Notify*Update()` 如何把配置变化转成节点任务？
- `NodeConfig` 是如何承载下发到边缘节点的完整配置？

Go 语言重点：

- DAO 方法组织方式。
- JSON 字段和结构体 tag。
- 指针接收者、共享单例 DAO。
- 大结构体如何分层管理配置。

验收产物：

- 一张“集群 -> 节点 -> IP -> 任务”的数据关系图。
- 一页节点状态字段说明。
- 一个 EdgeOne 对照点：海外节点、区域、线路、POP、IP 池这些概念如何映射到这个模型。

## 第 3 周：DNS 调度与记录同步

阅读入口：

- `EdgeCommon/pkg/dnsconfigs/cluster_dns_config.go`
- `EdgeCommon/pkg/dnsconfigs/record_types.go`
- `EdgeAPI/internal/tasks/dns_task_executor.go`
- `EdgeAPI/internal/db/models/dns/dns_task_dao.go`
- `EdgeAPI/internal/db/models/dns/dns_domain_dao.go`
- `EdgeAPI/internal/dnsclients/provider_interface.go`
- `EdgeAPI/internal/dnsclients/provider_*.go`

要回答的问题：

- 集群域名、业务 CNAME、节点 A/AAAA 记录分别由谁生成和同步？
- `NodesAutoSync`、`ServersAutoSync`、`CNAMEAsDomain`、`IncludingLnNodes` 的含义是什么？
- DNS provider 抽象支持哪些操作？
- DNS 任务如何保证一次任务执行完成后更新状态？
- TTL 对故障切流速度有什么影响？

Go 语言重点：

- 接口抽象和 provider 多态。
- 长函数阅读：先按 task type 拆主线，再读分支。
- JSON encode/decode 与外部 API 数据转换。

验收产物：

- 一张“配置变更 -> DNS task -> provider API -> records cache”的链路图。
- 一页 DNS 记录同步笔记。
- 一个 EdgeOne 对照点：GSLB、权威 DNS、线路解析、TTL、灰度切流如何扩展这个模型。

## 第 4 周：健康检查和故障摘除

阅读入口：

- `EdgeAPI/internal/tasks/health_check_task.go`
- `EdgeAPI/internal/tasks/health_check_cluster_task.go`
- `EdgeAPI/internal/tasks/health_check_executor.go`
- `EdgeAPI/internal/tasks/health_check_result.go`
- `EdgeCommon/pkg/serverconfigs/health_check_config.go`
- `EdgeAPI/internal/db/models/node_ip_address_dao.go`
- `EdgeAPI/internal/db/models/node_threshold_dao.go`

要回答的问题：

- 健康检查任务如何按集群启动、停止、重置配置？
- `HealthCheckExecutor.Run()` 如何收集待检查节点和 IP？
- `CountTries`、`TryDelay`、`CountUp`、`CountDown`、`AutoDown` 分别影响什么？
- 健康检查失败后，是修改 IP 状态还是节点状态？
- 消息、阈值和节点动作在故障路径里分别何时触发？

Go 语言重点：

- ticker 任务循环。
- goroutine + channel + wait group 并发检查。
- 错误处理和延迟重试。
- 用测试定位业务入口。

验收产物：

- 一张“探测失败 -> 状态计数 -> 自动下线 -> 消息/动作”的链路图。
- 一页健康检查字段表。
- 一个 EdgeOne 对照点：海外调度探测点、拨测维度、误摘除保护、恢复策略如何设计。

## 第 5 周：边缘节点请求调度和回源

阅读入口：

- `EdgeNode/internal/nodes/node.go`
- `EdgeNode/internal/nodes/listener_http.go`
- `EdgeNode/internal/nodes/listener_tcp.go`
- `EdgeNode/internal/nodes/listener_udp.go`
- `EdgeNode/internal/nodes/http_request.go`
- `EdgeNode/internal/nodes/http_request_reverse_proxy.go`
- `EdgeNode/internal/nodes/origin_utils.go`
- `EdgeNode/internal/nodes/origin_state_manager.go`
- `EdgeNode/internal/nodes/http_client_pool.go`

要回答的问题：

- HTTP/TCP/UDP 请求分别从哪里进入节点？
- HTTP 请求经过哪些处理阶段后进入反向代理？
- `NextOrigin()` 和 `AnyOrigin()` 在首次选择和失败重试时如何切换？
- 源站连接失败后如何记录状态并触发调度重置？
- `OriginStateManager` 如何定期探测恢复？

Go 语言重点：

- `net/http`、`net.Conn`、TLS dial。
- 请求对象作为状态机推进。
- 并发安全：锁、共享状态和运行时状态。
- 连接池和超时配置。

验收产物：

- 一张“用户请求 -> EdgeNode -> reverse proxy -> origin”的请求链路图。
- 一页回源失败重试笔记。
- 一个 EdgeOne 对照点：边缘节点内局部调度和全局流量调度的边界。

## 第 6 周：调度算法专题和业务对标

阅读入口：

- `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling.go`
- `EdgeCommon/pkg/serverconfigs/schedulingconfigs/candidate.go`
- `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_random.go`
- `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_round_robin.go`
- `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_hash.go`
- `EdgeCommon/pkg/serverconfigs/schedulingconfigs/scheduling_sticky.go`
- `EdgeCommon/pkg/serverconfigs/scheduling_group.go`
- `EdgeCommon/pkg/serverconfigs/reverse_proxy_config.go`
- `EdgeCommon/pkg/serverconfigs/origin_config.go`

要回答的问题：

- `SchedulingInterface` 和 `CandidateInterface` 如何解耦算法和候选对象？
- random、round-robin、hash、sticky 的输入、状态、适用场景分别是什么？
- `SchedulingGroup` 如何处理主备源站和 `IsOk` 过滤？
- `ReverseProxyConfig` 如何按域名分组源站？
- 如果要做海外 GSLB，需要在哪些层新增地域、ISP、链路质量和成本信号？

Go 语言重点：

- 小接口设计。
- 算法状态初始化和运行期选择。
- 单测作为算法说明书。
- 用最小 demo 验证算法行为。

验收产物：

- 一张四种调度算法对比表。
- 一页“GoEdge 调度能力 vs EdgeOne 海外调度能力”差距表。
- 一个小实验：新增一个只读 demo 或测试，构造 3 个 origin，验证某个算法的选择结果。

## 每周固定节奏

1. 先读数据结构和配置对象。
2. 再读入口函数和正常路径。
3. 最后读失败路径、重试、状态变更和测试。
4. 用 `weekly-note-template.md` 记录：链路图、关键类型、调用链、疑问、EdgeOne 对照。
5. 每周只深入一个主链路，WAF、证书、缓存只作为旁路理解。

## 建议测试命令

在 `EdgeAPI` 下：

```powershell
go test ./internal/tasks -run DNS
go test ./internal/tasks -run HealthCheck
go test ./internal/dnsclients/...
```

在 `EdgeCommon` 下：

```powershell
go test ./pkg/serverconfigs/schedulingconfigs ./pkg/serverconfigs -run Scheduling
```

在 `EdgeNode` 下：

```powershell
go test ./internal/nodes -run "Origin|Reverse|Listener"
```

这些测试可能依赖本地配置、数据库或平台能力。学习阶段遇到失败时，先记录失败依赖和入口，不急着修。
