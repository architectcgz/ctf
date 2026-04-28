#!/bin/sh
set -eu
python3 -m http.server 80 --directory /srv/site &
exec socat TCP-LISTEN:9999,reuseaddr,fork EXEC:/opt/chal/challenge,stderr
