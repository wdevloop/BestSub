# Provider Filter Configuration

This document explains how to group checked proxies into different providers using a YAML file. Each provider mirrors the nested fields in `proxy/info/info.go`.

## Basic Structure

```yaml
providers:
  ProviderName:
    output: filename.yaml   # optional, defaults to <ProviderName>.yaml
    lowerbound: 5           # minimum nodes to keep after each rule
    rules:
      - order: speed        # numeric expression for sorting
        desc: true          # descending order
        restriction: country == "US" && alive
      - order: delay
        restriction: delay < 50
```

## Expressions

`restriction` and `order` use the [Expr](https://github.com/expr-lang/expr) language.
All fields of `ProxyInfo` are available directly inside expressions. Several helper
functions are exposed:

- `re(pattern, value)` – regular expression match.
- `size(v)` – length of an array, map or string.
- `sum(v1, v2, ...)` – sum numeric values.
- `map(cond1, val1, cond2, val2, default)` – return the value associated with the first `cond` that is `true`.

A proxy must satisfy all rules in sequence. Each rule sorts the current list using
`order` and filters it by `restriction`.

## Available Filter Fields

The main keys correspond to fields of `ProxyInfo`:

- `country` *(string)* – ISO country code.
- `alive` *(bool)* – whether the node responded to the connectivity test.
- `speed` *(int)* – download speed in KB/s.
- `delay` *(int)* – TCP handshake delay in milliseconds.
- `rate` *(float)* – subscription rate multiplier.
- `risk` *(int)* – overall risk score.
- `unlock` *(map)* – streaming unlock results. Keys include `google`, `chatgpt`, `netflix`, `disney`, `youtube`, `cloudflare`, `tiktok`, `spotify`, `amazon`.
- `ip` *(map)* – IP quality information:
  - `ipUsage` – integer codes (0=home, 1=business, 2=hosting).
  - `ipRisk` – integer scores (0=very low, 1=low, 2=medium, 3=high, 4=very high).
  - `ipRiskFactor` – nested booleans such as `proxy`, `vpn`, `hosting`, `spam`.
  - `ipBanned` – counts of `normal`, `tagged` and `banned` appearances.
- `net` *(map)* – network information:
  - `latency` – latency results with categories `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `International`. Each contains city names and delay in ms.
  - `route` – return-route information, mapping route name to ASN string.

Any field in `ProxyInfo` may be referenced in the filter using this nested style.

## Example

```yaml
providers:
  USFastClean:
    output: us_fast.yaml
    lowerbound: 5
    rules:
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

Set the path to this file with `provider-file` in `config.yaml`. After checks complete, matching proxies are written to the specified provider files and the full results are saved in `results.json`.
