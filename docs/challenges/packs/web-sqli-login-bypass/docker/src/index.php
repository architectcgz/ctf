<!DOCTYPE html>
<html>
<head>
    <title>登录系统</title>
    <style>
        body { font-family: Arial; max-width: 400px; margin: 50px auto; }
        input { width: 100%; padding: 8px; margin: 5px 0; }
        button { width: 100%; padding: 10px; background: #007bff; color: white; border: none; cursor: pointer; }
        .error { color: red; }
    </style>
</head>
<body>
    <h2>登录系统</h2>
    <?php
    if (isset($_GET['error'])) {
        echo '<p class="error">登录失败！</p>';
    }
    ?>
    <form method="POST" action="login.php">
        <input type="text" name="username" placeholder="用户名" required>
        <input type="password" name="password" placeholder="密码" required>
        <button type="submit">登录</button>
    </form>
</body>
</html>
