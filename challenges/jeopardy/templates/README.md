# 教师出题模板

这里放的是“可复制起步模板”，不是正式题目。

使用方式：

1. 从这里选择一个最接近的模板目录
2. 复制到 `challenges/jeopardy/packs/<your-slug>/`
3. 把模板中的占位内容全部替换掉
4. 按 [teacher-authoring-guide.md](/home/azhi/workspace/projects/ctf/challenges/teacher-authoring-guide.md) 完成自测
5. 生成 `challenges/jeopardy/dist/<your-slug>.zip`

当前模板：

- `offline-static-template/`：离线静态题模板，适合编码、逆向、取证、基础 crypto
- `container-web-template/`：单容器 Web 题模板，适合最小化 Web 靶机
- `container-pwn-template/`：单容器 Pwn 题模板，适合 HTTP 下载页 + TCP 利用端口

## 推荐复制命令

```bash
cd /home/azhi/workspace/projects/ctf/challenges/jeopardy
cp -R templates/offline-static-template packs/<your-slug>
```

或：

```bash
cd /home/azhi/workspace/projects/ctf/challenges/jeopardy
cp -R templates/container-web-template packs/<your-slug>
```

或：

```bash
cd /home/azhi/workspace/projects/ctf/challenges/jeopardy
cp -R templates/container-pwn-template packs/<your-slug>
```

复制后至少要改这些内容：

- `challenge.yml` 里的 `slug`、`title`、`category`、`difficulty`、`tags`
- `statement.md`
- `attachments/` 或 `docker/` 中的真实题目材料
- 容器题的镜像入口固定为 `docker/Dockerfile`
- 所有 `<...>` 占位符
- 所有默认 flag 或演示逻辑
- Pwn 模板里的默认下载页、演示漏洞和 `9999/tcp` 端口说明

## 生成分发 zip

模板复制并修改完成后，可用下面的方式刷新外层分发包：

```bash
cd /home/azhi/workspace/projects/ctf/challenges/jeopardy
mkdir -p dist
cd packs
zip -r ../dist/<your-slug>.zip <your-slug>
```

## 最低自测要求

离线题至少确认：

- `packs/<your-slug>/challenge.yml`、`statement.md`、附件齐全
- `dist/<your-slug>.zip` 解压后根目录仍为 `<your-slug>/`

容器题至少额外执行：

```bash
docker build -t test-<your-slug> packs/<your-slug>/docker
docker run --rm -p 18080:80 test-<your-slug>
```

如果是 Pwn 容器题，至少还应验证：

```bash
docker run --rm -p 18080:80 -p 19999:9999 -e FLAG='flag{local_test}' test-<your-slug>
curl http://127.0.0.1:18080/
nc 127.0.0.1 19999
```
