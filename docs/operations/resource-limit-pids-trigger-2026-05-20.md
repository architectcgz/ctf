# 容器进程数限制触发实验记录 2026-05-20

> 状态：Current
> 事实源：真实本地运行环境、`docker inspect`、`docker exec`、`curl`
> 替代：无

## 定位

这份记录只覆盖一组最小、受控、可回收的 `PidsLimit` 触发实验：对正在运行的 AWD 样本靶机容器短时间拉起大量子进程，验证容器进程数上限是否真实生效，以及限制触发后靶机服务和核心依赖容器是否仍保持正常状态。

- 负责：给论文第 5 章补一组“进程数限制不仅存在于配置中，而且能在真实运行态下触发”的实验事实。
- 不负责：把这次实验扩展成 CPU 节流、宿主机级安全隔离或长稳运行结论。

## 实验目标

验证以下三件事：

1. 样本靶机容器确实带有显式 `PidsLimit`，而不是允许无限制派生子进程。
2. 当容器内临时进程持续增长时，达到上限后会出现明确的进程创建失败，而不是无限制继续扩张。
3. 限制触发并清理后，样本靶机主服务和核心依赖容器仍保持正常状态，没有出现整容器退出。

## 实验环境

- 实验日期：2026-05-20
- 仓库位置：`ctf 仓库根目录`
- 宿主机状态：
  - `load average`: `0.77 0.61 0.78`
  - 可用内存：约 `12519 MiB`
  - 交换分区占用：约 `21 MiB / 4096 MiB`
- 样本容器：`ctf-instance-challenge-c8-t15-s21`
- 样本容器对外服务端口：`30004 -> 8080/tcp`
- 样本靶机健康接口：`http://127.0.0.1:30004/health`

本轮实验没有把 `127.0.0.1:8080/health` 作为平台基线，因为当前运行环境下该端口没有对外监听；因此，平台侧只复查核心依赖容器状态，不把“API 端口对外健康”作为本轮结论的一部分。

## 配额基线

实验前先检查 Docker 运行配置和容器内当前进程数：

```bash
docker inspect -f 'PidsLimit={{.HostConfig.PidsLimit}} Memory={{.HostConfig.Memory}} MemorySwap={{.HostConfig.MemorySwap}} NanoCPUs={{.HostConfig.NanoCpus}}' \
  ctf-instance-challenge-c8-t15-s21

docker exec ctf-instance-challenge-c8-t15-s21 sh -lc '
echo pids.max=$(cat /sys/fs/cgroup/pids.max 2>/dev/null || cat /sys/fs/cgroup/pids/pids.max 2>/dev/null)
echo current_processes=$(find /proc -maxdepth 1 -type d -name "[0-9]*" | wc -l)
'
```

实测结果：

- `PidsLimit=100`
- `Memory=268435456`，即 `256 MiB`
- `MemorySwap=536870912`，即 `512 MiB`
- `NanoCPUs=500000000`，即 `0.5 CPU`
- 容器内 `pids.max=100`
- 实验前容器当前进程数约为 `5`

这说明该容器不是只做了 CPU 和内存配额，还显式限制了总进程数。

## 实验步骤

### 1. 记录实验前状态

```bash
curl -fsS http://127.0.0.1:30004/health
docker ps --format '{{.Names}} {{.Status}}' | rg 'ctf-instance-challenge-c8-t15-s21|ctf-frontend|ctf-postgres|ctf-redis'
```

判定基线：

- 样本靶机 `/health` 返回 `{"status":"ok"}`
- `ctf-instance-challenge-c8-t15-s21` 为 `Up`
- `ctf-frontend`、`ctf-postgres`、`ctf-redis` 保持健康状态

### 2. 在容器内持续派生短生命周期子进程

执行以下命令，让 Python 在容器内持续拉起 `sleep 20` 子进程，直到 `fork/exec` 失败；随后在同一个父进程里立刻回收这些子进程：

