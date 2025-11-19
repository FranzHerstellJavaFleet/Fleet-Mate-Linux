# Fleet Mate Linux

**Version:** 1.1.0
**Platform:** Linux x86_64 / ARM64 (Raspberry Pi)

Fleet Mate ist ein autonomer Hardware-Monitoring-Agent, der von Fleet Navigator gesteuert wird.

---

## 🎯 Funktionen

- ✅ **CPU-Monitoring**: Auslastung, Kerne, Modell, Frequenz
- ✅ **RAM-Monitoring**: Total, Used, Free, Swap
- ✅ **GPU-Monitoring**: NVIDIA GPU/VRAM Auslastung, Temperatur (NEU in v1.1.0)
- ✅ **Disk-Monitoring**: Mount Points, Usage, Free Space
- ✅ **Temperatur-Monitoring**: CPU/System/GPU Temperaturen
- ✅ **Netzwerk-Monitoring**: Traffic, Errors, Interfaces
- ✅ **WebSocket**: Echtzeit-Kommunikation mit Fleet Navigator
- ✅ **Auto-Reconnect**: Automatische Wiederverbindung
- ✅ **YAML Konfiguration**: Flexibel konfigurierbar

---

## 🚀 Installation

### 1. Build

```bash
# Build für Linux x86_64
go build -o fleet-mate main.go

# Build für Raspberry Pi (ARM64)
GOOS=linux GOARCH=arm64 go build -o fleet-mate-arm64 main.go
```

### 2. Konfiguration

Bearbeiten Sie `config.yml`:

```yaml
mate:
  id: "ubuntu-desktop-01"          # Eindeutige ID
  name: "Ubuntu Desktop Trainer"

navigator:
  url: "ws://localhost:2025/api/fleet-mate/ws"

monitoring:
  interval: 5s                      # Daten alle 5 Sekunden
  enabled:
    cpu: true
    memory: true
    gpu: true                       # GPU Monitoring aktivieren
    disk: true
    temperature: true
    network: true

hardware:
  gpu:
    nvidia_only: true               # Nur NVIDIA GPUs (aktuell unterstützt)
```

### 3. Starten

```bash
# Mit Standard-Config (config.yml)
./fleet-mate

# Mit custom Config
./fleet-mate -config /path/to/config.yml

# Version anzeigen
./fleet-mate -version
```

### GPU Monitoring (NVIDIA)

Fleet Mate unterstützt NVIDIA GPU Monitoring via `nvidia-smi`. Voraussetzungen:

```bash
# NVIDIA Treiber installiert?
nvidia-smi

# Sollte GPU-Informationen anzeigen
```

Überwachte GPU-Metriken:
- ✅ GPU Auslastung (%)
- ✅ VRAM Total (MB)
- ✅ VRAM Used (MB)
- ✅ VRAM Free (MB)
- ✅ VRAM Used (%)
- ✅ GPU Temperatur (°C)

---

## 🔧 Als Service installieren (systemd)

```bash
# Service-Datei erstellen
sudo nano /etc/systemd/system/fleet-mate.service
```

```ini
[Unit]
Description=Fleet Mate - Hardware Monitoring Agent
After=network.target

[Service]
Type=simple
User=trainer
WorkingDirectory=/home/trainer/fleet-mate
ExecStart=/home/trainer/fleet-mate/fleet-mate -config /home/trainer/fleet-mate/config.yml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Service aktivieren und starten
sudo systemctl daemon-reload
sudo systemctl enable fleet-mate
sudo systemctl start fleet-mate

# Status prüfen
sudo systemctl status fleet-mate

# Logs ansehen
sudo journalctl -u fleet-mate -f
```

---

## 📊 WebSocket Protokoll

### Messages vom Mate zum Navigator:

#### 1. Registration
```json
{
  "type": "register",
  "mate_id": "ubuntu-desktop-01",
  "data": {
    "name": "Ubuntu Desktop Trainer",
    "description": "Primary development machine"
  },
  "timestamp": "2025-11-05T14:30:00Z"
}
```

#### 2. Hardware Stats
```json
{
  "type": "stats",
  "mate_id": "ubuntu-desktop-01",
  "data": {
    "timestamp": "2025-11-05T14:30:05Z",
    "cpu": {
      "usage_percent": 45.5,
      "cores": 8,
      "model": "Intel Core i7-9750H",
      "mhz": 2600.0
    },
    "memory": {
      "total": 16842752000,
      "used": 8421376000,
      "used_percent": 50.0
    },
    "gpu": [
      {
        "index": 0,
        "name": "NVIDIA GeForce RTX 3060",
        "utilization_gpu": 19.0,
        "memory_total": 12288,
        "memory_used": 2256,
        "memory_free": 9657,
        "memory_used_percent": 18.4,
        "temperature": 44.0
      }
    ],
    "disk": [
      {
        "mount_point": "/",
        "total": 500107862016,
        "used": 250053931008,
        "used_percent": 50.0
      }
    ],
    "temperature": {
      "sensors": [
        {
          "name": "coretemp",
          "temperature": 65.0
        }
      ]
    }
  },
  "timestamp": "2025-11-05T14:30:05Z"
}
```

