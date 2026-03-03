<?php
$host = 'localhost';
$user = 'root';
$pass = 'root';
$db = 'ctf_db';

$conn = new mysqli($host, $user, $pass, $db);

if ($conn->connect_error) {
    die("连接失败");
}

$username = $_POST['username'];
$password = $_POST['password'];

// 存在 SQL 注入漏洞的查询
$sql = "SELECT * FROM users WHERE username='$username' AND password='$password'";
$result = $conn->query($sql);

if ($result && $result->num_rows > 0) {
    $user = $result->fetch_assoc();
    if ($user['role'] === 'admin') {
        $flag = getenv('FLAG') ?: 'flag{test_flag}';
        echo "<!DOCTYPE html><html><head><title>成功</title></head><body>";
        echo "<h2>欢迎，管理员！</h2>";
        echo "<p>Flag: <strong>$flag</strong></p>";
        echo "</body></html>";
    } else {
        echo "<!DOCTYPE html><html><head><title>成功</title></head><body>";
        echo "<h2>欢迎，$username！</h2>";
        echo "<p>你不是管理员，无法查看 flag。</p>";
        echo "</body></html>";
    }
} else {
    header("Location: index.php?error=1");
}

$conn->close();
?>
