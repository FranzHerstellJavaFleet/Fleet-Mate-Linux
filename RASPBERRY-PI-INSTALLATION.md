# Fleet Officer - Raspberry Pi Installation via SSH

**Version:** 1.0
**Datum:** 5. November 2025
**Zielplattform:** Raspberry Pi 2/3/4/5 (ARMv7/ARM64)

---

## 📋 Voraussetzungen

### Auf Ihrem Computer (wo Fleet Navigator läuft):
- ✅ Fleet Navigator läuft auf Port 2025
- ✅ SSH-Zugriff auf Raspberry Pi(s)
- ✅ Netzwerk-Verbindung zu den Pis

### Auf dem Raspberry Pi:
- ✅ Raspberry Pi OS (Bullseye oder neuer)
- ✅ SSH aktiviert
- ✅ Netzwerk-Verbindung
- ✅ Mindestens 50 MB freier Speicher

---

## 🎯 Quick Start (Copy & Paste)

### Schritt 1: Welcher Raspberry Pi?

**Prüfen Sie Ihre Pi-Version:**

```bash
# Auf dem Raspberry Pi ausführen:
uname -m

# Ausgabe interpretieren:
# aarch64  → ARM64 (Raspberry Pi 3/4/5 mit 64-bit OS) → fleet-officer-arm64
# armv7l   → ARMv7 (Raspberry Pi 2/3 mit 32-bit OS)   → fleet-officer-armv7
# armv6l   → ARMv6 (Raspberry Pi Zero/1)              → fleet-officer-armv7 (kompatibel)
```

---

## 📦 Installation Schritt-für-Schritt

### Schritt 1: Binary auf den Pi kopieren

**Vom Computer aus (wo Fleet Navigator läuft):**

```bash
# Ersetzen Sie:
# - PI_USER mit Ihrem Pi-Benutzernamen (meist: pi)
# - PI_IP mit der IP-Adresse Ihres Pis
# - Wählen Sie die richtige Binary (arm64 oder armv7)

# Für ARM64 (64-bit):
scp ~/NetBeansProjects/ProjekteFMH/Fleet-Officer-Linux/fleet-officer-arm64 \
    PI_USER@PI_IP:/home/PI_USER/fleet-officer

# ODER für ARMv7 (32-bit):
scp ~/NetBeansProjects/ProjekteFMH/Fleet-Officer-Linux/fleet-officer-armv7 \
    PI_USER@PI_IP:/home/PI_USER/fleet-officer

# Beispiel (ARM64):
scp ~/NetBeansProjects/ProjekteFMH/Fleet-Officer-Linux/fleet-officer-arm64 \
    pi@192.168.1.100:/home/pi/fleet-officer
```

**Erwartete Ausgabe:**
```
fleet-officer-arm64    100%   7.4MB   2.5MB/s   00:03
```

---

### Schritt 2: Config-Datei erstellen

**Auf dem Raspberry Pi (via SSH):**

```bash
# SSH-Verbindung zum Pi:
ssh PI_USER@PI_IP

# Beispiel:
ssh pi@192.168.1.100
```

**Config-Datei erstellen:**

```bash
# Erstelle config.yml im selben Verzeichnis wie die Binary
cat > ~/config.yml << 'EOF'
officer:
  id: "raspberry-pi-01"
  name: "Raspberry Pi Wohnzimmer"
  description: "Monitoring Pi für HomeAutomation"

navigator:
  url: "ws://NAVIGATOR_IP:2025/api/fleet-officer/ws"
  max_reconnect_attempts: 0  # 0 = infinite
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

# Ersetzen Sie NAVIGATOR_IP mit der IP Ihres Navigator-Servers!
# Beispiel: sed -i 's/NAVIGATOR_IP/192.168.1.50/g' ~/config.yml
```

**WICHTIG:** Passen Sie die Config an:

```bash
# Ersetzen Sie NAVIGATOR_IP mit Ihrer tatsächlichen IP:
nano ~/config.yml

# Ändern Sie:
# - officer.id: Eindeutiger Name für diesen Pi
# - officer.name: Beschreibender Name
# - officer.description: Was macht dieser Pi?
# - navigator.url: IP-Adresse wo Fleet Navigator läuft
```

**Beispiel-Config:**
```yaml
officer:
  id: "raspi-wohnzimmer"
  name: "Raspberry Pi 4 Wohnzimmer"
  description: "Smart Home Controller mit Zigbee Bridge"

navigator:
  url: "ws://192.168.1.50:2025/api/fleet-officer/ws"
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
```

