# 解法

利用 `__reduce__` 把对象还原过程改成 `eval('open("flag.txt").read()')`。服务端 `pickle.loads` 后会返回文件内容，页面直接把它打印出来。
