<?php
$step = $_GET['step'] ?? '1';
if ($step === '1') {
    header('Location: ?step=2');
} elseif ($step === '2') {
    header('Location: ?step=3');
} elseif ($step === '3') {
    echo 'Flag: ' . (getenv('FLAG') ?: 'flag{follow_redirects}');
}
?>