```bash
docker exec ctf-instance-challenge-c8-t15-s21 python3 -u -c '
import subprocess, time
children = []
try:
    for i in range(1, 130):
        p = subprocess.Popen(["sleep", "20"])
        children.append(p)
        print(f"spawned={len(children)}", flush=True)
        time.sleep(0.03)
except Exception as e:
    print(f"spawn_error={type(e).__name__}:{e}", flush=True)
finally:
    print(f"cleanup_children={len(children)}", flush=True)
    for p in reversed(children):
        try:
            p.terminate()
        except Exception:
            pass
    time.sleep(0.3)
    for p in reversed(children):
        try:
            p.wait(timeout=0.5)
        except Exception:
            try:
                p.kill()
            except Exception:
                pass
    print("cleanup_done", flush=True)
'
```

### 3. 记录实验后状态

```bash
docker exec ctf-instance-challenge-c8-t15-s21 sh -lc '
echo current_processes=$(find /proc -maxdepth 1 -type d -name "[0-9]*" | wc -l)
'

curl -fsS http://127.0.0.1:30004/health
docker ps --format '{{.Names}} {{.Status}}' | rg 'ctf-instance-challenge-c8-t15-s21|ctf-frontend|ctf-postgres|ctf-redis'
```

## 实验结果

### 1. 进程数限制触发结果

本次 `docker exec` 命令最终正常返回，关键输出如下：

```text
spawned=1
...
spawned=98
spawn_error=BlockingIOError:[Errno 11] Resource temporarily unavailable
cleanup_children=98
cleanup_done
```

这说明：

- 容器内持续派生子进程时，并没有无限制增长。
- 当派生到 `98` 个临时 `sleep` 子进程时，新的进程创建开始失败，错误为 `BlockingIOError: [Errno 11] Resource temporarily unavailable`。
- 结合实验前容器内原有进程数约 `5`，以及本次 `docker exec` 运行时的 shell/Python 父进程开销，可以判定失败点与 `PidsLimit=100` 的配额一致。
- 失败后同一个父进程完成了子进程回收，没有把临时压力遗留在容器里。

### 2. 触发后的样本容器与依赖状态

实验后复查结果：

- 容器内当前进程数恢复为 `5`
- `ctf-instance-challenge-c8-t15-s21` 仍为 `Up`
- 样本靶机 `http://127.0.0.1:30004/health` 继续返回：

```json
{"status":"ok"}
```

- `ctf-frontend`、`ctf-postgres`、`ctf-redis` 仍保持健康状态

这说明本次被限制的是容器内临时派生进程的数量，而不是把整只靶机容器直接拖死；同时，受控清理后容器恢复到实验前的稳定状态。

## 结论

这组实验可以支持如下结论：

1. 平台为靶机容器设置的 `PidsLimit` 不是停留在静态配置说明里，而是能在真实运行态下触发。
2. 在 `PidsLimit=100` 的配置下，容器内持续派生子进程会在接近该上限时出现明确的进程创建失败，无法无限制扩张。
3. 限制触发并完成清理后，样本靶机健康接口仍保持正常响应，核心依赖容器也未出现失健康，说明进程数配额对异常进程膨胀具有约束作用。

## 论文可直接引用的表述

可在论文中写成：

> 2026 年 5 月 20 日，本文在真实运行的 AWD 样本靶机容器上执行了一组受控进程数压力实验。样本容器配置为 `PidsLimit=100`。实验中，容器内通过 Python 持续派生短生命周期子进程，当累计派生到 98 个临时子进程时，新的进程创建返回 `BlockingIOError: [Errno 11] Resource temporarily unavailable`，表明进程数配额已被真实触发。实验完成并回收子进程后，容器内进程数恢复到实验前水平，靶机健康接口仍保持正常响应。这说明平台配置的进程数限制能够在真实运行态下抑制异常进程膨胀，而不是仅停留在静态配置层面。

## 边界与后续

- 这次只验证了 `PidsLimit` 会触发，不等于已经覆盖 CPU 节流统计。
- 这次实验没有证明宿主机级安全隔离，也不等于长稳运行。
- 如果后续还需要补更完整的资源限制证据，建议把本记录与内存配额触发记录一起引用，再补一组 `NanoCPUs=0.5` 的节流统计实验。
