# Provider 配置说明

本文档介绍如何通过 YAML 配置文件按检测结果对节点进行分组。每个 provider 的字段与 `proxy/info/info.go` 中的 `ProxyInfo` 结构保持一致。

## 基本结构

```yaml
providers:
  Provider名称:
    output: 文件名.yaml   # 可选，默认为 Provider 名称 + .yaml
    lowerbound: 5         # 每条规则筛选后的最少数量
    # 简单写法：仅包含一组规则
    rules:
      - order: speed      # 用于排序的数值表达式
        desc: true        # 是否降序
        restriction: country == "US" && alive
      - order: delay
        restriction: delay < 50
    # 进阶写法：多组规则依次执行
    # ruleSets:
    #   - rules:
    #       - order: ...
    #       - ...
```

## 表达式说明

`restriction` 和 `order` 使用 [Expr](https://github.com/expr-lang/expr) 语法，可直接引用 `ProxyInfo` 中的字段，并内置以下辅助函数：

- `re(pattern, value)` – 正则匹配。
- `size(v)` – 求数组、map 或字符串的长度。
- `sum(v1, v2, ...)` – 将多个数值相加。
- `map(cond1, val1, cond2, val2, default)` – 按条件返回对应的值。

规则按顺序依次执行，每条规则会先根据 `order` 排序，再用 `restriction` 过滤结果。

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
    lowerbound: 5
    ruleSets:
      - rules:
          - order: speed
            desc: true
            restriction: re("^US$", country) && alive
          - order: delay
            restriction: delay < 50
  CNQuality:
    output: cn_quality.yaml
    rules:
      - order: map(speed >= 20000, 1, speed >= 10000, 2, 3)
        desc: false
        restriction: country == "CN" && speed > 2048
```

在 `config.yaml` 中通过 `provider-file` 选项指定此文件路径。程序运行结束后，会按配置生成各个 provider 文件，完整的检测结果存储在 `results.json` 中。
