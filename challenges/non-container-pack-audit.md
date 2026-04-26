# Non-Container Pack Audit

## 判定规则

- 仅审计 `runtime.type != container` 的题目包。
- 若包内除 `challenge.yml`、`statement.md` 和 `docker/` 外，还包含实际可交付材料，则归为“可作为离线题发布”。
- 若只剩题面与元数据，没有附件、样本、二进制、压缩包等交付物，则归为“仅是题卡、不能发布”。
- 这个报告只解决“学生是否有材料可做”，不判断题目质量、答案正确性或附件内容是否充分。

## 汇总

- 非容器题总数：`180`
- 可作为离线题发布：`180`
- 仅是题卡、不能发布：`0`

### 分类统计

- `crypto`: 离线可发布 `44`，题卡 `0`
- `forensics`: 离线可发布 `53`，题卡 `0`
- `misc`: 离线可发布 `30`，题卡 `0`
- `reverse`: 离线可发布 `53`，题卡 `0`

## 可作为离线题发布

### crypto（44）

- `crypto-advanced-lattice`: 有离线材料 `attachments/params.txt`
- `crypto-ascii-shift`: 有离线材料 `attachments/challenge.txt`
- `crypto-atbash`: 有离线材料 `attachments/challenge.txt`
- `crypto-base32`: 有离线材料 `attachments/challenge.txt`
- `crypto-base64-decode`: 有离线材料 `attachments/encoded.txt`
- `crypto-binary-decode`: 有离线材料 `attachments/challenge.txt`
- `crypto-caesar-cipher`: 有离线材料 `attachments/challenge.txt`
- `crypto-coppersmith`: 有离线材料 `attachments/params.txt`
- `crypto-crt-rsa`: 有离线材料 `attachments/params.txt`
- `crypto-dsa-weak-k`: 有离线材料 `attachments/params.txt`
- `crypto-ecdsa-nonce`: 有离线材料 `attachments/params.txt`
- `crypto-elgamal`: 有离线材料 `attachments/params.txt`
- `crypto-fault-attack`: 有离线材料 `attachments/params.txt`
- `crypto-frequency`: 有离线材料 `attachments/challenge.txt`
- `crypto-full-break`: 有离线材料 `attachments/params.txt`
- `crypto-gcm-nonce`: 有离线材料 `attachments/params.txt`
- `crypto-hex-decode`: 有离线材料 `attachments/challenge.txt`
- `crypto-hnp`: 有离线材料 `attachments/params.txt`
- `crypto-isogeny`: 有离线材料 `attachments/params.txt`
- `crypto-keyboard-shift`: 有离线材料 `attachments/challenge.txt`
- `crypto-known-plaintext`: 有离线材料 `attachments/params.txt`
- `crypto-lattice`: 有离线材料 `attachments/params.txt`
- `crypto-md5-collision`: 有离线材料 `attachments/collision-notes.txt`, `attachments/left.txt`, `attachments/right.txt`
- `crypto-meet-middle`: 有离线材料 `attachments/params.txt`
- `crypto-morse-code`: 有离线材料 `attachments/challenge.txt`
- `crypto-multi-encoding`: 有离线材料 `attachments/challenge.txt`
- `crypto-otp-reuse`: 有离线材料 `attachments/params.txt`
- `crypto-pohlig-hellman`: 有离线材料 `attachments/params.txt`
- `crypto-quantum-resistant`: 有离线材料 `attachments/params.txt`
- `crypto-rail-fence`: 有离线材料 `attachments/challenge.txt`
- `crypto-rootme-hash-md5`: 有离线材料 `attachments/hash.txt`
- `crypto-rootme-shift-cipher`: 有离线材料 `attachments/challenge.txt`
- `crypto-rootme-vigenere`: 有离线材料 `attachments/challenge.txt`
- `crypto-rsa-broadcast`: 有离线材料 `attachments/params.txt`
- `crypto-rsa-common-mod`: 有离线材料 `attachments/params.txt`
- `crypto-rsa-factor`: 有离线材料 `attachments/params.txt`
- `crypto-rsa-small-e`: 有离线材料 `attachments/params.txt`
- `crypto-rsa-wiener`: 有离线材料 `attachments/params.txt`
- `crypto-stream-reuse`: 有离线材料 `attachments/params.txt`
- `crypto-substitution`: 有离线材料 `attachments/challenge.txt`
- `crypto-url-decode`: 有离线材料 `attachments/challenge.txt`
- `crypto-vigenere`: 有离线材料 `attachments/challenge.txt`
- `crypto-weak-prng`: 有离线材料 `attachments/params.txt`
- `crypto-xor-single`: 有离线材料 `attachments/challenge.txt`

### forensics（53）

