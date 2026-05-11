#!/bin/sh
pool="aaflag{reverse_shell_parameter_maze}zz"
left="${pool#??}"
right="${left%??}"
if [ "$1" = "$right" ]; then
  echo ok
fi
