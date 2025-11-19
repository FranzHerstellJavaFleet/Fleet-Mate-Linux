# Fleet Mate Linux - systemd Service Installation

Diese Anleitung beschreibt die Installation von Fleet Mate Linux als systemd Service.

## Voraussetzungen

- Linux System mit systemd
- Go-Binary `fleet-mate` wurde gebaut
- `config.yml` ist konfiguriert
- Root-Rechte (sudo)

## Schnell-Installation

```bash
# 1. Service installieren
sudo ./install-service.sh

# 2. Service starten
sudo systemctl start fleet-mate-linux

# 3. Status prüfen
sudo systemctl status fleet-mate-linux
```

## Dateien

- **fleet-mate-linux.service** - systemd Service-Unit-Datei
- **install-service.sh** - Installations-Script
- **uninstall-service.sh** - Deinstallations-Script

## Service-Verwaltung

### Service starten
```bash
sudo systemctl start fleet-mate-linux
```

### Service stoppen
```bash
sudo systemctl stop fleet-mate-linux
```

### Service neu starten
```bash
sudo systemctl restart fleet-mate-linux
```

### Service-Status anzeigen
```bash
sudo systemctl status fleet-mate-linux
```

### Logs ansehen
```bash
# Live-Logs (folgen)
sudo journalctl -u fleet-mate-linux -f

# Letzte 100 Zeilen
sudo journalctl -u fleet-mate-linux -n 100

# Logs seit heute
sudo journalctl -u fleet-mate-linux --since today
```

### Service beim Boot aktivieren/deaktivieren
```bash
# Aktivieren (startet automatisch beim Boot)
sudo systemctl enable fleet-mate-linux

# Deaktivieren
sudo systemctl disable fleet-mate-linux
```

## Service-Konfiguration

Der Service wird mit folgenden Eigenschaften installiert:

- **User/Group**: trainer
- **Working Directory**: `/home/trainer/NetBeansProjects/ProjekteFMH/Fleet-Mate-Linux`
- **Config**: `config.yml` im Working Directory
- **Log Directory**: `/var/log/fleet-mate/`
- **Auto-Restart**: Ja (nach 10 Sekunden)
- **Start Limit**: 5 Versuche in 300 Sekunden

## Sicherheits-Features

- `NoNewPrivileges=true` - Verhindert Privileg-Eskalation
- `PrivateTmp=true` - Isoliertes /tmp Verzeichnis
- Logging über systemd journal

## Deinstallation

```bash
# Service komplett entfernen
sudo ./uninstall-service.sh

# Optional: Log-Verzeichnis löschen
sudo rm -rf /var/log/fleet-mate
```

## Troubleshooting

### Service startet nicht

```bash
# Detaillierte Logs anzeigen
sudo journalctl -u fleet-mate-linux -n 50 --no-pager

# Service-Konfiguration prüfen
systemctl cat fleet-mate-linux

# Binary testen
./fleet-mate -version
```

### Keine Verbindung zum Navigator

```bash
# Config prüfen
cat config.yml | grep -A 3 "navigator:"

# Navigator-URL testen
curl http://localhost:2025/actuator/health
```

### Permission-Fehler

```bash
# Log-Verzeichnis Berechtigungen prüfen
ls -la /var/log/fleet-mate/

# Berechtigungen korrigieren
sudo chown -R trainer:trainer /var/log/fleet-mate/
sudo chmod 755 /var/log/fleet-mate/
```

### Service-Status zeigt "failed"

```bash
# Komplette Fehlerausgabe
sudo systemctl status fleet-mate-linux -l

# Service zurücksetzen
sudo systemctl reset-failed fleet-mate-linux

# Neu starten
sudo systemctl start fleet-mate-linux
```

## Monitoring

### Performance überwachen
```bash
# CPU/RAM Nutzung des Services
systemctl status fleet-mate-linux

# Detaillierte Resource-Nutzung
sudo systemctl show fleet-mate-linux --property=CPUUsageNSec,MemoryCurrent
```

## Manuelle Installation (Alternative)

Falls die Scripts nicht funktionieren:

```bash
# 1. Log-Verzeichnis erstellen
sudo mkdir -p /var/log/fleet-mate
sudo chown trainer:trainer /var/log/fleet-mate

# 2. Service-Datei kopieren
sudo cp fleet-mate-linux.service /etc/systemd/system/

# 3. systemd neu laden
sudo systemctl daemon-reload

# 4. Service aktivieren
sudo systemctl enable fleet-mate-linux

# 5. Service starten
sudo systemctl start fleet-mate-linux
```

## Produktions-Tipps

### 1. Log-Rotation mit journald

Die Logs werden automatisch von systemd journald verwaltet. Konfiguration in `/etc/systemd/journald.conf`:

```ini
[Journal]
SystemMaxUse=100M
SystemKeepFree=500M
SystemMaxFileSize=10M
```

### 2. Resource Limits

Service-Datei erweitern für Resource-Limits:

```ini
[Service]
MemoryMax=500M
CPUQuota=50%
TasksMax=100
```

### 3. Monitoring mit Systemd

```bash
# Service-Start-Zeit
systemd-analyze blame | grep fleet-mate-linux

# Service-Dependencies
systemctl list-dependencies fleet-mate-linux
```

## Support

Bei Problemen:
1. Logs prüfen: `sudo journalctl -u fleet-mate-linux -f`
2. Service-Status: `sudo systemctl status fleet-mate-linux`
3. Binary testen: `./fleet-mate -version`
4. Config validieren: `cat config.yml`

---

**Version:** 1.1.0
**Datum:** 2025-11-19
