<?php
$secret = $_GET['secret'] ?? '';
?>
<!DOCTYPE html>
<html>
<head><title>参数传递</title></head>
<body>
<h2>参数传递挑战</h2>
<?php if ($secret === 'give_me_flag'): ?>
    <p>Flag: <?php echo getenv('FLAG') ?: 'flag{get_parameter_works}'; ?></p>
<?php else: ?>
    <p>提示：传递正确的 secret 参数</p>
<?php endif; ?>
</body>
</html>
