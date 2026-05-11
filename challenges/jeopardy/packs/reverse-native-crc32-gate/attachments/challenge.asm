
/tmp/tmpidr_4e49/challenge.bin:     file format elf64-x86-64


Disassembly of section .init:

0000000000001000 <_init>:
    1000:	f3 0f 1e fa          	endbr64
    1004:	48 83 ec 08          	sub    rsp,0x8
    1008:	48 8b 05 d9 2f 00 00 	mov    rax,QWORD PTR [rip+0x2fd9]        # 3fe8 <__gmon_start__@Base>
    100f:	48 85 c0             	test   rax,rax
    1012:	74 02                	je     1016 <_init+0x16>
    1014:	ff d0                	call   rax
    1016:	48 83 c4 08          	add    rsp,0x8
    101a:	c3                   	ret

Disassembly of section .plt:

0000000000001020 <.plt>:
    1020:	ff 35 72 2f 00 00    	push   QWORD PTR [rip+0x2f72]        # 3f98 <_GLOBAL_OFFSET_TABLE_+0x8>
    1026:	ff 25 74 2f 00 00    	jmp    QWORD PTR [rip+0x2f74]        # 3fa0 <_GLOBAL_OFFSET_TABLE_+0x10>
    102c:	0f 1f 40 00          	nop    DWORD PTR [rax+0x0]
    1030:	f3 0f 1e fa          	endbr64
    1034:	68 00 00 00 00       	push   0x0
    1039:	e9 e2 ff ff ff       	jmp    1020 <_init+0x20>
    103e:	66 90                	xchg   ax,ax
    1040:	f3 0f 1e fa          	endbr64
    1044:	68 01 00 00 00       	push   0x1
    1049:	e9 d2 ff ff ff       	jmp    1020 <_init+0x20>
    104e:	66 90                	xchg   ax,ax
    1050:	f3 0f 1e fa          	endbr64
    1054:	68 02 00 00 00       	push   0x2
    1059:	e9 c2 ff ff ff       	jmp    1020 <_init+0x20>
    105e:	66 90                	xchg   ax,ax
    1060:	f3 0f 1e fa          	endbr64
    1064:	68 03 00 00 00       	push   0x3
    1069:	e9 b2 ff ff ff       	jmp    1020 <_init+0x20>
    106e:	66 90                	xchg   ax,ax
    1070:	f3 0f 1e fa          	endbr64
    1074:	68 04 00 00 00       	push   0x4
    1079:	e9 a2 ff ff ff       	jmp    1020 <_init+0x20>
    107e:	66 90                	xchg   ax,ax
    1080:	f3 0f 1e fa          	endbr64
    1084:	68 05 00 00 00       	push   0x5
    1089:	e9 92 ff ff ff       	jmp    1020 <_init+0x20>
    108e:	66 90                	xchg   ax,ax

Disassembly of section .plt.got:

0000000000001090 <__cxa_finalize@plt>:
    1090:	f3 0f 1e fa          	endbr64
    1094:	ff 25 5e 2f 00 00    	jmp    QWORD PTR [rip+0x2f5e]        # 3ff8 <__cxa_finalize@GLIBC_2.2.5>
    109a:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

Disassembly of section .plt.sec:

00000000000010a0 <puts@plt>:
    10a0:	f3 0f 1e fa          	endbr64
    10a4:	ff 25 fe 2e 00 00    	jmp    QWORD PTR [rip+0x2efe]        # 3fa8 <puts@GLIBC_2.2.5>
    10aa:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010b0 <crc32@plt>:
    10b0:	f3 0f 1e fa          	endbr64
    10b4:	ff 25 f6 2e 00 00    	jmp    QWORD PTR [rip+0x2ef6]        # 3fb0 <crc32@Base>
    10ba:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010c0 <strlen@plt>:
    10c0:	f3 0f 1e fa          	endbr64
    10c4:	ff 25 ee 2e 00 00    	jmp    QWORD PTR [rip+0x2eee]        # 3fb8 <strlen@GLIBC_2.2.5>
    10ca:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010d0 <__stack_chk_fail@plt>:
    10d0:	f3 0f 1e fa          	endbr64
    10d4:	ff 25 e6 2e 00 00    	jmp    QWORD PTR [rip+0x2ee6]        # 3fc0 <__stack_chk_fail@GLIBC_2.4>
    10da:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010e0 <strcspn@plt>:
    10e0:	f3 0f 1e fa          	endbr64
    10e4:	ff 25 de 2e 00 00    	jmp    QWORD PTR [rip+0x2ede]        # 3fc8 <strcspn@GLIBC_2.2.5>
    10ea:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010f0 <fgets@plt>:
    10f0:	f3 0f 1e fa          	endbr64
    10f4:	ff 25 d6 2e 00 00    	jmp    QWORD PTR [rip+0x2ed6]        # 3fd0 <fgets@GLIBC_2.2.5>
    10fa:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

