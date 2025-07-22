# CacheShield - 防止缓存击穿的 Go 库

CacheShield 通过分布式锁确保只有一个请求执行数据重建，其他请求自动等待结果，彻底解决雪崩式穿透问题。

### ✨ 核心特性
1. 🚦 分布式锁控制 - 基于 Redis 的互斥锁保证单实例重建缓存
2. 🔄 重试机制 - 等待期间自动重试读取缓存（可配置策略）
3. 🧩 多版本兼容 - 支持 go-redis v6 到 v9

### 🚀 快速开始

```
    c := cacheshield.NewV7()
    c.LoadOrStore()
```