@echo off
set pool=xxflag{reverse_batch_substring_gate}yy
set out=%pool:~2,8%%pool:~10,8%%pool:~18,8%%pool:~26,8%%pool:~34,2%
if "%1"=="%out%" echo ok
