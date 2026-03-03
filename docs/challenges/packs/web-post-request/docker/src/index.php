<?php
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $data = $_POST['data'] ?? '';
    if ($data === 'secret_data') {
        echo 'Flag: ' . (getenv('FLAG') ?: 'flag{post_request_success}');
    } else {
        echo '数据错误';
    }
} else {
    echo '<!DOCTYPE html><html><body><h2>POST 请求挑战</h2><p>需要发送 POST 请求</p></body></html>';
}
?>
