# Provider 配置说明

本文档介绍如何通过 YAML 配置文件按检测结果对节点进行分组。每个 provider 的字段与 `proxy/info/info.go` 中的 `ProxyInfo` 结构保持一致。

## 基本结构

```yaml
providers:
  Provider名称:
    output: 文件名.yaml   # 可选，默认为 Provider 名称 + .yaml
    filter:
      ...
```

## 编写规则

- **字符串**：使用正则表达式匹配。
- **数字**：支持 `>=`、`<=`、`>`、`<`、`[min,max]` 或单个数字。
- **布尔值**：直接写 `true` 或 `false`。
- **map/结构体**：在子项中继续编写约束。

只有所有条件同时满足时，节点才会被划入对应的 provider。

## 可用的过滤字段

主要键与 `ProxyInfo` 中的字段对应：

- `country` *(string)*：国家代码。
- `alive` *(bool)*：节点是否在存活检测中成功连接。
- `speed` *(int)*：下载速度，单位 KB/s。
- `delay` *(int)*：TCP 握手延迟，单位毫秒。
- `rate` *(float)*：订阅倍率。
- `risk` *(int)*：综合风险分数。
- `unlock` *(map)*：流媒体解锁情况，可用键有 `google`、`chatgpt`、`netflix`、`disney`、`youtube`、`cloudflare`、`tiktok`、`spotify`、`amazon`。
- `ip` *(map)*：IP 纯净度相关信息：
  - `ipUsage`：整数类型 (0=住宅，1=商业，2=托管)。
  - `ipRisk`：整数评分 (0=极低，1=低，2=中，3=高，4=极高)。
  - `ipRiskFactor`：布尔值字段，如 `proxy`、`vpn`、`hosting`、`spam` 等。
  - `ipBanned`：`normal`、`tagged`、`banned` 的出现次数。
- `net` *(map)*：网络质量相关信息：
  - `latency`：延迟结果，包含 `ChinaTelecom`、`ChinaUnicom`、`ChinaMobile`、`International` 等分类，每个分类下为城市名到毫秒值的映射。
  - `route`：回程路由信息，键为线路名称，值为 AS 号。

只要在 `ProxyInfo` 中存在的字段，都可以用相同的嵌套写法在过滤器中引用。

## 示例

```yaml
providers:
  USFastClean:
    output: us_fast.yaml
    filter:
      country: "^US$"
      alive: true
      speed: ">=10240"
      delay: "<=50"
      rate: "[0,1.5]"
      unlock:
        netflix: true
        disney: true
        youtube: true
        chatgpt: true
      ip:
        ipUsage:
          ipInfo: "[0,1]"
        ipRisk:
          ipqs: "<=1"
          dbip: "<=2"
        ipRiskFactor:
          ip2Location:
            hosting: false
        ipBanned:
          banned: 0
      net:
        latency:
          International:
            Tokyo: "<=70"
          ChinaTelecom:
            Shanghai: "<=150"
        route:
          Beijing-ChinaUnicom-TCP: "AS4837"
  CNQuality:
    filter:
      country: "^CN$"
      speed: ">2048"
      unlock:
        tiktok: true
      ip:
        ipBanned:
          banned: 0
        ipRisk:
          ipqs: "<=2"
```

在 `config.yaml` 中通过 `provider-file` 选项指定此文件路径。程序运行结束后，会按配置生成各个 provider 文件，完整的检测结果存储在 `results.json` 中。
