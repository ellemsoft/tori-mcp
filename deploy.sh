#!/bin/bash
# Deploy tori + oikotie MCP to Hetzner VPS
# Run from ~/Development/tori-cli (or any dir with built binaries)
set -euo pipefail

HOST="${TORI_HOST:-}"
USER="root"
REMOTE_TORI="/opt/tori"
REMOTE_OIKOTIE="/opt/oikotie"

echo "=== Building binaries for Linux ==="
cd ~/Development/tori-cli
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o tori .
cd ~/Development/oikotie-cli
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o oikotie .

echo "=== Uploading binaries ==="
ssh "$USER@$HOST" "mkdir -p $REMOTE_TORI $REMOTE_OIKOTIE"
scp ~/Development/tori-cli/tori "$USER@$HOST:$REMOTE_TORI/"
scp ~/Development/oikotie-cli/oikotie "$USER@$HOST:$REMOTE_OIKOTIE/"

echo "=== Creating system users ==="
ssh "$USER@$HOST" << 'SSH'
  id tori &>/dev/null || useradd -r -s /bin/false -d /opt/tori tori
  id oikotie &>/dev/null || useradd -r -s /bin/false -d /opt/oikotie oikotie
  chown -R tori:tori /opt/tori
  chown -R oikotie:oikotie /opt/oikotie
  chmod +x /opt/tori/tori /opt/oikotie/oikotie
SSH

echo "=== Installing systemd services ==="
ssh "$USER@$HOST" << 'SSH'
  cat > /etc/systemd/system/tori-mcp.service << 'UNIT'
[Unit]
Description=Tori.fi MCP
After=network-online.target
Wants=network-online.target
[Service]
Type=simple
User=tori
Group=tori
WorkingDirectory=/opt/tori
ExecStart=/opt/tori/tori --mcp 127.0.0.1:8081
Restart=always
RestartSec=5s
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
LimitNOFILE=4096
StandardOutput=journal
StandardError=journal
SyslogIdentifier=tori-mcp
[Install]
WantedBy=multi-user.target
UNIT

  cat > /etc/systemd/system/oikotie-mcp.service << 'UNIT'
[Unit]
Description=Oikotie.fi MCP
After=network-online.target
Wants=network-online.target
[Service]
Type=simple
User=oikotie
Group=oikotie
WorkingDirectory=/opt/oikotie
ExecStart=/opt/oikotie/oikotie --mcp 127.0.0.1:8082
Restart=always
RestartSec=5s
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
LimitNOFILE=4096
StandardOutput=journal
StandardError=journal
SyslogIdentifier=oikotie-mcp
[Install]
WantedBy=multi-user.target
UNIT

  systemctl daemon-reload
  systemctl enable --now tori-mcp oikotie-mcp
SSH

echo "=== Installing Caddy configs ==="
ssh "$USER@$HOST" << 'SSH'
  cat > /etc/caddy/apps-enabled/tori.caddy << 'CADDY'
tori.ellemsoft.com {
	handle /mcp {
		reverse_proxy 127.0.0.1:8081
	}
	handle {
		respond "tori-mcp — /mcp endpoint" 200
	}
}
CADDY

  cat > /etc/caddy/apps-enabled/oikotie.caddy << 'CADDY'
oikotie.ellemsoft.com {
	handle /mcp {
		reverse_proxy 127.0.0.1:8082
	}
	handle {
		respond "oikotie-mcp — /mcp endpoint" 200
	}
}
CADDY

  systemctl reload caddy
SSH

echo "=== Checking services ==="
ssh "$USER@$HOST" << 'SSH'
  systemctl is-active tori-mcp && echo "tori-mcp: running" || echo "tori-mcp: FAILED"
  systemctl is-active oikotie-mcp && echo "oikotie-mcp: running" || echo "oikotie-mcp: FAILED"
  sleep 1
  curl -s http://127.0.0.1:8081/health && echo " tori health OK"
  curl -s http://127.0.0.1:8082/health && echo " oikotie health OK"
SSH

echo ""
echo "=== Done ==="
echo "MCP URLs:"
echo "  https://tori.ellemsoft.com/mcp"
echo "  https://oikotie.ellemsoft.com/mcp"
echo ""
echo "Claude Desktop config:"
echo '  {"mcpServers": {'
echo '    "tori":    {"url": "https://tori.ellemsoft.com/mcp"},'
echo '    "oikotie": {"url": "https://oikotie.ellemsoft.com/mcp"}'
echo '  }}'
