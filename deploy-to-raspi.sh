#!/bin/bash
# Fleet Officer - Raspberry Pi Deployment Script
# Version: 1.0
# Autor: JavaFleet Systems Consulting

set -e  # Exit on error

echo "🚢 Fleet Officer - Raspberry Pi Deployment"
echo "==========================================="
echo ""

# === KONFIGURATION ===
# Passen Sie diese Werte an!

# Raspberry Pi Details
PI_USER="${PI_USER:-pi}"
PI_HOST="${PI_HOST}"
PI_PORT="${PI_PORT:-22}"

# Fleet Navigator Details
NAVIGATOR_IP="${NAVIGATOR_IP}"
NAVIGATOR_PORT="${NAVIGATOR_PORT:-2025}"

# Officer Details
OFFICER_ID="${OFFICER_ID:-raspberry-pi-01}"
OFFICER_NAME="${OFFICER_NAME:-Raspberry Pi}"
OFFICER_DESC="${OFFICER_DESC:-Fleet Officer Monitoring Agent}"

# Binary Auswahl (auto-detect oder manuell setzen)
ARCH="${ARCH:-auto}"  # auto, arm64, armv7

# === VALIDIERUNG ===

if [ -z "$PI_HOST" ]; then
    echo "❌ Fehler: PI_HOST nicht gesetzt!"
    echo ""
    echo "Verwendung:"
    echo "  PI_HOST=192.168.1.100 NAVIGATOR_IP=192.168.1.50 ./deploy-to-raspi.sh"
    echo ""
    echo "Oder setzen Sie Umgebungsvariablen:"
    echo "  export PI_HOST=192.168.1.100"
    echo "  export PI_USER=pi"
    echo "  export NAVIGATOR_IP=192.168.1.50"
    echo "  export OFFICER_ID=raspi-wohnzimmer"
    echo "  export OFFICER_NAME='Raspberry Pi Wohnzimmer'"
    echo "  ./deploy-to-raspi.sh"
    exit 1
fi

if [ -z "$NAVIGATOR_IP" ]; then
    echo "❌ Fehler: NAVIGATOR_IP nicht gesetzt!"
    echo "Beispiel: NAVIGATOR_IP=192.168.1.50"
    exit 1
fi

echo "📋 Konfiguration:"
echo "  Raspberry Pi: $PI_USER@$PI_HOST:$PI_PORT"
echo "  Navigator:    ws://$NAVIGATOR_IP:$NAVIGATOR_PORT"
echo "  Officer ID:   $OFFICER_ID"
echo "  Officer Name: $OFFICER_NAME"
echo ""

# === ARCHITEKTUR ERKENNEN ===

if [ "$ARCH" == "auto" ]; then
    echo "🔍 Erkenne Raspberry Pi Architektur..."

    PI_ARCH=$(ssh -p $PI_PORT $PI_USER@$PI_HOST "uname -m" 2>/dev/null || echo "unknown")

    case "$PI_ARCH" in
        aarch64)
            ARCH="arm64"
            BINARY="fleet-officer-arm64"
            echo "✅ Erkannt: ARM64 (64-bit)"
            ;;
        armv7l|armv6l)
            ARCH="armv7"
            BINARY="fleet-officer-armv7"
            echo "✅ Erkannt: ARMv7 (32-bit)"
            ;;
        *)
            echo "❌ Fehler: Unbekannte Architektur: $PI_ARCH"
            echo "Bitte setzen Sie ARCH manuell: ARCH=arm64 oder ARCH=armv7"
            exit 1
            ;;
    esac
else
    case "$ARCH" in
        arm64)
            BINARY="fleet-officer-arm64"
            echo "✅ Manuell gewählt: ARM64"
            ;;
        armv7)
            BINARY="fleet-officer-armv7"
            echo "✅ Manuell gewählt: ARMv7"
            ;;
        *)
            echo "❌ Fehler: Ungültige Architektur: $ARCH"
            echo "Erlaubt: arm64, armv7"
            exit 1
            ;;
    esac
fi

echo ""

# === BINARY PRÜFEN ===

if [ ! -f "$BINARY" ]; then
    echo "❌ Fehler: Binary nicht gefunden: $BINARY"
    echo ""
    echo "Bitte kompilieren Sie die Binary zuerst:"
    if [ "$ARCH" == "arm64" ]; then
        echo "  GOOS=linux GOARCH=arm64 go build -o fleet-officer-arm64 main.go"
    else
        echo "  GOOS=linux GOARCH=arm GOARM=7 go build -o fleet-officer-armv7 main.go"
    fi
    exit 1
fi

echo "✅ Binary gefunden: $BINARY ($(ls -lh $BINARY | awk '{print $5}'))"
echo ""

# === SSH-VERBINDUNG TESTEN ===

