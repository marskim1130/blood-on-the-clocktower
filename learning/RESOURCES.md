# 学习资源

## 官方文档
- **Taro官方文档**: https://taro-docs.jd.com/taro/docs/
  - 核心概念、API参考、配置指南
  - 权威性: 高

## 项目相关
- **Blood on the Clocktower项目**: 当前工作目录
  - Taro 4.x + React实现
  - 实际案例，可直接学习

## Go 与后端知识
- **A Tour of Go**: https://go.dev/tour/
  - Go 的包、结构体、方法、接口、错误处理和并发入门；适合第一次系统学习 Go。
- **Effective Go**: https://go.dev/doc/effective_go
  - Go 惯用写法；适合阅读本项目的接口、错误和并发代码时查阅。
- **Go `net/http` package**: https://pkg.go.dev/net/http
  - HTTP 服务与请求处理 API；对应 `cmd/server/main.go` 的服务启动入口。
- **Go `context` package**: https://pkg.go.dev/context
  - 取消、截止时间和跨函数传递请求范围；对应存储接口的 `ctx` 参数。
- **Go `sync` package**: https://pkg.go.dev/sync
  - 互斥锁、条件变量等同步原语；对应房间串行提交和出站队列。

## 学习路径
1. **基础概念**: Taro项目结构、配置文件、编译流程
2. **核心API**: 路由、网络请求、存储、设备API
3. **组件系统**: 内置组件、自定义组件、跨平台适配
4. **调试技巧**: 错误定位、日志分析、性能优化

## 参考资源
- Taro GitHub仓库: https://github.com/nervjs/taro
- Taro社区: https://taro-club.jd.com/
- 微信小程序开发文档: https://developers.weixin.qq.com/miniprogram/dev/framework/
