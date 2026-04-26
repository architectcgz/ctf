#!/bin/sh
set -eu

exec socat TCP-LISTEN:8080,reuseaddr,fork EXEC:/opt/chal/challenge,stderr