Disassembly of section .text:

0000000000001100 <_start>:
    1100:	f3 0f 1e fa          	endbr64
    1104:	31 ed                	xor    ebp,ebp
    1106:	49 89 d1             	mov    r9,rdx
    1109:	5e                   	pop    rsi
    110a:	48 89 e2             	mov    rdx,rsp
    110d:	48 83 e4 f0          	and    rsp,0xfffffffffffffff0
    1111:	50                   	push   rax
    1112:	54                   	push   rsp
    1113:	45 31 c0             	xor    r8d,r8d
    1116:	31 c9                	xor    ecx,ecx
    1118:	48 8d 3d ca 00 00 00 	lea    rdi,[rip+0xca]        # 11e9 <main>
    111f:	ff 15 b3 2e 00 00    	call   QWORD PTR [rip+0x2eb3]        # 3fd8 <__libc_start_main@GLIBC_2.34>
    1125:	f4                   	hlt
    1126:	66 2e 0f 1f 84 00 00 	cs nop WORD PTR [rax+rax*1+0x0]
    112d:	00 00 00 

0000000000001130 <deregister_tm_clones>:
    1130:	48 8d 3d d9 2e 00 00 	lea    rdi,[rip+0x2ed9]        # 4010 <stdin@GLIBC_2.2.5>
    1137:	48 8d 05 d2 2e 00 00 	lea    rax,[rip+0x2ed2]        # 4010 <stdin@GLIBC_2.2.5>
    113e:	48 39 f8             	cmp    rax,rdi
    1141:	74 15                	je     1158 <deregister_tm_clones+0x28>
    1143:	48 8b 05 96 2e 00 00 	mov    rax,QWORD PTR [rip+0x2e96]        # 3fe0 <_ITM_deregisterTMCloneTable@Base>
    114a:	48 85 c0             	test   rax,rax
    114d:	74 09                	je     1158 <deregister_tm_clones+0x28>
    114f:	ff e0                	jmp    rax
    1151:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]
    1158:	c3                   	ret
    1159:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]

0000000000001160 <register_tm_clones>:
    1160:	48 8d 3d a9 2e 00 00 	lea    rdi,[rip+0x2ea9]        # 4010 <stdin@GLIBC_2.2.5>
    1167:	48 8d 35 a2 2e 00 00 	lea    rsi,[rip+0x2ea2]        # 4010 <stdin@GLIBC_2.2.5>
    116e:	48 29 fe             	sub    rsi,rdi
    1171:	48 89 f0             	mov    rax,rsi
    1174:	48 c1 ee 3f          	shr    rsi,0x3f
    1178:	48 c1 f8 03          	sar    rax,0x3
    117c:	48 01 c6             	add    rsi,rax
    117f:	48 d1 fe             	sar    rsi,1
    1182:	74 14                	je     1198 <register_tm_clones+0x38>
    1184:	48 8b 05 65 2e 00 00 	mov    rax,QWORD PTR [rip+0x2e65]        # 3ff0 <_ITM_registerTMCloneTable@Base>
    118b:	48 85 c0             	test   rax,rax
    118e:	74 08                	je     1198 <register_tm_clones+0x38>
    1190:	ff e0                	jmp    rax
    1192:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]
    1198:	c3                   	ret
    1199:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]

00000000000011a0 <__do_global_dtors_aux>:
    11a0:	f3 0f 1e fa          	endbr64
    11a4:	80 3d 6d 2e 00 00 00 	cmp    BYTE PTR [rip+0x2e6d],0x0        # 4018 <completed.0>
    11ab:	75 2b                	jne    11d8 <__do_global_dtors_aux+0x38>
    11ad:	55                   	push   rbp
    11ae:	48 83 3d 42 2e 00 00 	cmp    QWORD PTR [rip+0x2e42],0x0        # 3ff8 <__cxa_finalize@GLIBC_2.2.5>
    11b5:	00 
    11b6:	48 89 e5             	mov    rbp,rsp
    11b9:	74 0c                	je     11c7 <__do_global_dtors_aux+0x27>
    11bb:	48 8b 3d 46 2e 00 00 	mov    rdi,QWORD PTR [rip+0x2e46]        # 4008 <__dso_handle>
    11c2:	e8 c9 fe ff ff       	call   1090 <__cxa_finalize@plt>
    11c7:	e8 64 ff ff ff       	call   1130 <deregister_tm_clones>
    11cc:	c6 05 45 2e 00 00 01 	mov    BYTE PTR [rip+0x2e45],0x1        # 4018 <completed.0>
    11d3:	5d                   	pop    rbp
    11d4:	c3                   	ret
    11d5:	0f 1f 00             	nop    DWORD PTR [rax]
    11d8:	c3                   	ret
    11d9:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]

