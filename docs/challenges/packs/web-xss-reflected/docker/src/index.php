<!DOCTYPE html>
<html>
<head>
    <title>搜索页面</title>
    <style>
        body { font-family: Arial; max-width: 600px; margin: 50px auto; }
        input { width: 70%; padding: 8px; }
        button { padding: 8px 20px; }
    </style>
</head>
<body>
    <h2>搜索功能</h2>
    <form method="GET">
        <input type="text" name="q" placeholder="输入搜索关键词">
        <button type="submit">搜索</button>
    </form>
    <?php
    if (isset($_GET['q'])) {
        $query = $_GET['q'];
        echo "<h3>搜索结果：$query</h3>";
        echo "<p>未找到相关结果。</p>";

        if (strpos($query, '<script>') !== false) {
            $flag = getenv('FLAG') ?: 'flag{xss_success}';
            echo "<p style='display:none'>Flag: $flag</p>";
        }
    }
    ?>
</body>
</html>
