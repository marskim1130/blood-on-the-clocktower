# Taro项目结构参考

## 核心目录结构

```
├── babel.config.js             # Babel 配置
├── .eslintrc.js                # ESLint 配置
├── config/                     # 编译配置目录
│   ├── dev.js                  # 开发模式配置
│   ├── index.js                # 默认配置
│   └── prod.js                 # 生产模式配置
├── package.json                # Node.js manifest
├── dist/                       # 打包目录
├── project.config.json         # 小程序项目配置
├── src/                        # 源码目录
│   ├── app.config.js           # 全局配置
│   ├── app.css                 # 全局 CSS
│   ├── app.js                  # 入口组件
│   ├── index.html              # H5 入口 HTML
│   └── pages/                  # 页面组件
│       └── index/
│           ├── index.config.js # 页面配置
│           ├── index.css       # 页面 CSS
│           └── index.jsx       # 页面组件
```

## 关键配置文件

### 1. app.config.js (全局配置)
- 定义页面路由 (`pages`)
- 配置窗口样式 (`window`)
- 设置tabBar等全局配置

### 2. config/ (编译配置)
- `index.js` - 默认配置
- `dev.js` - 开发环境配置
- `prod.js` - 生产环境配置
- 控制编译目标、插件、优化等

### 3. project.config.json (小程序配置)
- 微信小程序项目配置
- appid、项目设置等

## 配置层次关系
1. **全局配置** - app.config.js (应用程序级别)
2. **页面配置** - page.config.js (页面级别，覆盖全局)
3. **编译配置** - config/ (控制编译过程)

## 跨平台支持
- 微信小程序 (主要)
- React Native (移动应用)
- H5 (网页应用)
- 支付宝小程序、百度小程序等

## 错误定位检查点
1. 页面路径是否在 app.config.js 中注册
2. 配置文件语法是否正确
3. 编译配置是否匹配目标平台
4. 依赖是否正确安装

## 常见错误模式
- **页面未找到**: 检查 app.config.js 中的 pages 配置
- **编译错误**: 检查 config/ 目录中的配置
- **平台兼容性**: 检查平台特定的API使用