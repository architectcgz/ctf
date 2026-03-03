<?php
$users = ['admin' => 'admin123', 'user' => 'password'];
$username = $_POST['username'] ?? '';
$password = $_POST['password'] ?? '';
if (isset($users[$username]) && $users[$username] === $password) {
    if ($username === 'admin') {
        echo 'Flag: ' . (getenv('FLAG') ?: 'flag{weak_password_cracked}');
    } else {
        echo '登录成功，但你不是管理员';
    }
} else {
    echo '<!DOCTYPE html><html><body><h2>登录</h2><form method="POST"><input name="username" placeholder="用户名"><input name="password" type="password" placeholder="密码"><button>登录</button></form></body></html>';
}
?>
