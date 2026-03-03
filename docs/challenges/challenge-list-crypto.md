# CTF 题目库 - 密码学类（50题）

> 来源：基于 picoCTF、CryptoHack 等平台题目整理
> 格式：challenge-pack-v1

---

## 入门级 (beginner) - 15题

### 1. crypto-base64-decode
- **标题**：Base64 解码
- **描述**：解码 Base64 编码的消息
- **标签**：`kp:base64`, `kp:encoding`
- **容器**：无

### 2. crypto-caesar-cipher
- **标题**：凯撒密码
- **描述**：破解凯撒密码（移位 13）
- **标签**：`kp:caesar`, `kp:rot13`
- **容器**：无

### 3. crypto-hex-decode
- **标题**：十六进制解码
- **描述**：将十六进制字符串转换为文本
- **标签**：`kp:hex`, `kp:encoding`
- **容器**：无

### 4. crypto-morse-code
- **标题**：摩斯密码
- **描述**：解码摩斯密码消息
- **标签**：`kp:morse`, `kp:classic-crypto`
- **容器**：无

### 5. crypto-ascii-shift
- **标题**：ASCII 移位
- **描述**：每个字符 ASCII 值加减固定数
- **标签**：`kp:shift-cipher`, `kp:ascii`
- **容器**：无

### 6. crypto-xor-single-byte
- **标题**：单字节 XOR
- **描述**：所有字节与单个密钥 XOR
- **标签**：`kp:xor`, `kp:brute-force`
- **容器**：无

### 7. crypto-substitution-cipher
- **标题**：简单替换密码
- **描述**：字母一对一替换
- **标签**：`kp:substitution`, `kp:frequency-analysis`
- **容器**：无

### 8. crypto-vigenere-cipher
- **标题**：维吉尼亚密码
- **描述**：多表替换密码，已知密钥长度
- **标签**：`kp:vigenere`, `kp:polyalphabetic`
- **容器**：无

### 9. crypto-rail-fence
- **标题**：栅栏密码
- **描述**：栅栏加密，已知栏数
- **标签**：`kp:rail-fence`, `kp:transposition`
- **容器**：无

### 10. crypto-atbash-cipher
- **标题**：埃特巴什码
- **描述**：字母表反转替换
- **标签**：`kp:atbash`, `kp:classic-crypto`
- **容器**：无

### 11. crypto-binary-decode
- **标题**：二进制解码
- **描述**：二进制字符串转文本
- **标签**：`kp:binary`, `kp:encoding`
- **容器**：无

### 12. crypto-url-decode
- **标题**：URL 解码
- **描述**：解码 URL 编码的字符串
- **标签**：`kp:url-encoding`, `kp:encoding`
- **容器**：无

### 13. crypto-base32-decode
- **标题**：Base32 解码
- **描述**：解码 Base32 编码
- **标签**：`kp:base32`, `kp:encoding`
- **容器**：无

### 14. crypto-multi-layer-encoding
- **标题**：多层编码
- **描述**：Base64 + Hex 多层编码
- **标签**：`kp:multi-encoding`, `kp:cyberchef`
- **容器**：无

### 15. crypto-keyboard-shift
- **标题**：键盘移位
- **描述**：QWERTY 键盘上的字符移位
- **标签**：`kp:keyboard-cipher`, `kp:classic-crypto`
- **容器**：无

---

## 简单级 (easy) - 15题

### 16. crypto-rsa-small-e
- **标题**：RSA 小公钥指数
- **描述**：e=3 的 RSA，直接开三次方
- **标签**：`kp:rsa`, `kp:small-exponent`
- **容器**：无

### 17. crypto-rsa-common-modulus
- **标题**：RSA 共模攻击
- **描述**：两个消息使用相同 n 不同 e
- **标签**：`kp:rsa`, `kp:common-modulus`
- **容器**：无

### 18. crypto-rsa-wiener
- **标题**：RSA Wiener 攻击
- **描述**：d 很小时的连分数攻击
- **标签**：`kp:rsa`, `kp:wiener-attack`
- **容器**：无

### 19. crypto-md5-collision
- **标题**：MD5 碰撞
- **描述**：找到两个 MD5 值相同的文件
- **标签**：`kp:hash`, `kp:collision`
- **容器**：无

### 20. crypto-hash-length-extension
- **标题**：哈希长度扩展攻击
- **描述**：利用 SHA1 长度扩展漏洞
- **标签**：`kp:hash`, `kp:length-extension`
- **容器**：无

### 21. crypto-weak-prng
- **标题**：弱随机数生成器
- **描述**：预测 LCG 随机数
- **标签**：`kp:prng`, `kp:lcg`
- **容器**：无

### 22. crypto-ecb-mode
- **标题**：ECB 模式检测
- **描述**：识别 AES-ECB 加密的重复块
- **标签**：`kp:aes`, `kp:ecb`
- **容器**：无

### 23. crypto-cbc-bit-flipping
- **标题**：CBC 比特翻转
- **描述**：修改密文影响明文
- **标签**：`kp:aes`, `kp:cbc`, `kp:bit-flipping`
- **容器**：无

### 24. crypto-padding-oracle
- **标题**：Padding Oracle 攻击
- **描述**：利用填充错误信息解密
- **标签**：`kp:aes`, `kp:padding-oracle`
- **容器**：Python

### 25. crypto-known-plaintext
- **标题**：已知明文攻击
- **描述**：已知部分明文推导密钥
- **标签**：`kp:known-plaintext`, `kp:xor`
- **容器**：无

### 26. crypto-frequency-analysis
- **标题**：频率分析
- **描述**：通过字母频率破解替换密码
- **标签**：`kp:frequency-analysis`, `kp:substitution`
- **容器**：无

