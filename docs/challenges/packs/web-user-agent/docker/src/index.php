<?php
$ua = $_SERVER['HTTP_USER_AGENT'] ?? '';
if (strpos($ua, 'CTFBot') !== false) {
    echo 'Flag: ' . (getenv('FLAG') ?: 'flag{user_agent_spoofed}');
} else {
    echo '<!DOCTYPE html><html><body><h2>User-Agent 检测</h2><p>你的浏览器不被允许访问</p></body></html>';
}
?>
