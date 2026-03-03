# CTF 题目库 - 取证分析类（50题）

> 来源：基于各大 CTF 平台取证题目整理
> 格式：challenge-pack-v1

---

## 入门级 (beginner) - 15题

### 1. forensics-file-header
- **标题**：文件头修复
- **描述**：修复损坏的 PNG 文件头
- **标签**：`kp:file-header`, `kp:magic-number`

### 2. forensics-exif-data
- **标题**：EXIF 元数据
- **描述**：从图片 EXIF 中提取 flag
- **标签**：`kp:exif`, `kp:metadata`

### 3. forensics-strings-search
- **标题**：字符串搜索
- **描述**：使用 strings 命令找 flag
- **标签**：`kp:strings`, `kp:text-search`

### 4. forensics-binwalk-extract
- **标题**：Binwalk 提取
- **描述**：从固件中提取隐藏文件
- **标签**：`kp:binwalk`, `kp:extraction`

### 5. forensics-zip-password
- **标题**：ZIP 密码破解
- **描述**：爆破 ZIP 文件密码
- **标签**：`kp:zip`, `kp:password-crack`

### 6. forensics-deleted-file
- **标题**：删除文件恢复
- **描述**：恢复已删除的文件
- **标签**：`kp:file-recovery`, `kp:foremost`

### 7. forensics-lsb-steganography
- **标题**：LSB 隐写
- **描述**：提取图片 LSB 隐藏的数据
- **标签**：`kp:lsb`, `kp:steganography`

### 8. forensics-audio-spectrogram
- **标题**：音频频谱分析
- **描述**：从音频频谱图中找 flag
- **标签**：`kp:audio`, `kp:spectrogram`

### 9. forensics-pcap-basic
- **标题**：流量包分析基础
- **描述**：从 HTTP 流量中提取文件
- **标签**：`kp:pcap`, `kp:wireshark`

### 10. forensics-memory-strings
- **标题**：内存字符串
- **描述**：从内存镜像中搜索字符串
- **标签**：`kp:memory`, `kp:strings`

### 11. forensics-pdf-hidden
- **标题**：PDF 隐藏内容
- **描述**：提取 PDF 中的隐藏文本
- **标签**：`kp:pdf`, `kp:hidden-text`

### 12. forensics-qr-code
- **标题**：二维码识别
- **描述**：扫描二维码获取 flag
- **标签**：`kp:qr-code`, `kp:barcode`

### 13. forensics-file-carving
- **标题**：文件雕刻
- **描述**：从磁盘镜像中雕刻文件
- **标签**：`kp:file-carving`, `kp:scalpel`

### 14. forensics-hex-dump
- **标题**：十六进制分析
- **描述**：分析十六进制数据找异常
- **标签**：`kp:hex`, `kp:analysis`

### 15. forensics-log-analysis
- **标题**：日志分析
- **描述**：从日志文件中找线索
- **标签**：`kp:log`, `kp:grep`

---

## 简单级 (easy) - 15题

### 16. forensics-volatility-basic
- **标题**：Volatility 基础
- **描述**：使用 Volatility 分析内存
- **标签**：`kp:volatility`, `kp:memory`

### 17. forensics-network-protocol
- **标题**：网络协议分析
- **描述**：分析 FTP 流量提取文件
- **标签**：`kp:ftp`, `kp:protocol`

### 18. forensics-usb-traffic
- **标题**：USB 流量分析
- **描述**：从 USB 流量还原键盘输入
- **标签**：`kp:usb`, `kp:keyboard`

### 19. forensics-disk-image
- **标题**：磁盘镜像分析
- **描述**：挂载并分析磁盘镜像
- **标签**：`kp:disk-image`, `kp:mount`

### 20. forensics-registry-analysis
- **标题**：注册表分析
- **描述**：分析 Windows 注册表
- **标签**：`kp:registry`, `kp:windows`

### 21. forensics-steghide
- **标题**：Steghide 隐写
- **描述**：使用 steghide 提取隐藏数据
- **标签**：`kp:steghide`, `kp:steganography`

### 22. forensics-outguess
- **标题**：Outguess 隐写
- **描述**：使用 outguess 提取数据
- **标签**：`kp:outguess`, `kp:steganography`

### 23. forensics-stegsolve
- **标题**：Stegsolve 分析
- **描述**：使用 stegsolve 分析图片
- **标签**：`kp:stegsolve`, `kp:image`

### 24. forensics-gif-frame
- **标题**：GIF 帧分析
- **描述**：提取 GIF 每一帧找 flag
- **标签**：`kp:gif`, `kp:animation`

### 25. forensics-video-frame
- **标题**：视频帧提取
- **描述**：从视频中提取特定帧
- **标签**：`kp:video`, `kp:ffmpeg`

