# Provider Filter Configuration

This document explains how to group checked proxies into different providers using a YAML file. Each provider mirrors the nested fields in `proxy/info/info.go`.

## Basic Structure

```yaml
providers:
  ProviderName:
    output: filename.yaml   # optional, defaults to <ProviderName>.yaml
    filter:
      ...
```

## Constraint Rules

- **String fields** use regular expressions.
- **Numeric fields** support `>=`, `<=`, `>`, `<`, an inclusive `[min,max]` range or a single value.
- **Boolean fields** match `true` or `false`.
- **Map or struct fields** continue with nested keys.

A proxy matches a provider only when all specified constraints are satisfied.

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
    output: cn_quality.yaml
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

Set the path to this file with `provider-file` in `config.yaml`. After checks complete, matching proxies are written to the specified provider files and the full results are saved in `results.json`.
