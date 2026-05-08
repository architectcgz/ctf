# 防守题解

## 可修改位置

- 平台防守入口实际对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内的源文件位置是 `docker/workspace/src/challenge_app.py`
- `docker/runtime/app.py`、`docker/runtime/ctf_runtime.py`、`docker/check/check.py` 都在保护边界内，不是这题的主要修补位置

## 主要漏洞点

### 1. 默认管理员口令过弱

`init_business_db()` 里把管理员默认口令写成了 `admin123`。这意味着攻击方只要访问 `/login`，就可以直接撞后台口令，再从 `/admin` 读取当前 Flag。

这题当前工作区里能直接改的就是这里：

- 不再保留 `admin123` 这种固定弱口令
- 至少改成你自己控制的高强度值
- 更稳妥的做法是只在 `ADMIN_PASSWORD` 存在时初始化管理员，或者把 fallback 改成每队自定义强口令

### 2. 通知接口存在模板注入

`/notify/<ticket_id>` 会把工单标题直接拼进模板，再次调用 `render_template_string()`。只要有人提交带模板表达式的标题，就能在通知页触发 SSTI。

最小修法：

- 不要再把用户输入当模板执行
- 先把标题和状态做转义
- 最后把纯文本通知内容包进普通 HTML，例如直接返回 `<p>...</p>`

核心原则只有一句话：用户提交的标题只能当数据展示，不能再当 Jinja 模板解释。

### 3. 工单内容原样回显

`new_ticket()` 保存的 `title` / `content` 在首页会被原样拼进 HTML。即使你已经修掉了 `/notify` 的 SSTI，首页这里也仍然会留下注入面。

建议一起做两步：

- 在 `new_ticket()` 对标题和内容做长度限制
- 在 `index()` 和通知页统一做 HTML 转义

如果你想把修补做得更稳一点，可以在写入前先 `strip()` 并限制长度，例如标题 80 字、正文 2000 字，再在展示时统一 `html.escape()`

## 推荐修改函数

- `init_business_db()`
- `index()`
- `new_ticket()`
- `notify()`

## 保活约束

修补时不要直接删接口，至少要保留下面这些能力：

- `/health` 正常返回 200
- `/api/flag` 继续支持平台 checker 的 PUT / GET
- 首页能查看工单
- `/new` 还能创建工单
- `/notify/<id>` 还能查看通知，只是不再执行模板
- `/login` 和 `/admin` 入口仍然可用

## 交付判断

这题修完后，至少应满足：

1. 使用原来的 `admin/admin123` 不能再直接进后台
2. 提交包含 `{{ 7*7 }}` 之类内容的工单标题后，访问 `/notify/<id>` 不会执行模板
3. 首页和通知页仍能正常查看工单内容
