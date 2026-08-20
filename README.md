# 编程 OJ (online-judge)

纯 Go 标准库实现的在线判题（Online Judge）后端服务，零第三方依赖，开箱即跑。

## 运行

```bash
go run ./cmd/server
# 自定义端口 / 启用鉴权与限流
PORT=9090 AUTH_TOKEN=secret RATE_LIMIT_PER_IP=50 go run ./cmd/server
```

环境变量：

| 变量 | 说明 | 默认 |
|------|------|------|
| `PORT` / `ADDR` | 监听地址 | `:8080` |
| `MAX_PAGE_SIZE` | 分页最大条数 | `100` |
| `AUTH_TOKEN` | 非空时启用 Bearer 鉴权 | 空 |
| `RATE_LIMIT_PER_IP` | 每 IP 每秒请求数上限（<=0 不限） | `0` |
| `JUDGE_DELAY_MS` | 模拟判题耗时（毫秒） | `0` |

## 架构分层

```
cmd/server            入口：配置加载、依赖装配、优雅关闭
internal/app          依赖装配 store -> service -> handler
internal/config       环境变量配置（含鉴权/限流参数）
internal/model        领域模型 + 校验 + 状态机
internal/store        数据访问接口 + 内存实现
internal/service      业务逻辑 + 判题 + 排行榜
internal/handler      HTTP 路由与处理器
internal/middleware   鉴权（Bearer Token）与限流中间件
pkg/httpx             统一响应、分页、JSON 解析
pkg/idgen             ID 生成
pkg/logger            分级日志
```

## 核心业务

- 用户：注册、登录（SHA-256 密码摘要）、改密、角色（admin/user）。
- 题目：CRUD、难度/标签/状态、批量导入、标签聚合。
- 提交：判题状态机 `pending -> judging -> accepted/wrong_answer/time_limit/memory_limit/runtime_error/compile_error`，支持重判，判题明细独立存储。
- 比赛：状态机 `registration -> running -> ended`，报名（唯一约束），比赛题目关联。
- 公告：置顶、关键词筛选。
- 报表：排行榜、判题结果分布、题目通过率、语言分布、用户提交汇总、比赛排名。

## API 一览

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/users` | 注册用户 |
| POST | `/api/auth/login` | 登录 |
| POST | `/api/users/{id}/change-password` | 修改密码 |
| GET | `/api/users` | 用户列表（role/status/keyword 筛选，分页） |
| GET/PUT/DELETE | `/api/users/{id}` | 用户详情 / 更新 / 删除 |
| POST | `/api/problems` | 创建题目 |
| POST | `/api/problems/import` | 批量导入题目 |
| GET | `/api/problems` | 题目列表（difficulty/status/tag/keyword 筛选，分页） |
| GET | `/api/problems/tags` | 标签聚合 |
| GET/PUT/DELETE | `/api/problems/{id}` | 题目详情 / 更新 / 删除 |
| POST | `/api/submissions` | 提交代码 |
| GET | `/api/submissions` | 提交列表（problem_id/user_id/status/language 筛选，分页） |
| GET | `/api/submissions/{id}` | 提交详情 |
| POST | `/api/submissions/{id}/judge` | 判题 |
| POST | `/api/submissions/{id}/rejudge` | 重判 |
| GET | `/api/submissions/{id}/result` | 判题明细 |
| POST | `/api/contests` | 创建比赛 |
| GET | `/api/contests` | 比赛列表（status/keyword 筛选，分页） |
| GET | `/api/contests/{id}` | 比赛详情 |
| POST | `/api/contests/{id}/transition` | 推进比赛状态 |
| POST | `/api/contests/{id}/problems` | 添加比赛题目 |
| DELETE | `/api/contests/{id}/problems/{problemID}` | 移除比赛题目 |
| GET | `/api/contests/{id}/problems` | 比赛题目列表 |
| GET | `/api/contests/{id}/ranking` | 比赛排名 |
| POST | `/api/registrations` | 报名比赛 |
| GET | `/api/registrations` | 报名列表 |
| POST | `/api/announcements` | 创建公告 |
| GET | `/api/announcements` | 公告列表（pinned/keyword 筛选，分页） |
| GET/PUT/DELETE | `/api/announcements/{id}` | 公告详情 / 更新 / 删除 |
| GET | `/api/leaderboard?limit=10` | 排行榜 |
| GET | `/api/stats/verdicts` | 判题结果分布 |
| GET | `/api/stats/pass-rates` | 题目通过率 |
| GET | `/api/stats/languages` | 语言分布 |
| GET | `/api/stats/users` | 用户提交汇总 |

统一响应结构：`{"code":0,"message":"ok","data":...}`；错误码 400/401/404/409/429/500。

> 判题结果为确定性模拟（基于代码特征），便于演示与测试。