- `forensics-5g-network`: 有离线材料 `challenge.zip`
- `forensics-ai-deepfake`: 有离线材料 `challenge.zip`
- `forensics-android-backup`: 有离线材料 `challenge.zip`
- `forensics-anti-forensics`: 有离线材料 `challenge.zip`
- `forensics-apfs`: 有离线材料 `challenge.zip`
- `forensics-audio-steg`: 有离线材料 `challenge.zip`
- `forensics-autopsy`: 有离线材料 `challenge.zip`
- `forensics-bitlocker`: 有离线材料 `challenge.zip`
- `forensics-blockchain`: 有离线材料 `challenge.zip`
- `forensics-browser-history`: 有离线材料 `challenge.zip`
- `forensics-car-forensics`: 有离线材料 `challenge.zip`
- `forensics-cloud-forensics`: 有离线材料 `challenge.zip`
- `forensics-deleted-file`: 有离线材料 `challenge.zip`
- `forensics-disk-image`: 有离线材料 `challenge.zip`
- `forensics-drone`: 有离线材料 `challenge.zip`
- `forensics-email-headers`: 有离线材料 `challenge.zip`
- `forensics-encrypted-volume`: 有离线材料 `challenge.zip`
- `forensics-exif-data`: 有离线材料 `challenge.zip`
- `forensics-ext4`: 有离线材料 `challenge.zip`
- `forensics-file-signature`: 有离线材料 `challenge.zip`
- `forensics-firmware-extract`: 有离线材料 `challenge.zip`
- `forensics-hex-edit`: 有离线材料 `challenge.zip`
- `forensics-image-steg`: 有离线材料 `challenge.zip`
- `forensics-ios-backup`: 有离线材料 `challenge.zip`
- `forensics-iot-forensics`: 有离线材料 `challenge.zip`
- `forensics-log-analysis`: 有离线材料 `challenge.zip`
- `forensics-malware-analysis`: 有离线材料 `challenge.zip`
- `forensics-memory-dump`: 有离线材料 `challenge.zip`
- `forensics-metadata`: 有离线材料 `challenge.zip`
- `forensics-mft-analysis`: 有离线材料 `challenge.zip`
- `forensics-network-capture`: 有离线材料 `challenge.zip`
- `forensics-ntfs-ads`: 有离线材料 `challenge.zip`
- `forensics-office-macro`: 有离线材料 `challenge.zip`
- `forensics-pcap-basic`: 有离线材料 `challenge.zip`
- `forensics-pdf-analysis`: 有离线材料 `challenge.zip`
- `forensics-quantum-crypto`: 有离线材料 `challenge.zip`
- `forensics-raid-recovery`: 有离线材料 `challenge.zip`
- `forensics-registry-analysis`: 有离线材料 `challenge.zip`
- `forensics-rootme-deleted-file`: 有离线材料 `attachments/deleted-image.zip`
- `forensics-rootme-lost-case-mobile`: 有离线材料 `attachments/case-evidence.zip`
- `forensics-rootme-malicious-word-macro`: 有离线材料 `attachments/malicious.docm`
- `forensics-satellite`: 有离线材料 `challenge.zip`
- `forensics-scada`: 有离线材料 `challenge.zip`
- `forensics-slack-space`: 有离线材料 `challenge.zip`
- `forensics-smart-contract`: 有离线材料 `challenge.zip`
- `forensics-sqlite-db`: 有离线材料 `challenge.zip`
- `forensics-stealth-malware`: 有离线材料 `challenge.zip`
- `forensics-strings-search`: 有离线材料 `challenge.zip`
- `forensics-timeline`: 有离线材料 `challenge.zip`
- `forensics-usb-forensics`: 有离线材料 `challenge.zip`
- `forensics-vmdk-analysis`: 有离线材料 `challenge.zip`
- `forensics-volatility`: 有离线材料 `challenge.zip`
- `forensics-zip-password`: 有离线材料 `challenge.zip`

### misc（30）

- `misc-barcode`: 有离线材料 `challenge.txt`
- `misc-brainfuck`: 有离线材料 `challenge.txt`
- `misc-chip-decap`: 有离线材料 `challenge.bin`
- `misc-color-code`: 有离线材料 `challenge.txt`
- `misc-dna-computing`: 有离线材料 `challenge.txt`
- `misc-emoji-encode`: 有离线材料 `challenge.txt`
- `misc-esoteric-lang`: 有离线材料 `challenge.txt`
- `misc-fault-injection`: 有离线材料 `challenge.bin`
- `misc-fpga-reverse`: 有离线材料 `challenge.bin`
- `misc-hardware-trojan`: 有离线材料 `challenge.bin`
- `misc-lora`: 有离线材料 `challenge.bin`
- `misc-memristor`: 有离线材料 `challenge.txt`
- `misc-molecular`: 有离线材料 `challenge.txt`
- `misc-music-notes`: 有离线材料 `challenge.txt`
- `misc-neuromorphic`: 有离线材料 `challenge.txt`
- `misc-photonic`: 有离线材料 `challenge.txt`
- `misc-piet`: 有离线材料 `challenge.txt`
- `misc-power-analysis`: 有离线材料 `challenge.bin`
- `misc-qr-code`: 有离线材料 `challenge.txt`
- `misc-quantum-computing`: 有离线材料 `challenge.txt`
- `misc-reversible`: 有离线材料 `challenge.txt`
- `misc-rf-analysis`: 有离线材料 `challenge.bin`
- `misc-sdr`: 有离线材料 `challenge.bin`
- `misc-side-channel`: 有离线材料 `challenge.bin`
- `misc-spintronics`: 有离线材料 `challenge.txt`
- `misc-superconducting`: 有离线材料 `challenge.txt`
- `misc-topological`: 有离线材料 `challenge.txt`
- `misc-unicode-trick`: 有离线材料 `challenge.txt`
- `misc-whitespace`: 有离线材料 `challenge.txt`
- `misc-zigbee`: 有离线材料 `challenge.bin`

