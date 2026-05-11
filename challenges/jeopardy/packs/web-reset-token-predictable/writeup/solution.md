# 解法

请求接口会回显 `issued=<timestamp>`，而 token 直接来自 `random.Random(issued).randint(...)`。用同一个时间戳在本地重算令牌，再请求确认接口即可拿到 flag。
