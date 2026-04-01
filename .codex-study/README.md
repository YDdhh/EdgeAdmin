# GoEdge 学习链路说明

我已将官方源码拉到当前工作区用于联读：

- `.codex-study/EdgeCommon`
- `.codex-study/EdgeAPI`
- `.codex-study/EdgeNode`

推荐用 3 层视角理解 GoEdge 调度链路：

1. `EdgeAdmin`：配置入口 + RPC 调用。
2. `EdgeCommon`：配置结构 + 调度算法。
3. `EdgeAPI/EdgeNode`：真正执行健康检查、DNS 同步、请求调度。
