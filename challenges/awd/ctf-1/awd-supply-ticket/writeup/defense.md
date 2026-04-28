# 防守思路

- 启动后立即修改 `ADMIN_PASSWORD`，或移除默认管理员。
- `SECRET_KEY` 使用随机值并按队伍区分。
- `/notify/<id>` 不要对用户字段再次调用 `render_template_string`。
- 对工单标题和内容做长度限制与 HTML 转义。
- 保留 `/health`、创建工单、查看工单功能，避免只靠禁用接口防守。
