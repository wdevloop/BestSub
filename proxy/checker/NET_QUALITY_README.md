# 网络质量检测功能说明

## 功能概览

本模块实现了对NetInfo结构中网络延迟和路由线路检测的完整功能，包括：

1. **三网TCP大包延迟检测**
   - 中国电信 (ChinaTelecom)
   - 中国联通 (ChinaUnicom) 
   - 中国移动 (ChinaMobile)

2. **国际线路延迟检测**
   - 基于iperf3服务器的延迟测试
   - 覆盖美国、英国、德国、日本、新加坡等主要地区

3. **回程路由线路检测**
   - 支持nexttrace和mtr工具进行路由跟踪
   - 自动识别CN2 GIA、CN2 GT、163、4837、CMI等线路类型
   - 支持TCP和UDP协议的路由检测

## 核心文件结构

```
proxy/checker/
├── net_quality_checker.go    # 主要的网络质量检测逻辑
├── resource_manager.go       # 资源文件管理（已增强）
├── net/
│   ├── iperf_targets.go     # iperf测试目标配置
│   └── speedtest_targets.go # 速度测试目标配置
├── net_quality_test.go      # 测试文件
└── NET_QUALITY_README.md    # 本说明文档
```

## 数据结构映射

### NetLatencyInfo
```go
type NetLatencyInfo struct {
    ChinaTelecom  map[string]uint16  // 电信延迟 "BJ-CT": 50ms
    ChinaUnicom   map[string]uint16  // 联通延迟 "SH-CU": 45ms  
    ChinaMobile   map[string]uint16  // 移动延迟 "GD-CM": 60ms
    International map[string]uint16  // 国际延迟 "US-LosAngeles": 180ms
}
```

### Route信息
```go
Route map[string]string  // "BJ-CT-TCP": "CN2GIA", "SH-CU-UDP": "4837"
```

## 主要功能实现

### 1. 延迟测试
- **三网延迟**: 针对北京、上海、广东三个主要地区的电信、联通、移动网络进行TCP连接延迟测试
- **国际延迟**: 测试到美国、欧洲、亚太等地区主要城市的延迟
- **超时控制**: 每个测试都有合理的超时设置，避免阻塞

### 2. 路由检测
- **工具支持**: 
  - 优先使用nexttrace进行路由跟踪
  - 回退使用mtr工具
  - 最后使用基础IP段判断
- **线路识别**: 基于ASN路径自动识别线路类型
  - CN2 GIA (AS4809)
  - CN2 GT (AS4134 + AS4809/23764)
  - 电信163 (AS4134)
  - 联通4837 (AS4837) 
  - 移动CMI (AS9808/58453)

### 3. 资源管理
- **自动下载**: 运行时自动下载所需的配置文件
- **本地缓存**: 资源文件缓存在系统临时目录
- **错误处理**: 优雅处理资源文件缺失的情况

## 配置文件

### 必需的资源文件
- `province.json`: 省份代码映射
- `iperf.json`: iperf3测试服务器列表
- `speedtest_cn.json`: 中国区速度测试服务器
- `AS_Mapping.txt`: ASN映射文件

## 使用方法

### 在Checker中调用
```go
func (c *Checker) CheckAll() {
    // 其他检测...
    
    // 网络质量检测
    if err := c.CheckNetworkQuality(); err != nil {
        log.Debug("网络质量检测完成，部分测试可能失败: %v", err)
    }
}
```

### 手动调用
```go
checker := NewChecker(proxy)
err := checker.CheckNetworkQuality()
```

## 测试结果示例

### 延迟测试结果
```
Latency.ChinaTelecom["BJ-CT"] = 45     // 北京电信45ms
Latency.ChinaUnicom["SH-CU"] = 38      // 上海联通38ms  
Latency.ChinaMobile["GD-CM"] = 52      // 广东移动52ms
Latency.International["US-LosAngeles"] = 180  // 洛杉矶180ms
```

### 路由检测结果
```
Route["BJ-CT-TCP"] = "CN2GIA"         // 北京电信TCP走CN2 GIA
Route["SH-CU-TCP"] = "4837"           // 上海联通TCP走4837
Route["GD-CM-UDP"] = "CMI"            // 广东移动UDP走CMI
```

## 性能优化

1. **并发控制**: 避免同时进行过多测试
2. **超时管理**: 合理的超时设置
3. **错误处理**: 单个测试失败不影响整体检测
4. **资源缓存**: 资源文件本地缓存避免重复下载

## 扩展性

- 易于添加新的测试目标
- 支持自定义路由识别规则
- 可配置的测试参数
- 模块化设计便于维护

## 注意事项

1. **网络依赖**: 需要良好的网络连接来下载资源文件
2. **工具依赖**: 路由检测功能依赖nexttrace或mtr工具
3. **权限要求**: 某些路由检测功能可能需要管理员权限
4. **代理兼容**: 确保代理支持TCP连接和DNS解析 