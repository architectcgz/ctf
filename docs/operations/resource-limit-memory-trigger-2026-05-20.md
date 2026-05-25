# 容器资源限制触发实验记录 2026-05-20

> 状态：Current
> 事实源：真实本地运行环境、`docker inspect`、`docker exec`、`curl`
> 替代：无

## 定位

这份记录只覆盖一组最小、受控、可回收的资源限制触发实验：对正在运行的 AWD 样本靶机容器执行匿名内存持续申请，验证容器内存配额是否真的生效，以及配额触发后平台和靶机服务是否仍保持可服务状态。

- 负责：给论文第 5 章补一组“资源限制不是只写在配置里，而是被真实触发过”的实验事实。
- 不负责：把这一次内存配额实验扩展成 CPU、PIDs、宿主机级隔离攻击或长稳运行结论。

## 实验目标

验证以下三件事：

1. 样本靶机容器确实带有显式内存配额，而不是默认无限制运行。
2. 当容器内进程持续申请并实际触碰匿名内存时，配额最终会触发并终止超限进程。
3. 配额触发后，靶机主服务和平台控制面仍保持可访问状态，没有出现整容器退出或平台依赖失健康。

## 实验环境

- 实验日期：2026-05-20
- 仓库位置：`ctf 仓库根目录`
- 宿主机状态：
  - `load average`: `0.36 0.51 0.85`
  - 可用内存：约 `12427 MiB`
  - 交换分区占用：`0 MiB / 4096 MiB`
- 平台健康基线：`http://127.0.0.1:8080/health` 返回 `status=ok`
- 样本容器：`ctf-instance-challenge-c8-t15-s21`
- 样本容器对外服务端口：`30004 -> 8080/tcp`

选择这个样本容器的原因是它已经稳定运行约 6 小时，且能通过 `http://127.0.0.1:30004/health` 返回 `{"status":"ok"}`，适合做单变量受控实验。

## 配额基线

实验前先检查 Docker 运行配置与容器内 cgroup 视图：

```bash
docker inspect -f 'Memory={{.HostConfig.Memory}} MemorySwap={{.HostConfig.MemorySwap}} NanoCPUs={{.HostConfig.NanoCpus}} PidsLimit={{.HostConfig.PidsLimit}}' \
  ctf-instance-challenge-c8-t15-s21

docker exec ctf-instance-challenge-c8-t15-s21 sh -lc '
echo memory.max=$(cat /sys/fs/cgroup/memory.max 2>/dev/null || cat /sys/fs/cgroup/memory/memory.limit_in_bytes 2>/dev/null)
echo pids.max=$(cat /sys/fs/cgroup/pids.max 2>/dev/null || cat /sys/fs/cgroup/pids/pids.max 2>/dev/null)
'
```

实测结果：

- `Memory=268435456`，即 `256 MiB`
- `MemorySwap=536870912`，即 `512 MiB`
- `NanoCPUs=500000000`，即 `0.5 CPU`
- `PidsLimit=100`
- 容器内 `memory.max=268435456`
- 容器内 `pids.max=100`

这里需要特别说明：这次实验不能把“256 MiB 内存上限”直接理解成“进程一超过 256 MiB 就立刻被杀”。因为 Docker 同时设置了 `MemorySwap=512 MiB`，所以匿名内存压力在触发纯内存上限后，还可能继续消耗一部分 swap-backed 配额，最终在更高的总占用附近被终止。

## 实验步骤

### 1. 记录实验前健康状态

```bash
curl -fsS http://127.0.0.1:8080/health
curl -fsS http://127.0.0.1:30004/health
docker ps --format 'table {{.Names}}\t{{.Status}}'
```

判定基线：

- 平台健康接口返回 `status=ok`
- 样本靶机 `/health` 返回 `{"status":"ok"}`
- 样本容器为 `Up` 状态

### 2. 在容器内执行受控匿名内存申请

执行以下命令，让 Python 每轮申请并实际触碰 `16 MiB` 匿名内存，直到进程被终止或循环结束：

```bash
docker exec ctf-instance-challenge-c8-t15-s21 python3 -u -c '
import time
chunks = []
chunk = 16 * 1024 * 1024
for i in range(1, 40):
    b = bytearray(chunk)
    for j in range(0, len(b), 4096):
        b[j] = 1
    chunks.append(b)
    print(f"allocated={i*16}MiB", flush=True)
    time.sleep(0.2)
'
```