### reverse（53）

- `reverse-0day`: 有离线材料 `attachments/challenge`
- `reverse-angr`: 有离线材料 `attachments/challenge`
- `reverse-anti-debug`: 有离线材料 `attachments/challenge`
- `reverse-apk`: 有离线材料 `attachments/challenge.apk`
- `reverse-array`: 有离线材料 `attachments/challenge`
- `reverse-ascii-art`: 有离线材料 `attachments/challenge`
- `reverse-base64-check`: 有离线材料 `attachments/challenge`
- `reverse-bootloader`: 有离线材料 `attachments/challenge`
- `reverse-control-flow`: 有离线材料 `attachments/challenge`
- `reverse-dotnet`: 有离线材料 `attachments/challenge.dll.txt`
- `reverse-elf-header`: 有离线材料 `attachments/challenge`
- `reverse-firmware`: 有离线材料 `attachments/challenge`
- `reverse-flag-format`: 有离线材料 `attachments/challenge`
- `reverse-full-obf`: 有离线材料 `attachments/challenge`
- `reverse-function`: 有离线材料 `attachments/challenge`
- `reverse-gdb`: 有离线材料 `attachments/challenge`
- `reverse-ghidra`: 有离线材料 `attachments/challenge`
- `reverse-golang`: 有离线材料 `attachments/challenge`
- `reverse-hardware`: 有离线材料 `attachments/challenge`
- `reverse-hex-editor`: 有离线材料 `attachments/challenge`
- `reverse-hook`: 有离线材料 `attachments/challenge`
- `reverse-hypervisor`: 有离线材料 `attachments/challenge`
- `reverse-iat`: 有离线材料 `attachments/challenge`
- `reverse-ida`: 有离线材料 `attachments/challenge`
- `reverse-ios`: 有离线材料 `attachments/challenge.bin`
- `reverse-java`: 有离线材料 `attachments/challenge.jar`
- `reverse-kernel`: 有离线材料 `attachments/challenge`
- `reverse-llvm`: 有离线材料 `attachments/challenge`
- `reverse-loop`: 有离线材料 `attachments/challenge`
- `reverse-ltrace`: 有离线材料 `attachments/challenge`
- `reverse-malware`: 有离线材料 `attachments/challenge`
- `reverse-obfuscation`: 有离线材料 `attachments/challenge`
- `reverse-opaque`: 有离线材料 `attachments/challenge`
- `reverse-packer`: 有离线材料 `attachments/challenge`
- `reverse-password`: 有离线材料 `attachments/challenge`
- `reverse-protocol`: 有离线材料 `attachments/challenge`
- `reverse-python`: 有离线材料 `attachments/challenge.pyc`
- `reverse-rootme-apk-introduction`: 有离线材料 `attachments/challenge.apk`
- `reverse-rootme-pe-x86-0-protection`: 有离线材料 `attachments/challenge`
- `reverse-rootme-unity-mono-basic-game-hacking`: 有离线材料 `attachments/challenge`
- `reverse-rust`: 有离线材料 `attachments/challenge`
- `reverse-seh`: 有离线材料 `attachments/challenge`
- `reverse-sgx`: 有离线材料 `attachments/challenge`
- `reverse-shellcode`: 有离线材料 `attachments/challenge`
- `reverse-simple-xor`: 有离线材料 `attachments/challenge`
- `reverse-strace`: 有离线材料 `attachments/challenge`
- `reverse-string-enc`: 有离线材料 `attachments/challenge`
- `reverse-strings-basic`: 有离线材料 `attachments/challenge`
- `reverse-tls`: 有离线材料 `attachments/challenge`
- `reverse-upx`: 有离线材料 `attachments/challenge`
- `reverse-vm-basic`: 有离线材料 `attachments/challenge`
- `reverse-vmprotect`: 有离线材料 `attachments/challenge`
- `reverse-z3`: 有离线材料 `attachments/challenge`

## 仅是题卡、不能发布
