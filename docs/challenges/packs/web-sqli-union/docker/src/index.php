<?php
$conn = new mysqli('localhost', 'root', 'root', 'ctf_db');
$id = $_GET['id'] ?? '1';
$sql = "SELECT id, name FROM users WHERE id = $id";
$result = $conn->query($sql);
echo '<!DOCTYPE html><html><body><h2>用户查询</h2>';
if ($result && $result->num_rows > 0) {
    while($row = $result->fetch_assoc()) {
        echo "<p>ID: {$row['id']}, Name: {$row['name']}</p>";
    }
} else {
    echo '<p>未找到用户</p>';
}
echo '</body></html>';
?>