echo "🔌 Teste SSH-Verbindung..."
if ! ssh -p $PI_PORT -o ConnectTimeout=5 $PI_USER@$PI_HOST "echo 'SSH OK'" > /dev/null 2>&1; then
    echo "❌ Fehler: Kann keine SSH-Verbindung zu $PI_USER@$PI_HOST:$PI_PORT herstellen!"
    echo ""
    echo "Tipps:"
    echo "  - Ist SSH auf dem Pi aktiviert?"
    echo "  - Ist die IP-Adresse korrekt?"
    echo "  - Ist der Pi im Netzwerk erreichbar? (ping $PI_HOST)"
    exit 1
fi
echo "✅ SSH-Verbindung OK"
echo ""

# === DEPLOYMENT ===

echo "📦 Starte Deployment..."
echo ""

# 1. Binary kopieren
echo "1️⃣  Kopiere Binary auf Pi..."
scp -P $PI_PORT $BINARY $PI_USER@$PI_HOST:/home/$PI_USER/fleet-officer
echo "   ✅ Binary kopiert"
echo ""

# 2. Config erstellen
echo "2️⃣  Erstelle config.yml..."
ssh -p $PI_PORT $PI_USER@$PI_HOST "cat > /home/$PI_USER/config.yml" << EOF
officer:
  id: "$OFFICER_ID"
  name: "$OFFICER_NAME"
  description: "$OFFICER_DESC"

navigator:
  url: "ws://$NAVIGATOR_IP:$NAVIGATOR_PORT/api/fleet-officer/ws"
  max_reconnect_attempts: 0
  reconnect_interval: 10s

monitoring:
  interval: 5s
  enabled: true
  collect_cpu: true
  collect_memory: true
  collect_disk: true
  collect_network: true
  collect_temperature: true
  collect_processes: true
EOF
echo "   ✅ Config erstellt"
echo ""

# 3. Ausführbar machen
echo "3️⃣  Setze Berechtigungen..."
ssh -p $PI_PORT $PI_USER@$PI_HOST "chmod +x /home/$PI_USER/fleet-officer"
echo "   ✅ Binary ist ausführbar"
echo ""

# 4. systemd Service erstellen
echo "4️⃣  Erstelle systemd Service..."
ssh -p $PI_PORT $PI_USER@$PI_HOST "sudo tee /etc/systemd/system/fleet-officer.service > /dev/null" << EOF
[Unit]
Description=Fleet Officer - Hardware Monitoring Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$PI_USER
WorkingDirectory=/home/$PI_USER
ExecStart=/home/$PI_USER/fleet-officer
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
echo "   ✅ Service erstellt"
echo ""

# 5. Service aktivieren und starten
echo "5️⃣  Aktiviere und starte Service..."
ssh -p $PI_PORT $PI_USER@$PI_HOST "sudo systemctl daemon-reload"
ssh -p $PI_PORT $PI_USER@$PI_HOST "sudo systemctl enable fleet-officer.service"
ssh -p $PI_PORT $PI_USER@$PI_HOST "sudo systemctl restart fleet-officer.service"
echo "   ✅ Service läuft"
echo ""

# 6. Status prüfen
echo "6️⃣  Prüfe Service-Status..."
sleep 2
STATUS=$(ssh -p $PI_PORT $PI_USER@$PI_HOST "sudo systemctl is-active fleet-officer.service")

if [ "$STATUS" == "active" ]; then
    echo "   ✅ Service ist AKTIV"
else
    echo "   ⚠️  Service-Status: $STATUS"
fi
echo ""

# === ZUSAMMENFASSUNG ===

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 Deployment abgeschlossen!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Details:"
echo "  Raspberry Pi:  $PI_USER@$PI_HOST"
echo "  Officer ID:    $OFFICER_ID"
echo "  Navigator:     http://$NAVIGATOR_IP:$NAVIGATOR_PORT"
echo "  Service:       fleet-officer.service"
echo ""
echo "🔧 Nützliche Befehle (auf dem Pi):"
echo "  Status:        sudo systemctl status fleet-officer"
echo "  Logs (live):   sudo journalctl -u fleet-officer -f"
echo "  Logs (last):   sudo journalctl -u fleet-officer -n 50"
echo "  Restart:       sudo systemctl restart fleet-officer"
echo "  Stop:          sudo systemctl stop fleet-officer"
echo ""
echo "🌐 Fleet Navigator öffnen:"
echo "  http://$NAVIGATOR_IP:$NAVIGATOR_PORT"
echo "  → Fleet Officers Dashboard"
echo "  → Suche nach: $OFFICER_NAME"
echo ""
echo "💡 Letzte Schritte:"
echo "  1. Öffnen Sie Fleet Navigator im Browser"
echo "  2. Klicken Sie auf das Server-Icon (orange)"
echo "  3. Ihr neuer Officer sollte online erscheinen!"
echo ""

# Optional: Logs anzeigen
read -p "Möchten Sie die Logs anzeigen? (j/N): " SHOW_LOGS
if [[ "$SHOW_LOGS" =~ ^[jJyY]$ ]]; then
    echo ""
    echo "📋 Letzte 20 Log-Zeilen:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    ssh -p $PI_PORT $PI_USER@$PI_HOST "sudo journalctl -u fleet-officer -n 20 --no-pager"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

echo ""
echo "✅ Fertig! 🚢🍓"
