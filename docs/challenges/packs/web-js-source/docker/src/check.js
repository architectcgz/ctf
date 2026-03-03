function checkPassword() {
    var pass = document.getElementById('pass').value;
    var flag = 'flag{javascript_is_client_side}';
    if (pass === 'secret123') {
        document.getElementById('result').innerText = 'Flag: ' + flag;
    } else {
        document.getElementById('result').innerText = '密码错误';
    }
}