00000000000011e0 <frame_dummy>:
    11e0:	f3 0f 1e fa          	endbr64
    11e4:	e9 77 ff ff ff       	jmp    1160 <register_tm_clones>

00000000000011e9 <main>:
    11e9:	f3 0f 1e fa          	endbr64
    11ed:	55                   	push   rbp
    11ee:	48 89 e5             	mov    rbp,rsp
    11f1:	48 83 ec 30          	sub    rsp,0x30
    11f5:	64 48 8b 04 25 28 00 	mov    rax,QWORD PTR fs:0x28
    11fc:	00 00 
    11fe:	48 89 45 f8          	mov    QWORD PTR [rbp-0x8],rax
    1202:	31 c0                	xor    eax,eax
    1204:	48 c7 45 d0 00 00 00 	mov    QWORD PTR [rbp-0x30],0x0
    120b:	00 
    120c:	48 c7 45 d8 00 00 00 	mov    QWORD PTR [rbp-0x28],0x0
    1213:	00 
    1214:	48 c7 45 e0 00 00 00 	mov    QWORD PTR [rbp-0x20],0x0
    121b:	00 
    121c:	48 c7 45 e8 00 00 00 	mov    QWORD PTR [rbp-0x18],0x0
    1223:	00 
    1224:	48 8b 15 e5 2d 00 00 	mov    rdx,QWORD PTR [rip+0x2de5]        # 4010 <stdin@GLIBC_2.2.5>
    122b:	48 8d 45 d0          	lea    rax,[rbp-0x30]
    122f:	be 20 00 00 00       	mov    esi,0x20
    1234:	48 89 c7             	mov    rdi,rax
    1237:	e8 b4 fe ff ff       	call   10f0 <fgets@plt>
    123c:	48 85 c0             	test   rax,rax
    123f:	75 07                	jne    1248 <main+0x5f>
    1241:	b8 01 00 00 00       	mov    eax,0x1
    1246:	eb 70                	jmp    12b8 <main+0xcf>
    1248:	48 8d 45 d0          	lea    rax,[rbp-0x30]
    124c:	48 8d 15 b1 0d 00 00 	lea    rdx,[rip+0xdb1]        # 2004 <_IO_stdin_used+0x4>
    1253:	48 89 d6             	mov    rsi,rdx
    1256:	48 89 c7             	mov    rdi,rax
    1259:	e8 82 fe ff ff       	call   10e0 <strcspn@plt>
    125e:	c6 44 05 d0 00       	mov    BYTE PTR [rbp+rax*1-0x30],0x0
    1263:	48 8d 45 d0          	lea    rax,[rbp-0x30]
    1267:	48 89 c7             	mov    rdi,rax
    126a:	e8 51 fe ff ff       	call   10c0 <strlen@plt>
    126f:	48 83 f8 04          	cmp    rax,0x4
    1273:	75 2f                	jne    12a4 <main+0xbb>
    1275:	48 8d 45 d0          	lea    rax,[rbp-0x30]
    1279:	ba 04 00 00 00       	mov    edx,0x4
    127e:	48 89 c6             	mov    rsi,rax
    1281:	bf 00 00 00 00       	mov    edi,0x0
    1286:	e8 25 fe ff ff       	call   10b0 <crc32@plt>
    128b:	48 3d f3 b4 6f 17    	cmp    rax,0x176fb4f3
    1291:	75 11                	jne    12a4 <main+0xbb>
    1293:	48 8d 05 6c 0d 00 00 	lea    rax,[rip+0xd6c]        # 2006 <_IO_stdin_used+0x6>
    129a:	48 89 c7             	mov    rdi,rax
    129d:	e8 fe fd ff ff       	call   10a0 <puts@plt>
    12a2:	eb 0f                	jmp    12b3 <main+0xca>
    12a4:	48 8d 05 5e 0d 00 00 	lea    rax,[rip+0xd5e]        # 2009 <_IO_stdin_used+0x9>
    12ab:	48 89 c7             	mov    rdi,rax
    12ae:	e8 ed fd ff ff       	call   10a0 <puts@plt>
    12b3:	b8 00 00 00 00       	mov    eax,0x0
    12b8:	48 8b 55 f8          	mov    rdx,QWORD PTR [rbp-0x8]
    12bc:	64 48 2b 14 25 28 00 	sub    rdx,QWORD PTR fs:0x28
    12c3:	00 00 
    12c5:	74 05                	je     12cc <main+0xe3>
    12c7:	e8 04 fe ff ff       	call   10d0 <__stack_chk_fail@plt>
    12cc:	c9                   	leave
    12cd:	c3                   	ret

Disassembly of section .fini:

00000000000012d0 <_fini>:
    12d0:	f3 0f 1e fa          	endbr64
    12d4:	48 83 ec 08          	sub    rsp,0x8
    12d8:	48 83 c4 08          	add    rsp,0x8
    12dc:	c3                   	ret