### 27. crypto-one-time-pad-reuse
- **标题**：一次性密码本重用
- **描述**：密钥重用导致的 XOR 破解
- **标签**：`kp:otp`, `kp:key-reuse`
- **容器**：无

### 28. crypto-diffie-hellman-small-subgroup
- **标题**：DH 小子群攻击
- **描述**：Diffie-Hellman 参数选择不当
- **标签**：`kp:diffie-hellman`, `kp:small-subgroup`
- **容器**：无

### 29. crypto-elgamal-weak
- **标题**：ElGamal 弱参数
- **描述**：ElGamal 加密参数过小
- **标签**：`kp:elgamal`, `kp:discrete-log`
- **容器**：无

### 30. crypto-stream-cipher-reuse
- **标题**：流密码密钥重用
- **描述**：RC4 密钥流重用
- **标签**：`kp:stream-cipher`, `kp:rc4`
- **容器**：无

---

## 中等级 (medium) - 12题

### 31. crypto-rsa-factorization
- **标题**：RSA 大数分解
- **描述**：分解 512 位 RSA 模数
- **标签**：`kp:rsa`, `kp:factorization`
- **容器**：无

### 32. crypto-rsa-broadcast-attack
- **标题**：RSA 广播攻击
- **描述**：相同消息发送给多个接收者
- **标签**：`kp:rsa`, `kp:broadcast-attack`
- **容器**：无

### 33. crypto-ecdsa-nonce-reuse
- **标题**：ECDSA nonce 重用
- **描述**：签名时 k 值重用导致私钥泄露
- **标签**：`kp:ecdsa`, `kp:nonce-reuse`
- **容器**：无

### 34. crypto-bleichenbacher-attack
- **标题**：Bleichenbacher 攻击
- **描述**：RSA PKCS#1 v1.5 填充 oracle
- **标签**：`kp:rsa`, `kp:bleichenbacher`
- **容器**：Python

### 35. crypto-cbc-padding-attack
- **标题**：CBC Padding 攻击
- **描述**：完整的 Padding Oracle 利用
- **标签**：`kp:aes`, `kp:cbc`, `kp:padding-oracle`
- **容器**：Python

### 36. crypto-gcm-nonce-reuse
- **标题**：GCM nonce 重用
- **描述**：AES-GCM nonce 重用恢复密钥
- **标签**：`kp:aes`, `kp:gcm`, `kp:nonce-reuse`
- **容器**：无

### 37. crypto-dsa-weak-k
- **标题**：DSA 弱 k 值
- **描述**：DSA 签名 k 值可预测
- **标签**：`kp:dsa`, `kp:weak-randomness`
- **容器**：无

### 38. crypto-pohlig-hellman
- **标题**：Pohlig-Hellman 算法
- **描述**：离散对数问题的特殊情况
- **标签**：`kp:discrete-log`, `kp:pohlig-hellman`
- **容器**：无

### 39. crypto-coppersmith-attack
- **标题**：Coppersmith 攻击
- **描述**：RSA 部分已知明文
- **标签**：`kp:rsa`, `kp:coppersmith`
- **容器**：无

### 40. crypto-lattice-reduction
- **标题**：格基规约攻击
- **描述**：使用 LLL 算法破解背包密码
- **标签**：`kp:lattice`, `kp:lll`
- **容器**：无

### 41. crypto-meet-in-the-middle
- **标题**：中间相遇攻击
- **描述**：双重 DES 的中间相遇
- **标签**：`kp:des`, `kp:meet-in-middle`
- **容器**：无

### 42. crypto-timing-attack
- **标题**：时间侧信道攻击
- **描述**：通过响应时间推断密钥
- **标签**：`kp:side-channel`, `kp:timing`
- **容器**：Python

---

## 困难级 (hard) - 6题

### 43. crypto-rsa-lsb-oracle
- **标题**：RSA LSB Oracle
- **描述**：利用最低位泄露二分搜索明文
- **标签**：`kp:rsa`, `kp:lsb-oracle`
- **容器**：Python

### 44. crypto-fault-attack
- **标题**：故障攻击
- **描述**：模拟 AES 故障注入
- **标签**：`kp:aes`, `kp:fault-attack`
- **容器**：Python

### 45. crypto-invalid-curve-attack
- **标题**：无效曲线攻击
- **描述**：ECC 无效曲线点攻击
- **标签**：`kp:ecc`, `kp:invalid-curve`
- **容器**：Python

### 46. crypto-crt-rsa
- **标题**：CRT-RSA 故障攻击
- **描述**：中国剩余定理优化的 RSA 故障
- **标签**：`kp:rsa`, `kp:crt`, `kp:fault`
- **容器**：无

### 47. crypto-hidden-number-problem
- **标题**：隐藏数问题
- **描述**：从部分泄露恢复 ECDSA 私钥
- **标签**：`kp:ecdsa`, `kp:hnp`
- **容器**：无

### 48. crypto-isogeny-based
- **标题**：同源密码学
- **描述**：SIDH/SIKE 相关挑战
- **标签**：`kp:post-quantum`, `kp:isogeny`
- **容器**：无

---

## 地狱级 (hell) - 2题

### 49. crypto-quantum-resistant
- **标题**：后量子密码破解
- **描述**：破解弱参数的格密码
- **标签**：`kp:post-quantum`, `kp:lattice`
- **容器**：无

### 50. crypto-full-break
- **标题**：完整密码系统破解
- **描述**：多层加密系统的完整攻击链
- **标签**：`kp:full-break`, `kp:multi-layer`
- **容器**：Python

---

## 参考来源

- [CryptoHack](https://cryptohack.org/)
- [picoCTF Cryptography](https://picoctf.org/)