### 3. 记录实验后健康状态

```bash
docker ps --format '{{.Names}} {{.Status}}' | rg 'ctf-instance-challenge-c8-t15-s21|ctf-frontend|ctf-postgres|ctf-redis'
curl -fsS http://127.0.0.1:8080/health
curl -fsS http://127.0.0.1:30004/health
```

## 实验结果

### 1. 资源限制触发结果

本次 `docker exec` 命令最终返回：

```text
exit_code=137
```

匿名内存申请过程打印到：

```text
allocated=16MiB
allocated=32MiB
allocated=48MiB
allocated=64MiB
allocated=80MiB
allocated=96MiB
allocated=112MiB
allocated=128MiB
allocated=144MiB
allocated=160MiB
allocated=176MiB
allocated=192MiB
allocated=208MiB
allocated=224MiB
allocated=240MiB
allocated=256MiB
allocated=272MiB
allocated=288MiB
allocated=304MiB
allocated=320MiB
allocated=336MiB
allocated=352MiB
allocated=368MiB
allocated=384MiB
allocated=400MiB
allocated=416MiB
allocated=432MiB
allocated=448MiB
allocated=464MiB
allocated=480MiB
```

这说明：

- 受控申请进程没有正常跑完 39 轮，而是在实验过程中被终止。
- 终止时机晚于 `256 MiB`，与 `MemorySwap=512 MiB` 的 Docker 配置一致，说明该容器不是单纯“只看 RSS 到 256 MiB 立即杀”，而是受 `Memory + Swap` 组合配额约束。
- 由于脚本对每个 `bytearray` 都按 `4096` 字节步长显式写入，内存页不是只被惰性分配，实验结果可以视为真实匿名内存压力，而不是只停留在虚拟地址空间申请。

### 2. 触发后的平台与靶机状态

实验后复查结果：

- `ctf-instance-challenge-c8-t15-s21` 仍为 `Up`
- `ctf-frontend`、`ctf-postgres`、`ctf-redis` 仍为健康状态
- `http://127.0.0.1:8080/health` 继续返回：

```json
{"code":0,"message":"success","data":{"status":"ok","service":"ctf-platform","environment":"dev","dependencies":{"postgres":"ok","redis":"ok"},"version":"dev"},"request_id":"..."}
```

- `http://127.0.0.1:30004/health` 继续返回：

```json
{"status":"ok"}
```

这说明本次被终止的是容器内临时压测进程，而不是把整只靶机容器或平台控制面一起拖垮。

## 结论

这组实验可以支持如下结论：

1. 平台为靶机容器设置的资源限制不是停留在配置层说明里，而是能在真实运行态下触发。
2. 在 `Memory=256 MiB`、`MemorySwap=512 MiB` 的配置下，容器内持续增长的匿名内存负载最终会被系统终止，无法无限制扩张。
3. 资源限制触发后，样本靶机主服务和平台健康接口仍保持可访问，说明资源配额对异常进程具有约束作用，同时没有直接导致平台整体失服务。

## 论文可直接引用的表述

可在论文中写成：

> 2026 年 5 月 20 日，本文在真实运行的 AWD 样本靶机容器上执行了一组受控内存压力实验。样本容器配置为 `Memory=256 MiB`、`MemorySwap=512 MiB`、`PidsLimit=100`、`NanoCPUs=0.5`。实验中，容器内匿名内存申请进程在持续分配并触碰内存页后被系统以 `exit code 137` 终止，而靶机健康接口与平台 `/health` 在实验后仍保持正常响应。这表明平台配置的资源限制能够在真实运行态下约束异常负载，而不是仅停留在静态配置层面。

## 边界与后续

- 这次只验证了“内存配额会触发并终止超限进程”，没有覆盖 CPU 限速统计或 PIDs 上限触发。
- 这次实验没有证明宿主机级安全隔离，也不等于 12 小时以上长稳运行。
- 如果后续还需要补更完整的资源限制证据，建议再加两组最小实验：
  - `PidsLimit=100` 触发实验
  - `cpu.max / NanoCPUs` 节流统计实验
