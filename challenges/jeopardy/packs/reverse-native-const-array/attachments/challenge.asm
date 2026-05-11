
/tmp/tmpkpu1llv9/challenge.bin:     file format elf64-x86-64


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
    1020:	ff 35 7a 2f 00 00    	push   QWORD PTR [rip+0x2f7a]        # 3fa0 <_GLOBAL_OFFSET_TABLE_+0x8>
    1026:	ff 25 7c 2f 00 00    	jmp    QWORD PTR [rip+0x2f7c]        # 3fa8 <_GLOBAL_OFFSET_TABLE_+0x10>
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

Disassembly of section .plt.got:

0000000000001080 <__cxa_finalize@plt>:
    1080:	f3 0f 1e fa          	endbr64
    1084:	ff 25 6e 2f 00 00    	jmp    QWORD PTR [rip+0x2f6e]        # 3ff8 <__cxa_finalize@GLIBC_2.2.5>
    108a:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

Disassembly of section .plt.sec:

0000000000001090 <puts@plt>:
    1090:	f3 0f 1e fa          	endbr64
    1094:	ff 25 16 2f 00 00    	jmp    QWORD PTR [rip+0x2f16]        # 3fb0 <puts@GLIBC_2.2.5>
    109a:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010a0 <strlen@plt>:
    10a0:	f3 0f 1e fa          	endbr64
    10a4:	ff 25 0e 2f 00 00    	jmp    QWORD PTR [rip+0x2f0e]        # 3fb8 <strlen@GLIBC_2.2.5>
    10aa:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010b0 <__stack_chk_fail@plt>:
    10b0:	f3 0f 1e fa          	endbr64
    10b4:	ff 25 06 2f 00 00    	jmp    QWORD PTR [rip+0x2f06]        # 3fc0 <__stack_chk_fail@GLIBC_2.4>
    10ba:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010c0 <strcspn@plt>:
    10c0:	f3 0f 1e fa          	endbr64
    10c4:	ff 25 fe 2e 00 00    	jmp    QWORD PTR [rip+0x2efe]        # 3fc8 <strcspn@GLIBC_2.2.5>
    10ca:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

00000000000010d0 <fgets@plt>:
    10d0:	f3 0f 1e fa          	endbr64
    10d4:	ff 25 f6 2e 00 00    	jmp    QWORD PTR [rip+0x2ef6]        # 3fd0 <fgets@GLIBC_2.2.5>
    10da:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]

Disassembly of section .text:

00000000000010e0 <_start>:
    10e0:	f3 0f 1e fa          	endbr64
    10e4:	31 ed                	xor    ebp,ebp
    10e6:	49 89 d1             	mov    r9,rdx
    10e9:	5e                   	pop    rsi
    10ea:	48 89 e2             	mov    rdx,rsp
    10ed:	48 83 e4 f0          	and    rsp,0xfffffffffffffff0
    10f1:	50                   	push   rax
    10f2:	54                   	push   rsp
    10f3:	45 31 c0             	xor    r8d,r8d
    10f6:	31 c9                	xor    ecx,ecx
    10f8:	48 8d 3d ca 00 00 00 	lea    rdi,[rip+0xca]        # 11c9 <main>
    10ff:	ff 15 d3 2e 00 00    	call   QWORD PTR [rip+0x2ed3]        # 3fd8 <__libc_start_main@GLIBC_2.34>
    1105:	f4                   	hlt
    1106:	66 2e 0f 1f 84 00 00 	cs nop WORD PTR [rax+rax*1+0x0]
    110d:	00 00 00 

