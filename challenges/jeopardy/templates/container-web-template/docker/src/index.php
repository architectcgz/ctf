<?php
$flag = getenv('FLAG') ?: 'flag{local_template_flag}';
?>
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>Template Web Challenge</title>
</head>
<body>
  <h1>Template Web Challenge</h1>
  <p>把这里替换成真实题目逻辑。</p>
  <p>本地调试默认 flag：<code><?php echo htmlspecialchars($flag, ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8'); ?></code></p>
</body>
</html>
