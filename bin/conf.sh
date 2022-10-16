#!/bin/bash
echo "Starting gen caddyfile..."
cat << EOF > /app/Caddyfile
:$PORT
reverse_proxy /ray 127.0.0.1:8089
EOF
chmod +x /app/caddy
chmod +x /app/helloworld
echo "Starting caddy..."
/app/caddy start -config /app/Caddyfile
echo "Starting helloworld..."
/app/helloworld run -c /app/helloworld.json > /dev/null 2>&1