0000000000001110 <deregister_tm_clones>:
    1110:	48 8d 3d f9 2e 00 00 	lea    rdi,[rip+0x2ef9]        # 4010 <stdin@GLIBC_2.2.5>
    1117:	48 8d 05 f2 2e 00 00 	lea    rax,[rip+0x2ef2]        # 4010 <stdin@GLIBC_2.2.5>
    111e:	48 39 f8             	cmp    rax,rdi
    1121:	74 15                	je     1138 <deregister_tm_clones+0x28>
    1123:	48 8b 05 b6 2e 00 00 	mov    rax,QWORD PTR [rip+0x2eb6]        # 3fe0 <_ITM_deregisterTMCloneTable@Base>
    112a:	48 85 c0             	test   rax,rax
    112d:	74 09                	je     1138 <deregister_tm_clones+0x28>
    112f:	ff e0                	jmp    rax
    1131:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]
    1138:	c3                   	ret
    1139:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]

0000000000001140 <register_tm_clones>:
    1140:	48 8d 3d c9 2e 00 00 	lea    rdi,[rip+0x2ec9]        # 4010 <stdin@GLIBC_2.2.5>
    1147:	48 8d 35 c2 2e 00 00 	lea    rsi,[rip+0x2ec2]        # 4010 <stdin@GLIBC_2.2.5>
    114e:	48 29 fe             	sub    rsi,rdi
    1151:	48 89 f0             	mov    rax,rsi
    1154:	48 c1 ee 3f          	shr    rsi,0x3f
    1158:	48 c1 f8 03          	sar    rax,0x3
    115c:	48 01 c6             	add    rsi,rax
    115f:	48 d1 fe             	sar    rsi,1
    1162:	74 14                	je     1178 <register_tm_clones+0x38>
    1164:	48 8b 05 85 2e 00 00 	mov    rax,QWORD PTR [rip+0x2e85]        # 3ff0 <_ITM_registerTMCloneTable@Base>
    116b:	48 85 c0             	test   rax,rax
    116e:	74 08                	je     1178 <register_tm_clones+0x38>
    1170:	ff e0                	jmp    rax
    1172:	66 0f 1f 44 00 00    	nop    WORD PTR [rax+rax*1+0x0]
    1178:	c3                   	ret
    1179:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]

0000000000001180 <__do_global_dtors_aux>:
    1180:	f3 0f 1e fa          	endbr64
    1184:	80 3d 8d 2e 00 00 00 	cmp    BYTE PTR [rip+0x2e8d],0x0        # 4018 <completed.0>
    118b:	75 2b                	jne    11b8 <__do_global_dtors_aux+0x38>
    118d:	55                   	push   rbp
    118e:	48 83 3d 62 2e 00 00 	cmp    QWORD PTR [rip+0x2e62],0x0        # 3ff8 <__cxa_finalize@GLIBC_2.2.5>
    1195:	00 
    1196:	48 89 e5             	mov    rbp,rsp
    1199:	74 0c                	je     11a7 <__do_global_dtors_aux+0x27>
    119b:	48 8b 3d 66 2e 00 00 	mov    rdi,QWORD PTR [rip+0x2e66]        # 4008 <__dso_handle>
    11a2:	e8 d9 fe ff ff       	call   1080 <__cxa_finalize@plt>
    11a7:	e8 64 ff ff ff       	call   1110 <deregister_tm_clones>
    11ac:	c6 05 65 2e 00 00 01 	mov    BYTE PTR [rip+0x2e65],0x1        # 4018 <completed.0>
    11b3:	5d                   	pop    rbp
    11b4:	c3                   	ret
    11b5:	0f 1f 00             	nop    DWORD PTR [rax]
    11b8:	c3                   	ret
    11b9:	0f 1f 80 00 00 00 00 	nop    DWORD PTR [rax+0x0]

00000000000011c0 <frame_dummy>:
    11c0:	f3 0f 1e fa          	endbr64
    11c4:	e9 77 ff ff ff       	jmp    1140 <register_tm_clones>