#### 3. Heartbeat
```json
{
  "type": "heartbeat",
  "mate_id": "ubuntu-desktop-01",
  "timestamp": "2025-11-05T14:30:30Z"
}
```

#### 4. Log Data (Response)
```json
{
  "type": "log_data",
  "mate_id": "ubuntu-desktop-01",
  "data": {
    "sessionId": "session-123",
    "chunk": "Log content chunk...",
    "progress": 50.0,
    "currentLine": 500,
    "totalLines": 1000,
    "chunkNumber": 1,
    "totalChunks": 2
  },
  "timestamp": "2025-11-05T14:30:10Z"
}
```

#### 5. Command Output (Response)
```json
{
  "type": "command_output",
  "mate_id": "ubuntu-desktop-01",
  "data": {
    "sessionId": "session-456",
    "content": "Filesystem      Size  Used Avail Use% Mounted on\n..."
  },
  "timestamp": "2025-11-05T14:30:15Z"
}
```

#### 6. Command Complete (Response)
```json
{
  "type": "command_complete",
  "mate_id": "ubuntu-desktop-01",
  "data": {
    "sessionId": "session-456",
    "exitCode": 0
  },
  "timestamp": "2025-11-05T14:30:16Z"
}
```

### Commands vom Navigator zum Mate:

#### 1. Ping
```json
{
  "type": "ping",
  "timestamp": "2025-11-05T14:30:00Z"
}
```

#### 2. Collect Stats Now
```json
{
  "type": "collect_stats",
  "timestamp": "2025-11-05T14:30:00Z"
}
```

#### 3. Read Log
```json
{
  "type": "read_log",
  "payload": {
    "sessionId": "session-123",
    "path": "/var/log/syslog",
    "mode": "smart",
    "lines": 1000
  },
  "timestamp": "2025-11-05T14:30:00Z"
}
```

#### 4. Execute Command
```json
{
  "type": "execute_command",
  "payload": {
    "sessionId": "session-456",
    "command": "df",
    "args": ["-h"],
    "workingDir": "/tmp",
    "timeout": 300
  },
  "timestamp": "2025-11-05T14:30:00Z"
}
```

#### 5. Shutdown
```json
{
  "type": "shutdown",
  "timestamp": "2025-11-05T14:30:00Z"
}
```

---

## 🔒 Sicherheit

- WebSocket-Verbindung nur zu vertrauenswürdigem Navigator
- Keine sensiblen Daten in Hardware-Stats
- Mate-ID zur Authentifizierung

---

## 🐛 Troubleshooting

### Mate verbindet nicht

```bash
# Prüfen ob Fleet Navigator läuft
curl http://localhost:2025/actuator/health

# Prüfen ob WebSocket Endpoint existiert
# (wird im nächsten Schritt im Navigator implementiert)
```

### Keine Hardware-Daten

```bash
# Logs prüfen
./fleet-mate

# Config validieren
cat config.yml
```

### GPU Monitoring funktioniert nicht

```bash
# NVIDIA Treiber prüfen
nvidia-smi

# GPU Monitoring in config.yml aktivieren
monitoring:
  enabled:
    gpu: true

# Logs prüfen
./fleet-mate
```

### Permission-Fehler bei Temperatur

```bash
# Als root laufen oder User zu 'sensors' Gruppe hinzufügen
sudo usermod -aG sensors $USER
```

---

## 📝 Entwicklung

```bash
# Dependencies installieren
go mod download

# Tests ausführen (wenn vorhanden)
go test ./...

# Build
go build -o fleet-mate main.go
```

---

## 🎯 Roadmap

- [x] GPU Monitoring (NVIDIA) - ✅ v1.1.0
- [x] Log-Analyse - ✅ v1.1.0 (smart/full/errors-only Modi)
- [x] Command Execution (von Navigator gesteuert) - ✅ v1.1.0 (Whitelist-basiert)
- [ ] AMD GPU Support
- [ ] Intel GPU Support
- [ ] Prozess-Monitoring (Top 10 CPU/RAM Prozesse)
- [ ] Service-Status (systemd services)
- [ ] TLS/SSL für WebSocket
- [ ] Authentifizierung mit API Key

---

## 📝 Changelog

### v1.1.0 (2025-11-18)
- ✅ **GPU Monitoring hinzugefügt**: NVIDIA GPU/VRAM Auslastung und Temperatur
- ✅ Unterstützung für `nvidia-smi` Integration
- ✅ GPU Metriken in WebSocket Stats
- ✅ **Log-Analyse**: Lesen und Filtern von Log-Dateien (smart/full/errors-only Modi)
- ✅ **Command Execution**: Sichere Remote-Befehlsausführung mit Whitelist/Blacklist

### v1.0.0 (2025-11-05)
- ✅ Initiales Release
- ✅ CPU, RAM, Disk, Temperature, Network Monitoring
- ✅ WebSocket Integration mit Fleet Navigator

---

**Entwickelt von:** JavaFleet Systems Consulting
**Lizenz:** MIT
