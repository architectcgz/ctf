<?php
$data = $_GET['data'] ?? '';
$decoded = base64_decode($data);
if ($decoded === 'admin') {
    echo 'Flag: ' . (getenv('FLAG') ?: 'flag{base64_decoded}');
} else {
    echo '<!DOCTYPE html><html><body><h2>Base64 参数</h2><p>传递正确的 data 参数（Base64 编码）</p></body></html>';
}
?>