00000000000011c9 <main>:
    11c9:	f3 0f 1e fa          	endbr64
    11cd:	55                   	push   rbp
    11ce:	48 89 e5             	mov    rbp,rsp
    11d1:	48 81 ec a0 00 00 00 	sub    rsp,0xa0
    11d8:	64 48 8b 04 25 28 00 	mov    rax,QWORD PTR fs:0x28
    11df:	00 00 
    11e1:	48 89 45 f8          	mov    QWORD PTR [rbp-0x8],rax
    11e5:	31 c0                	xor    eax,eax
    11e7:	48 c7 85 70 ff ff ff 	mov    QWORD PTR [rbp-0x90],0x0
    11ee:	00 00 00 00 
    11f2:	48 c7 85 78 ff ff ff 	mov    QWORD PTR [rbp-0x88],0x0
    11f9:	00 00 00 00 
    11fd:	48 c7 45 80 00 00 00 	mov    QWORD PTR [rbp-0x80],0x0
    1204:	00 
    1205:	48 c7 45 88 00 00 00 	mov    QWORD PTR [rbp-0x78],0x0
    120c:	00 
    120d:	48 c7 45 90 00 00 00 	mov    QWORD PTR [rbp-0x70],0x0
    1214:	00 
    1215:	48 c7 45 98 00 00 00 	mov    QWORD PTR [rbp-0x68],0x0
    121c:	00 
    121d:	48 c7 45 a0 00 00 00 	mov    QWORD PTR [rbp-0x60],0x0
    1224:	00 
    1225:	48 c7 45 a8 00 00 00 	mov    QWORD PTR [rbp-0x58],0x0
    122c:	00 
    122d:	48 c7 45 b0 00 00 00 	mov    QWORD PTR [rbp-0x50],0x0
    1234:	00 
    1235:	48 c7 45 b8 00 00 00 	mov    QWORD PTR [rbp-0x48],0x0
    123c:	00 
    123d:	48 c7 45 c0 00 00 00 	mov    QWORD PTR [rbp-0x40],0x0
    1244:	00 
    1245:	48 c7 45 c8 00 00 00 	mov    QWORD PTR [rbp-0x38],0x0
    124c:	00 
    124d:	48 c7 45 d0 00 00 00 	mov    QWORD PTR [rbp-0x30],0x0
    1254:	00 
    1255:	48 c7 45 d8 00 00 00 	mov    QWORD PTR [rbp-0x28],0x0
    125c:	00 
    125d:	48 c7 45 e0 00 00 00 	mov    QWORD PTR [rbp-0x20],0x0
    1264:	00 
    1265:	48 c7 45 e8 00 00 00 	mov    QWORD PTR [rbp-0x18],0x0
    126c:	00 
    126d:	48 8b 15 9c 2d 00 00 	mov    rdx,QWORD PTR [rip+0x2d9c]        # 4010 <stdin@GLIBC_2.2.5>
    1274:	48 8d 85 70 ff ff ff 	lea    rax,[rbp-0x90]
    127b:	be 80 00 00 00       	mov    esi,0x80
    1280:	48 89 c7             	mov    rdi,rax
    1283:	e8 48 fe ff ff       	call   10d0 <fgets@plt>
    1288:	48 85 c0             	test   rax,rax
    128b:	75 0a                	jne    1297 <main+0xce>
    128d:	b8 01 00 00 00       	mov    eax,0x1
    1292:	e9 d5 00 00 00       	jmp    136c <main+0x1a3>
    1297:	48 8d 85 70 ff ff ff 	lea    rax,[rbp-0x90]
    129e:	48 8d 15 9b 0d 00 00 	lea    rdx,[rip+0xd9b]        # 2040 <encoded+0x20>
    12a5:	48 89 d6             	mov    rsi,rdx
    12a8:	48 89 c7             	mov    rdi,rax
    12ab:	e8 10 fe ff ff       	call   10c0 <strcspn@plt>
    12b0:	c6 84 05 70 ff ff ff 	mov    BYTE PTR [rbp+rax*1-0x90],0x0
    12b7:	00 
    12b8:	48 8d 85 70 ff ff ff 	lea    rax,[rbp-0x90]
    12bf:	48 89 c7             	mov    rdi,rax
    12c2:	e8 d9 fd ff ff       	call   10a0 <strlen@plt>
    12c7:	48 83 f8 20          	cmp    rax,0x20
    12cb:	74 19                	je     12e6 <main+0x11d>
    12cd:	48 8d 05 6e 0d 00 00 	lea    rax,[rip+0xd6e]        # 2042 <encoded+0x22>
    12d4:	48 89 c7             	mov    rdi,rax
    12d7:	e8 b4 fd ff ff       	call   1090 <puts@plt>
    12dc:	b8 01 00 00 00       	mov    eax,0x1
    12e1:	e9 86 00 00 00       	jmp    136c <main+0x1a3>
    12e6:	48 c7 85 68 ff ff ff 	mov    QWORD PTR [rbp-0x98],0x0
    12ed:	00 00 00 00 
    12f1:	eb 5b                	jmp    134e <main+0x185>
    12f3:	48 8d 95 70 ff ff ff 	lea    rdx,[rbp-0x90]
    12fa:	48 8b 85 68 ff ff ff 	mov    rax,QWORD PTR [rbp-0x98]
    1301:	48 01 d0             	add    rax,rdx
    1304:	0f b6 00             	movzx  eax,BYTE PTR [rax]
    1307:	0f b6 d0             	movzx  edx,al
    130a:	48 8b 85 68 ff ff ff 	mov    rax,QWORD PTR [rbp-0x98]
    1311:	48 01 c2             	add    rdx,rax
    1314:	48 8d 0d 05 0d 00 00 	lea    rcx,[rip+0xd05]        # 2020 <encoded>
    131b:	48 8b 85 68 ff ff ff 	mov    rax,QWORD PTR [rbp-0x98]
    1322:	48 01 c8             	add    rax,rcx
    1325:	0f b6 00             	movzx  eax,BYTE PTR [rax]
    1328:	0f b6 c0             	movzx  eax,al
    132b:	48 39 c2             	cmp    rdx,rax
    132e:	74 16                	je     1346 <main+0x17d>
    1330:	48 8d 05 0b 0d 00 00 	lea    rax,[rip+0xd0b]        # 2042 <encoded+0x22>
    1337:	48 89 c7             	mov    rdi,rax
    133a:	e8 51 fd ff ff       	call   1090 <puts@plt>
    133f:	b8 01 00 00 00       	mov    eax,0x1
    1344:	eb 26                	jmp    136c <main+0x1a3>
    1346:	48 83 85 68 ff ff ff 	add    QWORD PTR [rbp-0x98],0x1
    134d:	01 
    134e:	48 83 bd 68 ff ff ff 	cmp    QWORD PTR [rbp-0x98],0x1f
    1355:	1f 
    1356:	76 9b                	jbe    12f3 <main+0x12a>
    1358:	48 8d 05 e8 0c 00 00 	lea    rax,[rip+0xce8]        # 2047 <encoded+0x27>
    135f:	48 89 c7             	mov    rdi,rax
    1362:	e8 29 fd ff ff       	call   1090 <puts@plt>
    1367:	b8 00 00 00 00       	mov    eax,0x0
    136c:	48 8b 55 f8          	mov    rdx,QWORD PTR [rbp-0x8]
    1370:	64 48 2b 14 25 28 00 	sub    rdx,QWORD PTR fs:0x28
    1377:	00 00 
    1379:	74 05                	je     1380 <main+0x1b7>
    137b:	e8 30 fd ff ff       	call   10b0 <__stack_chk_fail@plt>
    1380:	c9                   	leave
    1381:	c3                   	ret

Disassembly of section .fini:

0000000000001384 <_fini>:
    1384:	f3 0f 1e fa          	endbr64
    1388:	48 83 ec 08          	sub    rsp,0x8
    138c:	48 83 c4 08          	add    rsp,0x8
    1390:	c3                   	ret