---

### Schritt 3: Binary ausführbar machen

```bash
chmod +x ~/fleet-officer
```

---

### Schritt 4: Test-Start (Vordergrund)

**Erst mal testen, ob es funktioniert:**

```bash
cd ~
./fleet-officer
```

**Erwartete Ausgabe:**
```
Fleet Officer Linux v1.0.0 starting...
Configuration loaded from config.yml
Officer ID: raspberry-pi-01
Officer Name: Raspberry Pi Wohnzimmer
Navigator URL: ws://192.168.1.50:2025/api/fleet-officer/ws
Connecting to Fleet Navigator at ws://192.168.1.50:2025/api/fleet-officer/ws/raspberry-pi-01
Connected to Fleet Navigator
Sending message: type=register, size=...
Stats collection started (interval: 5s)
Heartbeat started (interval: 30s)
```

**Wenn es funktioniert:**
- ✅ Sie sehen "Connected to Fleet Navigator"
- ✅ Im Fleet Navigator Dashboard erscheint der neue Officer
- ✅ Hardware-Stats werden angezeigt

**Mit Ctrl+C beenden.**

---

### Schritt 5: Als systemd Service einrichten (empfohlen)

**Damit Fleet Officer automatisch beim Boot startet:**

```bash
# Service-Datei erstellen
sudo nano /etc/systemd/system/fleet-officer.service
```

**Inhalt der Service-Datei:**

```ini
[Unit]
Description=Fleet Officer - Hardware Monitoring Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi
ExecStart=/home/pi/fleet-officer
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security Settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/pi

[Install]
WantedBy=multi-user.target
```

**WICHTIG:** Passen Sie `User` und `WorkingDirectory` an, falls Ihr Benutzer nicht `pi` ist!

**Service aktivieren und starten:**

```bash
# Service neu laden
sudo systemctl daemon-reload

# Service aktivieren (startet automatisch beim Boot)
sudo systemctl enable fleet-officer.service

# Service jetzt starten
sudo systemctl start fleet-officer.service

# Status prüfen
sudo systemctl status fleet-officer.service
```

**Erwartete Ausgabe:**
```
● fleet-officer.service - Fleet Officer - Hardware Monitoring Agent
     Loaded: loaded (/etc/systemd/system/fleet-officer.service; enabled)
     Active: active (running) since Tue 2025-11-05 20:35:00 CET; 5s ago
   Main PID: 1234 (fleet-officer)
      Tasks: 8 (limit: 4915)
        CPU: 120ms
     CGroup: /system.slice/fleet-officer.service
             └─1234 /home/pi/fleet-officer

Nov 05 20:35:00 raspberry-pi systemd[1]: Started Fleet Officer.
Nov 05 20:35:01 raspberry-pi fleet-officer[1234]: Fleet Officer Linux v1.0.0 starting...
Nov 05 20:35:01 raspberry-pi fleet-officer[1234]: Connected to Fleet Navigator
```

**Service-Management:**

```bash
# Status anzeigen
sudo systemctl status fleet-officer

# Logs anzeigen (Live)
sudo journalctl -u fleet-officer -f

# Logs anzeigen (letzte 50 Zeilen)
sudo journalctl -u fleet-officer -n 50

# Service stoppen
sudo systemctl stop fleet-officer

# Service neustarten
sudo systemctl restart fleet-officer

# Service deaktivieren (nicht mehr beim Boot starten)
sudo systemctl disable fleet-officer
```

---

## 🔧 Troubleshooting

### Problem 1: "Connection refused"

**Symptom:**
```
Failed to connect: dial tcp 192.168.1.50:2025: connection refused
```

**Lösung:**

```bash
# 1. Prüfen Sie, ob Fleet Navigator läuft:
curl http://NAVIGATOR_IP:2025/actuator/health

# 2. Prüfen Sie die Firewall auf dem Navigator-Server:
sudo ufw status
sudo ufw allow 2025/tcp

# 3. Prüfen Sie die IP-Adresse in config.yml:
cat ~/config.yml | grep url
```

---

### Problem 2: "No such file or directory"

**Symptom:**
```
bash: ./fleet-officer: No such file or directory
```

**Lösung:**

```bash
# Falsche Architektur kopiert! Prüfen Sie:
uname -m

# Wenn aarch64: Nutzen Sie fleet-officer-arm64
# Wenn armv7l:  Nutzen Sie fleet-officer-armv7

# Richtige Binary kopieren (siehe Schritt 1)
```

