# AWD 二期运维指标待办

更新日期：2026-03-11

## 目标

围绕原始需求中的“竞赛管理”主线，继续补齐 AWD 二期 Checker / 运维链路，本轮只处理轮次级真实运维指标闭环，不扩散到 TCP 探测或额外 UI 打磨。

## 本轮范围

1. 后端 `AWD Round Summary` 增加轮次级运维指标
2. 管理端 AWD 面板展示新增指标
3. 补齐对应最小测试验证

## 本轮任务

1. [x] 在后端轮次汇总接口中输出以下真实统计：
   - 服务总数、正常数、下线数、失陷数
   - 攻击日志总数、成功数、失败数
   - 受攻击服务数、防守成功服务数
   - 调度巡检 / 当前轮手动 / 指定轮次重跑 / 人工补录 巡检数量
   - 学员提交 / 人工补录 / 历史记录 攻击日志数量
2. [x] 在 AWD 管理面板“回合态势”区域展示轮次级运维指标，支持直接用于值守与复盘
3. [x] 更新复盘包导出内容，携带轮次级运维指标
4. [x] 运行最小后端单测与前端类型检查/测试

## 验证记录

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest`
2. `cd /home/azhi/workspace/projects/ctf/code/frontend && npm run typecheck`
3. `cd /home/azhi/workspace/projects/ctf/code/frontend && npm run test -- --run src/views/admin/__tests__/ContestManage.test.ts src/api/__tests__/admin.test.ts`

## 不在本轮处理

1. 非 HTTP/TCP 级别的更深探测
2. 更深 exploit payload 复盘
3. 年级级趋势与更深层教学干预建议
