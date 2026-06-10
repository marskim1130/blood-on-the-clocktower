# Work Log

## 2026-06-10 14:27 — 修复 TypeScript 类型检查错误和 ESLint 配置

### 问题
1. TypeScript 类型检查有多个错误：
   - 未使用的导入：`GamePhase`, `Player`, `PlayerId`
   - 未使用的变量：`initialState`, `ProtoCharacter`
   - 类型错误：`GameId` 未定义
   - 可选链警告：`Object is possibly 'undefined'`
   - 联合类型访问错误：`GameServerEvent` 属性访问问题

2. ESLint 配置缺失

### 解决方案
1. **修复 `packages/core/src/state-machine/index.ts`**：
   - 移除未使用的导入：`GamePhase`, `Player`, `PlayerId`
   - 添加缺失的导入：`GameId`
   - 将 `initialState` 重命名为 `defaultState` 并使用它

2. **修复 `packages/core/src/types/generated/__tests__/proto-types.test.ts`**：
   - 移除未使用的导入：`ProtoCharacter`

3. **修复 `packages/core/src/websocket/__tests__/websocket-client.test.ts`**：
   - 使用非空断言 `!` 修复可选链警告

4. **修复 `packages/frontend/src/lib/taro-websocket-transport.test.ts`**：
   - 使用类型断言修复 `resolveTask` 调用问题

5. **修复 `packages/frontend/src/pages/index/index.tsx`**：
   - 使用 `in` 操作符进行类型守卫，正确访问 `GameServerEvent` 联合类型的属性

6. **修复 `packages/core/tsconfig.json`**：
   - 添加 `"composite": true` 以支持项目引用

7. **创建 `.eslintrc.json` 配置文件**：
   - 配置 TypeScript ESLint 规则
   - 设置 `no-console` 为 `off`

8. **修复 `packages/core/src/websocket/index.ts`**：
   - 将 `console.error` 改为 `console.warn`

### 修改文件
- `packages/core/src/state-machine/index.ts`
- `packages/core/src/types/generated/__tests__/proto-types.test.ts`
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`
- `packages/frontend/src/lib/taro-websocket-transport.test.ts`
- `packages/frontend/src/pages/index/index.tsx`
- `packages/core/tsconfig.json`
- `.eslintrc.json`
- `packages/core/src/websocket/index.ts`

### 测试结果
- ✅ TypeScript 类型检查：通过
- ✅ ESLint 检查：通过（0 错误，0 警告）
- ✅ 单元测试：13 个测试全部通过
- ✅ 后端 Go 测试：通过
- ✅ 完整构建：成功

### 撤回方式
```bash
git checkout -- packages/core/src/state-machine/index.ts
git checkout -- packages/core/src/types/generated/__tests__/proto-types.test.ts
git checkout -- packages/core/src/websocket/__tests__/websocket-client.test.ts
git checkout -- packages/frontend/src/lib/taro-websocket-transport.test.ts
git checkout -- packages/frontend/src/pages/index/index.tsx
git checkout -- packages/core/tsconfig.json
rm .eslintrc.json
git checkout -- packages/core/src/websocket/index.ts
```

---

## 2026-06-10 10:53 — 启动项目服务

### 操作
1. 启动后端 Go 服务器（端口 8080）
2. 启动前端 Taro 开发服务器（watch 模式）

### 服务状态
- 后端服务器：✅ 运行中 (`:8080`)
- 前端开发服务器：✅ 运行中 (watch 模式)

### 验证
- 健康检查：`curl http://localhost:8080/health` → 返回 `ok`
- 前端编译：成功（2.50s）

---

## 2026-05-30 15:10 — 前端状态机第一阶段初始化
### 问题
需要引导用户学习前端的 Branded Types 模式以及 Zustand 状态管理设计。
### 解决方案
1. 切换工作区至第一个 Git 提交（哈希 `17fe71b9031fc051176aa78775ca5b8aab8e8536`）。
2. 在根目录下重新建立教学审计日志，开启 TypeScript 前端核心库的第一阶段教学。
### 修改文件
- `work.md` (重建)
### 撤回方式
```bash
git checkout main
```

---

## 2026-05-30 15:11 — 修复 state-machine/index.ts 的 GameId 漏导 Bug
### 问题
Zustand 初始化状态机时对 id 字段使用了 `'' as GameId` 类型转换，但第 2 行的 import 语句漏导了 `GameId` 类型，导致 TypeScript 编译器报错 `Cannot find name 'GameId'`。
### 解决方案
在 `packages/core/src/state-machine/index.ts` 导入声明中加上 `GameId` 类型。
### 修改文件
- `packages/core/src/state-machine/index.ts`
### 撤回方式
```bash
git checkout -- packages/core/src/state-machine/index.ts
```

---

## 2026-05-30 15:56 — 前端状态机 TDD 红（RED）阶段初始化
### 问题
需要实战测试 TypeScript 类型与 Zustand 状态机，根据 TDD 规范，需先建立失败断言。
### 解决方案
在 `packages/core/src/state-machine/__tests__` 下创建了 `state_syntax.test.ts`，并写入一个必定断言失败的测试用例。
### 修改文件
- `packages/core/src/state-machine/__tests__/state_syntax.test.ts`
### 撤回方式
```bash
rm packages/core/src/state-machine/__tests__/state_syntax.test.ts
```

---

## 2026-05-30 16:15 — 前端状态机 TDD 绿（GREEN）阶段测试通过
### 问题
第一阶段的测试用例已完成编写，需要验证 Branded Types 的编译防护与 Zustand 状态机的状态转移正确性，并完成第一阶段的测试闭环。
### 解决方案
1. 用户在本地执行 `pnpm --filter @clocktower/core test`。
2. 测试通过，进入重构阶段。
### 修改文件
- 无（测试通过，无需修改）
### 撤回方式
- 无需撤回