---

### Problem 3: "Permission denied"

**Symptom:**
```
bash: ./fleet-officer: Permission denied
```

**Lösung:**

```bash
chmod +x ~/fleet-officer
```

---

### Problem 4: Service startet nicht

**Symptom:**
```
● fleet-officer.service - Fleet Officer
     Loaded: loaded
     Active: failed (Result: exit-code)
```

**Lösung:**

```bash
# Logs anschauen:
sudo journalctl -u fleet-officer -n 50

# Häufige Ursachen:
# - config.yml fehlt oder falsch formatiert
# - Binary nicht ausführbar (chmod +x)
# - Navigator nicht erreichbar (IP falsch)

# Config prüfen:
cat ~/config.yml

# Binary prüfen:
ls -la ~/fleet-officer

# Manuelle Ausführung zum Testen:
cd ~
./fleet-officer
```

---

### Problem 5: Officer erscheint nicht im Dashboard

**Symptom:** Fleet Officer läuft, aber nicht im Navigator sichtbar.

**Lösung:**

```bash
# 1. Prüfen Sie Logs auf dem Pi:
sudo journalctl -u fleet-officer -n 50

# Suchen Sie nach:
# - "Connected to Fleet Navigator" ✅
# - "Failed to connect" ❌

# 2. Prüfen Sie Navigator Backend Logs:
# (auf dem Computer wo Navigator läuft)
tail -f ~/NetBeansProjects/ProjekteFMH/Fleet-Navigator/logs/fleet-navigator.log | grep officer

# 3. Prüfen Sie officer.id in config.yml:
cat ~/config.yml | grep id

# officer.id muss eindeutig sein! Wenn zwei Pis dieselbe ID haben,
# überschreiben sie sich gegenseitig.
```

---

## 📊 Monitoring & Verwaltung

### Wichtige Befehle auf dem Raspberry Pi:

```bash
# Service Status
sudo systemctl status fleet-officer

# Live Logs (Ctrl+C zum Beenden)
sudo journalctl -u fleet-officer -f

# Letzte 100 Zeilen Logs
sudo journalctl -u fleet-officer -n 100

# Service neustarten (z.B. nach Config-Änderung)
sudo systemctl restart fleet-officer

# Config bearbeiten
nano ~/config.yml
sudo systemctl restart fleet-officer  # Danach neustarten!

# Netzwerk-Verbindung prüfen
ping NAVIGATOR_IP

# Port-Verbindung prüfen
nc -zv NAVIGATOR_IP 2025
# Oder:
telnet NAVIGATOR_IP 2025
```

---

## 🚀 Mehrere Raspberry Pis einrichten

**Für jeden weiteren Pi:**

1. **Wiederholen Sie Schritt 1-5**
2. **WICHTIG:** Ändern Sie `officer.id` in der `config.yml`:

```yaml
# Pi 1:
officer:
  id: "raspi-wohnzimmer"
  name: "Raspberry Pi Wohnzimmer"

# Pi 2:
officer:
  id: "raspi-schlafzimmer"
  name: "Raspberry Pi Schlafzimmer"

# Pi 3:
officer:
  id: "raspi-keller"
  name: "Raspberry Pi Keller"
```

**Jede `officer.id` muss eindeutig sein!**

---

## 📝 Beispiel-Installation (kompletter Workflow)

**Szenario:** Raspberry Pi 4 (64-bit) mit IP 192.168.1.100, Navigator läuft auf 192.168.1.50

```bash
# === AUF IHREM COMPUTER ===

# 1. Binary kopieren
scp ~/NetBeansProjects/ProjekteFMH/Fleet-Officer-Linux/fleet-officer-arm64 \
    pi@192.168.1.100:/home/pi/fleet-officer

# 2. SSH zum Pi
ssh pi@192.168.1.100

# === AUF DEM RASPBERRY PI ===

# 3. Config erstellen
cat > ~/config.yml << 'EOF'
officer:
  id: "raspi-wohnzimmer"
  name: "Raspberry Pi 4 Wohnzimmer"
  description: "Smart Home Controller"

navigator:
  url: "ws://192.168.1.50:2025/api/fleet-officer/ws"
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

# 4. Ausführbar machen
chmod +x ~/fleet-officer

# 5. Test (optional)
./fleet-officer
# Ctrl+C zum Beenden

# 6. Service einrichten
sudo tee /etc/systemd/system/fleet-officer.service > /dev/null << 'EOF'
[Unit]
Description=Fleet Officer - Hardware Monitoring Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi
ExecStart=/home/pi/fleet-officer
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 7. Service aktivieren und starten
sudo systemctl daemon-reload
sudo systemctl enable fleet-officer.service
sudo systemctl start fleet-officer.service

# 8. Status prüfen
sudo systemctl status fleet-officer.service

# 9. Live Logs anschauen
sudo journalctl -u fleet-officer -f
```

