package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/javafleet/fleet-mate-linux/internal/commands"
	"github.com/javafleet/fleet-mate-linux/internal/config"
	"github.com/javafleet/fleet-mate-linux/internal/hardware"
)

// Client represents a WebSocket client
type Client struct {
	config       *config.Config
	conn         *websocket.Conn
	connMutex    sync.Mutex    // Protects WebSocket writes
	monitor      *hardware.Monitor
	commands     chan Command
	done         chan struct{}
	disconnected chan struct{} // Signal für Verbindungsverlust
	wakeup       chan struct{} // Signal vom UDP Discovery Listener
}

// Command represents a command from the Navigator
type Command struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Message represents a message to the Navigator
type Message struct {
	Type      string                 `json:"type"`
	MateID    string                 `json:"mate_id"`
	Data      interface{}            `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewClient creates a new WebSocket client
func NewClient(cfg *config.Config, monitor *hardware.Monitor) *Client {
	return &Client{
		config:       cfg,
		monitor:      monitor,
		commands:     make(chan Command, 10),
		done:         make(chan struct{}),
		disconnected: make(chan struct{}),
		wakeup:       make(chan struct{}, 1),
	}
}

// Connect establishes a WebSocket connection to the Navigator
func (c *Client) Connect() error {
	url := fmt.Sprintf("%s/%s", c.config.Navigator.URL, c.config.Mate.ID)
	log.Printf("Connecting to Fleet Navigator at %s", url)

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	log.Printf("Connected to Fleet Navigator")

	// Send registration message
	if err := c.sendRegistration(); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to register: %w", err)
	}

	return nil
}

// Start begins the client operations
func (c *Client) Start() error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Start reading commands from Navigator
	go c.readCommands()

	// Start sending hardware stats
	go c.sendStats()

	// Start sending heartbeats
	go c.sendHeartbeats()

	return nil
}

// Stop closes the connection and stops all operations
func (c *Client) Stop() {
	close(c.done)
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}
	log.Println("Fleet Mate stopped")
}

// sendRegistration sends registration information to Navigator
func (c *Client) sendRegistration() error {
	msg := Message{
		Type:   "register",
		MateID: c.config.Mate.ID,
		Data: map[string]interface{}{
			"name":        c.config.Mate.Name,
			"description": c.config.Mate.Description,
		},
		Timestamp: time.Now(),
	}

	return c.sendMessage(msg)
}

// sendStats periodically sends hardware statistics
func (c *Client) sendStats() {
	ticker := time.NewTicker(c.config.Monitoring.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			stats, err := c.monitor.Collect()
			if err != nil {
				log.Printf("Failed to collect stats: %v", err)
				continue
			}

			msg := Message{
				Type:   "stats",
				MateID: c.config.Mate.ID,
				Data:   stats,
				Timestamp: time.Now(),
			}

			if err := c.sendMessage(msg); err != nil {
				log.Printf("Failed to send stats: %v", err)
			}
		}
	}
}

// sendHeartbeats periodically sends heartbeat messages
func (c *Client) sendHeartbeats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			msg := Message{
				Type:   "heartbeat",
				MateID: c.config.Mate.ID,
				Timestamp: time.Now(),
			}

			if err := c.sendMessage(msg); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}

// readCommands reads commands from the Navigator
func (c *Client) readCommands() {
	errorCount := 0
	maxConsecutiveErrors := 5 // Nach 5 aufeinanderfolgenden Fehlern reconnecten

	for {
		select {
		case <-c.done:
			return
		default:
			var cmd Command
			err := c.conn.ReadJSON(&cmd)
			if err != nil {
				errorCount++

				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Println("Connection closed normally")
					return
				}

				// Bei broken pipe oder zu vielen Fehlern: Verbindung ist tot
				if websocket.IsUnexpectedCloseError(err) || errorCount >= maxConsecutiveErrors {
					log.Printf("Connection lost after %d errors, triggering reconnect: %v", errorCount, err)
					c.conn.Close()
					c.conn = nil
					// Signal Disconnection für Reconnect-Logik
					select {
					case c.disconnected <- struct{}{}:
					default:
					}
					return
				}

				log.Printf("Failed to read command: %v", err)
				time.Sleep(time.Second)
				continue
			}

			// Erfolgreicher Read → Error Counter zurücksetzen
			errorCount = 0
			log.Printf("Received command: %s", cmd.Type)
			c.handleCommand(cmd)
		}
	}
}

// handleCommand processes commands from the Navigator
func (c *Client) handleCommand(cmd Command) {
	switch cmd.Type {
	case "ping":
		c.sendPong()
	case "collect_stats":
		c.sendStatsNow()
	case "read_log":
		c.handleReadLog(cmd.Payload)
	case "execute_command":
		c.handleExecuteCommand(cmd.Payload)
	case "shutdown":
		log.Println("Shutdown command received")
		go func() {
			time.Sleep(time.Second)
			c.Stop()
		}()
	default:
		log.Printf("Unknown command type: %s", cmd.Type)
	}
}

// handleReadLog processes the read_log command
func (c *Client) handleReadLog(payload map[string]interface{}) {
	// Parse request - WICHTIG: sessionId aus Payload lesen!
	request := commands.ReadLogRequest{
		SessionID: getStringFromPayload(payload, "sessionId", ""),
		Path:      getStringFromPayload(payload, "path", "/var/log/syslog"),
		Mode:      getStringFromPayload(payload, "mode", "smart"),
		Lines:     getIntFromPayload(payload, "lines", 1000),
	}

	// Create log reader
	logReader := commands.NewLogReader(c.config.Mate.ID)

	// Execute log reading with callback to send messages
	go func() {
		err := logReader.HandleReadLogCommand(request, func(msgType string, data interface{}) {
			msg := Message{
				Type:   msgType,
				MateID: c.config.Mate.ID,
				Data:   data,
				Timestamp: time.Now(),
			}
			if err := c.sendMessage(msg); err != nil {
				log.Printf("Failed to send %s message: %v", msgType, err)
			}
		})

		if err != nil {
			log.Printf("Failed to read log file: %v", err)
		}
	}()
}

// handleExecuteCommand processes the execute_command command
func (c *Client) handleExecuteCommand(payload map[string]interface{}) {
	// Parse request
	sessionID := getStringFromPayload(payload, "sessionId", "")
	command := getStringFromPayload(payload, "command", "")
	workingDir := getStringFromPayload(payload, "workingDir", "/tmp")
	timeout := getIntFromPayload(payload, "timeout", 300)

	// Parse args array
	var args []string
	if argsInterface, ok := payload["args"]; ok {
		if argsSlice, ok := argsInterface.([]interface{}); ok {
			for _, arg := range argsSlice {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}

	request := commands.ExecuteCommandRequest{
		SessionID:  sessionID,
		Command:    command,
		Args:       args,
		WorkingDir: workingDir,
		Timeout:    timeout,
	}

	// Create command executor
	executor := commands.NewCommandExecutor(c.config.Mate.ID)

	// Execute command with callback to send messages
	go func() {
		err := executor.HandleExecuteCommand(request, func(msgType string, data interface{}) {
			msg := Message{
				Type:   msgType,
				MateID: c.config.Mate.ID,
				Data:   data,
				Timestamp: time.Now(),
			}
			if err := c.sendMessage(msg); err != nil {
				log.Printf("Failed to send %s message: %v", msgType, err)
			}
		})

		if err != nil {
			log.Printf("Failed to execute command: %v", err)
		}
	}()
}

// Helper functions to safely extract values from payload
func getStringFromPayload(payload map[string]interface{}, key string, defaultValue string) string {
	if val, ok := payload[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntFromPayload(payload map[string]interface{}, key string, defaultValue int) int {
	if val, ok := payload[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// sendPong responds to ping with pong
func (c *Client) sendPong() {
	msg := Message{
		Type:   "pong",
		MateID: c.config.Mate.ID,
		Timestamp: time.Now(),
	}
	c.sendMessage(msg)
}

// sendStatsNow immediately sends current stats
func (c *Client) sendStatsNow() {
	stats, err := c.monitor.Collect()
	if err != nil {
		log.Printf("Failed to collect stats: %v", err)
		return
	}

	msg := Message{
		Type:   "stats",
		MateID: c.config.Mate.ID,
		Data:   stats,
		Timestamp: time.Now(),
	}

	c.sendMessage(msg)
}

// sendMessage sends a message to the Navigator
func (c *Client) sendMessage(msg Message) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("Sending message: type=%s, size=%d bytes", msg.Type, len(data))

	// Lock to prevent concurrent writes to WebSocket
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// startUDPDiscoveryListener startet einen UDP-Listener für Navigator Discovery
func (c *Client) startUDPDiscoveryListener() {
	addr := net.UDPAddr{
		Port: 9090,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Failed to start UDP discovery listener: %v", err)
		return
	}
	defer conn.Close()

	log.Println("UDP Discovery Listener started on port 9090")

	buffer := make([]byte, 1024)
	for {
		select {
		case <-c.done:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				// Timeout ist OK, einfach weiter warten
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("UDP read error: %v", err)
				continue
			}

			message := strings.TrimSpace(string(buffer[:n]))
			log.Printf("Received UDP broadcast from %s: %s", remoteAddr.IP, message)

			// Prüfe ob es ein Navigator Discovery Signal ist
			if message == "FLEET_NAVIGATOR_READY" {
				log.Println("Navigator discovered! Triggering reconnect...")
				// Signal zum Reconnect senden (non-blocking)
				select {
				case c.wakeup <- struct{}{}:
				default:
				}
			}
		}
	}
}

// Run starts the client with automatic reconnection
func (c *Client) Run() error {
	attemptCount := 0
	maxAttempts := c.config.Navigator.MaxReconnectAttempts

	// Starte UDP Discovery Listener (läuft parallel)
	go c.startUDPDiscoveryListener()

	for {
		// Neue Channels für diese Verbindung erstellen
		c.done = make(chan struct{})
		c.disconnected = make(chan struct{}, 1)

		err := c.Connect()
		if err != nil {
			attemptCount++

			// Bei zu vielen Fehlversuchen: In Listener Mode gehen
			if maxAttempts > 0 && attemptCount >= maxAttempts {
				log.Printf("Max reconnect attempts reached. Entering listener mode...")
				log.Println("Waiting for Navigator discovery signal...")

				// Warte auf UDP Discovery Signal
				select {
				case <-c.wakeup:
					log.Println("Wakeup signal received, attempting reconnect...")
					attemptCount = 0 // Reset counter
					continue
				case <-c.done:
					return nil
				}
			}

			log.Printf("Connection failed (attempt %d): %v", attemptCount, err)
			log.Printf("Retrying in %s...", c.config.Navigator.ReconnectInterval)
			time.Sleep(c.config.Navigator.ReconnectInterval)
			continue
		}

		attemptCount = 0
		log.Println("Connected successfully")

		if err := c.Start(); err != nil {
			log.Printf("Failed to start client: %v", err)
			time.Sleep(c.config.Navigator.ReconnectInterval)
			continue
		}

		// Warte auf Disconnect, Wakeup oder Done Signal
		select {
		case <-c.disconnected:
			// Verbindung verloren → In Listener Mode gehen
			log.Println("Connection lost, entering listener mode...")
			log.Println("Waiting for Navigator discovery signal...")

			// Warte auf UDP Discovery Signal
			select {
			case <-c.wakeup:
				log.Println("Wakeup signal received, attempting reconnect...")
			case <-time.After(5 * time.Minute):
				log.Println("No discovery signal received for 5 minutes, trying reconnect anyway...")
			case <-c.done:
				return nil
			}

		case <-c.done:
			// Manueller Stop
			log.Println("Client stopped")
			return nil
		}
	}
}
