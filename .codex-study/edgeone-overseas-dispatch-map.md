# GoEdge 机制到 EdgeOne 海外调度的对照

这份文档用于把 GoEdge 的可见实现映射到海外 CDN 调度常见能力。它不是腾讯内部实现说明，只是帮助你建立阅读开源项目时的业务参照系。

## 总览

| 海外调度关注点 | GoEdge 可读入口 | 能学到什么 | 开源版缺口 |
| --- | --- | --- | --- |
| 控制面配置 | `EdgeAPI/internal/db/models/*` | 集群、节点、IP、区域、任务如何建模 | 缺少真实海外 POP/线路/容量策略 |
| DNS 调度 | `EdgeAPI/internal/tasks/dns_task_executor.go` | A/AAAA/CNAME、TTL、provider 同步 | 缺少完整智能线路解析和流量权重体系 |
| 健康检查 | `EdgeAPI/internal/tasks/health_check_executor.go` | 探测、上下线计数、自动下线、消息 | 探测点分布和多维质量评分较弱 |
| 节点内源站调度 | `EdgeCommon/pkg/serverconfigs/schedulingconfigs/*` | random/round-robin/hash/sticky 和主备源 | 这是回源级调度，不是全局用户到 POP 调度 |
| 边缘执行面 | `EdgeNode/internal/nodes/http_request_reverse_proxy.go` | 请求进入节点后的重试、摘除、恢复 | LN 二级节点在开源版是 stub |
| 观测与反馈 | `EdgeAPI/internal/db/models/*stat*`、`EdgeNode/internal/stats/*` | 流量统计和部分运行状态 | 缺少完整实时质量闭环和调度决策闭环 |

## 可以直接迁移的思维模型

1. 控制面和执行面分离  
   控制面负责“配置、任务、状态、DNS”，执行面负责“请求处理、源站选择、失败重试”。海外调度同样需要把策略计算和边缘执行分层。

2. 可用性状态要分层  
   GoEdge 里至少有集群、节点、节点 IP、源站这几层状态。海外调度里还会增加 POP、线路、国家/地区、ISP、AS、探测点等层级。

3. DNS 切流和节点内重试是两种不同调度  
   DNS/GSLB 决定用户先到哪个边缘入口，节点内源站调度决定请求到达节点后怎么回源。不要把两者混在一起。

4. 健康检查不能只看一次失败  
   `CountTries`、`CountUp`、`CountDown`、`TryDelay` 体现了抗抖动思路。海外场景还要考虑跨境链路抖动、区域性故障、探测点偏差。

5. 调度算法需要候选对象接口  
   `CandidateInterface` 把算法和候选对象解耦。业务上可以把候选对象从 origin 扩展为 POP、IP 池、线路、机房、区域等。

## 对标专题

### 1. DNS/GSLB 层

GoEdge 当前主线：

- 集群配置 DNS 名称。
- 根据节点访问 IP 同步 A/AAAA。
- 根据业务站点同步 CNAME。
- 通过 provider API 拉取和更新记录。

海外调度需要额外考虑：

- 递归 DNS 出口识别和 EDNS Client Subnet。
- 国家/地区、ISP、AS、城市级线路。
- 权重、容量、成本、故障状态共同决策。
- TTL 与切流速度、缓存命中、权威 DNS 压力的权衡。

建议练习：

- 画出 GoEdge `doCluster()` 如何生成记录。
- 设计一个伪字段：`routeKey = country + isp + asn`，思考它应该放在 DNS route、node IP 还是独立策略表里。

### 2. 健康检查层

GoEdge 当前主线：

- 主 API 节点定期启动集群健康检查。
- 收集集群下所有可访问节点 IP。
- 并发请求健康检查 URL。
- 更新 IP 或节点上下线计数。
- 触发消息、阈值、动作。

海外调度需要额外考虑：

- 多探测点、多协议、多端口、多 URL。
- 探测点自身故障和误报过滤。
- 按地区/ISP 观测质量，而不是全局二值可用。
- 探测结果如何进入 DNS/GSLB 权重或摘除策略。

建议练习：

- 把 `HealthCheckResult` 扩展成包含 `probeRegion`、`rttMs`、`loss` 的伪结构。
- 思考 `AutoDown` 应该按 IP、节点、POP 还是线路生效。

### 3. 边缘节点回源层

GoEdge 当前主线：

- `ReverseProxyConfig` 保存主备源和调度算法。
- `NextOrigin()` 正常选择源站。
- `AnyOrigin()` 在失败重试时避开失败源站。
- `OriginStateManager` 记录失败并定期探测恢复。
- 源站恢复后重置调度。

海外调度需要额外考虑：

- 源站地域和边缘 POP 的距离。
- 源站健康、回源链路质量、连接池压力。
- 多级回源、父节点、区域缓存层。
- 客户配置、成本和合规限制。

建议练习：

- 对比 `random`、`roundRobin`、`hash`、`sticky` 在回源场景中的适用性。
- 设计一个“优先同区域源站，失败后跨区域”的伪算法。

### 4. 数据闭环层

GoEdge 可见线索：

- `EdgeNode/internal/stats/*`
- `EdgeAPI/internal/db/models/stats/*`
- `node_value`、`node_threshold`、`node_log`

海外调度需要额外考虑：

- 实时指标：RTT、丢包、可用性、错误率、回源失败、带宽、连接数。
- 离线指标：成本、容量、水位、SLA、历史质量。
- 策略变更审计：谁切流、切了多少、为什么、何时回滚。
- 决策解释：某个用户为何被调度到某个 POP/IP。

建议练习：

- 为每个调度决策写一条“解释日志”的字段清单。
- 思考这些字段来自控制面、DNS 层、节点层还是探测系统。

## 最小扩展思考题

如果只允许在 GoEdge 上做一个学习型扩展，不写生产级功能，优先做：

1. 新增一个只在测试里使用的 `LatencyAwareScheduling`。
2. 候选源站包含 `Weight` 和伪造的 `RTT`。
3. 选择分数为 `score = weight / max(rtt, 1)`。
4. 写单测验证低延迟高权重源站更容易被选中。

这个练习能把 Go 接口、调度算法、测试、业务指标都串起来。
