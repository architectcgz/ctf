<?php
session_start();
if (!isset($_COOKIE['role'])) {
    setcookie('role', 'guest', time() + 3600, '/');
    $_COOKIE['role'] = 'guest';
}
?>
<!DOCTYPE html>
<html>
<head>
    <title>用户系统</title>
    <style>
        body { font-family: Arial; max-width: 600px; margin: 50px auto; }
        .info { padding: 20px; background: #f0f0f0; border-radius: 5px; }
        .admin { background: #d4edda; }
    </style>
</head>
<body>
    <h2>用户系统</h2>
    <div class="info <?php echo $_COOKIE['role'] === 'admin' ? 'admin' : ''; ?>">
        <p>当前身份：<strong><?php echo htmlspecialchars($_COOKIE['role']); ?></strong></p>
        <?php if ($_COOKIE['role'] === 'admin'): ?>
            <p>欢迎管理员！</p>
            <p>Flag: <code><?php echo getenv('FLAG') ?: 'flag{cookie_is_not_secure}'; ?></code></p>
        <?php else: ?>
            <p>你是普通用户，无法查看 flag。</p>
            <p>提示：检查你的 Cookie...</p>
        <?php endif; ?>
    </div>
</body>
</html>
