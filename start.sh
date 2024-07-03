#!/bin/bash
echo "Starting gen caddyfile..."
cat << EOF > ./Caddyfile
:80
reverse_proxy /ray 127.0.0.1:8089
EOF
chmod +x ./caddy
chmod +x ./helloworld
echo "Starting caddy..."
./caddy start -config ./Caddyfile
echo "Starting helloworld..."
./helloworld run -c ./helloworld.json