### 26. forensics-sqlite-database
- **标题**：SQLite 数据库
- **描述**：分析 SQLite 数据库文件
- **标签**：`kp:sqlite`, `kp:database`

### 27. forensics-browser-history
- **标题**：浏览器历史
- **描述**：分析浏览器历史记录
- **标签**：`kp:browser`, `kp:history`

### 28. forensics-email-analysis
- **标题**：邮件分析
- **描述**：分析 EML 邮件文件
- **标签**：`kp:email`, `kp:eml`

### 29. forensics-office-macro
- **标题**：Office 宏分析
- **描述**：提取并分析 Office 宏
- **标签**：`kp:office`, `kp:macro`

### 30. forensics-android-backup
- **标题**：Android 备份分析
- **描述**：分析 Android 备份文件
- **标签**：`kp:android`, `kp:backup`

---

## 中等级 (medium) - 12题

### 31. forensics-memory-process
- **标题**：内存进程分析
- **描述**：从内存中提取进程信息
- **标签**：`kp:volatility`, `kp:process`

### 32. forensics-memory-password
- **标题**：内存密码提取
- **描述**：从内存中提取密码
- **标签**：`kp:volatility`, `kp:mimikatz`

### 33. forensics-ntfs-ads
- **标题**：NTFS 交替数据流
- **描述**：提取 NTFS ADS 隐藏数据
- **标签**：`kp:ntfs`, `kp:ads`

### 34. forensics-ext4-journal
- **标题**：EXT4 日志分析
- **描述**：分析 EXT4 文件系统日志
- **标签**：`kp:ext4`, `kp:journal`

### 35. forensics-encrypted-zip
- **标题**：加密 ZIP 分析
- **描述**：已知明文攻击破解 ZIP
- **标签**：`kp:zip`, `kp:known-plaintext`

### 36. forensics-rar-recovery
- **标题**：RAR 恢复记录
- **描述**：使用恢复记录修复 RAR
- **标签**：`kp:rar`, `kp:recovery`

### 37. forensics-polyglot-file
- **标题**：多语言文件
- **描述**：同时是 PNG 和 ZIP 的文件
- **标签**：`kp:polyglot`, `kp:file-format`

### 38. forensics-timeline-analysis
- **标题**：时间线分析
- **描述**：构建事件时间线
- **标签**：`kp:timeline`, `kp:plaso`

### 39. forensics-malware-traffic
- **标题**：恶意流量分析
- **描述**：分析恶意软件网络流量
- **标签**：`kp:malware`, `kp:traffic`

### 40. forensics-covert-channel
- **标题**：隐蔽信道
- **描述**：发现并提取隐蔽信道数据
- **标签**：`kp:covert-channel`, `kp:network`

### 41. forensics-firmware-extract
- **标题**：固件提取分析
- **描述**：提取并分析路由器固件
- **标签**：`kp:firmware`, `kp:iot`

### 42. forensics-docker-forensics
- **标题**：Docker 取证
- **描述**：分析 Docker 容器镜像
- **标签**：`kp:docker`, `kp:container`

---

## 困难级 (hard) - 6题

### 43. forensics-anti-forensics
- **标题**：反取证技术
- **描述**：绕过反取证措施
- **标签**：`kp:anti-forensics`, `kp:evasion`

### 44. forensics-encrypted-disk
- **标题**：加密磁盘分析
- **描述**：分析 LUKS 加密磁盘
- **标签**：`kp:encryption`, `kp:luks`

### 45. forensics-memory-rootkit
- **标题**：内存 Rootkit 检测
- **描述**：检测并分析内存 rootkit
- **标签**：`kp:rootkit`, `kp:memory`

### 46. forensics-cloud-forensics
- **标题**：云取证
- **描述**：分析云服务日志和快照
- **标签**：`kp:cloud`, `kp:aws`

### 47. forensics-mobile-forensics
- **标题**：移动设备取证
- **描述**：完整的手机取证分析
- **标签**：`kp:mobile`, `kp:ios-android`

### 48. forensics-network-forensics
- **标题**：网络取证
- **描述**：大规模网络流量分析
- **标签**：`kp:network`, `kp:big-data`

---

## 地狱级 (hell) - 2题

### 49. forensics-apt-investigation
- **标题**：APT 攻击调查
- **描述**：完整的 APT 攻击链分析
- **标签**：`kp:apt`, `kp:investigation`

### 50. forensics-incident-response
- **标题**：应急响应
- **描述**：真实的安全事件响应
- **标签**：`kp:incident-response`, `kp:real-world`

---

## 参考来源

- [CTF Forensics Challenges](https://hackerdna.com/blog/ctf-categories)

