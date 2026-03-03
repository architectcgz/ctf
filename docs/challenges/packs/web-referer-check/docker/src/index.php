<?php
$referer = $_SERVER['HTTP_REFERER'] ?? '';
if (strpos($referer, 'trusted-site.com') !== false) {
    echo 'Flag: ' . (getenv('FLAG') ?: 'flag{referer_can_be_spoofed}');
} else {
    echo '<!DOCTYPE html><html><body><h2>访问被拒绝</h2><p>只允许从 trusted-site.com 访问</p></body></html>';
}
?>