**Fertig!** Der Pi sollte jetzt im Fleet Navigator Dashboard erscheinen. 🎉

---

## 🔄 Updates

**Wenn Sie eine neue Version von Fleet Officer deployen möchten:**

```bash
# === AUF IHREM COMPUTER ===

# 1. Neue Binary kompilieren (falls geändert)
cd ~/NetBeansProjects/ProjekteFMH/Fleet-Officer-Linux
GOOS=linux GOARCH=arm64 go build -o fleet-officer-arm64 main.go

# 2. Auf Pi kopieren
scp fleet-officer-arm64 pi@192.168.1.100:/home/pi/fleet-officer-new

# === AUF DEM RASPBERRY PI ===

# 3. Service stoppen
sudo systemctl stop fleet-officer

# 4. Alte Binary sichern (optional)
mv ~/fleet-officer ~/fleet-officer.backup

# 5. Neue Binary aktivieren
mv ~/fleet-officer-new ~/fleet-officer
chmod +x ~/fleet-officer

# 6. Service starten
sudo systemctl start fleet-officer

# 7. Status prüfen
sudo systemctl status fleet-officer
```

---

## 📋 Checkliste für Inbetriebnahme

- [ ] Fleet Navigator läuft auf Port 2025
- [ ] Raspberry Pi OS installiert und aktuell
- [ ] SSH aktiviert auf dem Pi
- [ ] IP-Adresse des Pis bekannt
- [ ] Netzwerk-Verbindung Pi ↔ Navigator funktioniert
- [ ] Richtige ARM-Version identifiziert (arm64 oder armv7)
- [ ] Binary auf Pi kopiert (`scp`)
- [ ] `config.yml` erstellt und angepasst
  - [ ] `officer.id` eindeutig
  - [ ] `navigator.url` korrekte IP
- [ ] Binary ausführbar gemacht (`chmod +x`)
- [ ] Test-Start erfolgreich (Vordergrund)
- [ ] systemd Service erstellt
- [ ] Service aktiviert und gestartet
- [ ] Service-Status OK (`systemctl status`)
- [ ] Officer erscheint im Fleet Navigator Dashboard
- [ ] Hardware-Daten werden angezeigt
- [ ] Remote Terminal funktioniert

---

## 🎯 Nächste Schritte

Nach erfolgreicher Installation können Sie:

1. **Remote Commands ausführen:**
   - Fleet Navigator öffnen → Officer anklicken → Tab "Remote Terminal"
   - Quick Actions testen: "Disk Space", "Memory Usage", etc.

2. **Hardware überwachen:**
   - Tab "Hardware Monitor" → Echtzeit-Daten
   - CPU, RAM, Disk, Temperatur, Netzwerk

3. **Logs analysieren:**
   - Tab "AI Log-Analyse" → System-Logs mit KI analysieren

4. **Command History ansehen:**
   - Remote Terminal → Button "History"
   - Letzte 100 Commands mit Exit Codes

---

## 📞 Support

**Bei Problemen:**

1. **Logs prüfen:**
   ```bash
   sudo journalctl -u fleet-officer -n 100
   ```

2. **Netzwerk testen:**
   ```bash
   ping NAVIGATOR_IP
   nc -zv NAVIGATOR_IP 2025
   ```

3. **Config validieren:**
   ```bash
   cat ~/config.yml
   ```

4. **Manuelle Ausführung:**
   ```bash
   cd ~
   ./fleet-officer
   ```

---

## 📚 Weiterführende Dokumentation

- **Fleet Officer Integration:** `FLEET-OFFICER-INTEGRATION.md`
- **Remote Command Execution:** `COMMAND-EXECUTION-IMPLEMENTATION.md`
- **Main README:** `README.md`

---

**Erstellt von:** JavaFleet Systems Consulting
**Version:** 1.0
**Datum:** 5. November 2025
**Für:** Raspberry Pi 2/3/4/5 (ARMv7/ARM64)

---

**Viel Erfolg bei der Installation! 🚢🍓**
