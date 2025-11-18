package commands

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// CommandExecutor handles remote command execution with security whitelisting
type CommandExecutor struct {
	MateID string
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(mateID string) *CommandExecutor {
	return &CommandExecutor{
		MateID: mateID,
	}
}

// ExecuteCommandRequest represents the execute_command payload
type ExecuteCommandRequest struct {
	SessionID  string   `json:"sessionId"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	WorkingDir string   `json:"workingDir"`
	Timeout    int      `json:"timeout"` // seconds
}

// CommandOutputMessage represents output chunk message
type CommandOutputMessage struct {
	SessionID string `json:"sessionId"`
	Content   string `json:"content"`
}

// CommandCompleteMessage represents completion message
type CommandCompleteMessage struct {
	SessionID string `json:"sessionId"`
	ExitCode  int    `json:"exitCode"`
}

// Whitelisted commands (base commands only, without arguments)
var allowedCommands = []string{
	// System info
	"df", "free", "uptime", "uname", "hostname", "whoami", "date",

	// File operations (read-only)
	"ls", "cat", "head", "tail", "grep", "find", "du", "pwd",

	// Process monitoring
	"ps", "top", "htop", "pgrep", "pidof",

	// System services
	"systemctl", "journalctl", "service",

	// Network
	"ping", "curl", "wget", "netstat", "ss", "ip", "ifconfig",

	// Package info (read-only)
	"dpkg", "apt", "yum", "rpm",

	// Other utilities
	"which", "whereis", "file", "stat", "wc", "sort", "uniq",
	"dmesg", "lsblk", "lsusb", "lspci", "env",
}

// Forbidden dangerous commands (explicit blacklist for extra safety)
var forbiddenCommands = []string{
	"rm", "dd", "mkfs", "fdisk", "parted",
	"chmod", "chown", "chgrp",
	"useradd", "userdel", "usermod", "passwd",
	"iptables", "ufw", "firewall-cmd",
	"shutdown", "reboot", "init", "halt", "poweroff",
}

// HandleExecuteCommand processes command execution request
func (ce *CommandExecutor) HandleExecuteCommand(request ExecuteCommandRequest, sendMessage func(msgType string, data interface{})) error {
	log.Printf("Executing command: %s %v (session: %s)", request.Command, request.Args, request.SessionID)

	// Security check
	if !ce.isCommandAllowed(request.Command) {
		errMsg := fmt.Sprintf("Command not whitelisted: %s", request.Command)
		log.Printf("Security: %s", errMsg)
		sendMessage("command_error", CommandOutputMessage{
			SessionID: request.SessionID,
			Content:   errMsg + "\n",
		})
		sendMessage("command_complete", CommandCompleteMessage{
			SessionID: request.SessionID,
			ExitCode:  127, // Command not found
		})
		return fmt.Errorf(errMsg)
	}

	// Create context with timeout
	timeout := time.Duration(request.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second // Default: 5 minutes
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Build command with args
	cmd := exec.CommandContext(ctx, request.Command, request.Args...)

	// Set working directory if specified
	if request.WorkingDir != "" {
		cmd.Dir = request.WorkingDir
	}

	// Execute command and capture output
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			sendMessage("command_error", CommandOutputMessage{
				SessionID: request.SessionID,
				Content:   fmt.Sprintf("Command timeout after %d seconds\n", request.Timeout),
			})
		} else {
			sendMessage("command_error", CommandOutputMessage{
				SessionID: request.SessionID,
				Content:   string(output),
			})
		}
	} else {
		// Send stdout output
		sendMessage("command_output", CommandOutputMessage{
			SessionID: request.SessionID,
			Content:   string(output),
		})
	}

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Send completion message
	sendMessage("command_complete", CommandCompleteMessage{
		SessionID: request.SessionID,
		ExitCode:  exitCode,
	})

	log.Printf("Command completed: session=%s, exitCode=%d", request.SessionID, exitCode)
	return nil
}

// isCommandAllowed checks if command is whitelisted
func (ce *CommandExecutor) isCommandAllowed(command string) bool {
	// Check if explicitly forbidden
	for _, forbidden := range forbiddenCommands {
		if command == forbidden {
			return false
		}
	}

	// Check if in whitelist
	for _, allowed := range allowedCommands {
		if command == allowed {
			return true
		}
	}

	// Also allow common paths
	// e.g., /usr/bin/df, /bin/ls
	if strings.HasPrefix(command, "/usr/bin/") || strings.HasPrefix(command, "/bin/") {
		baseName := strings.TrimPrefix(command, "/usr/bin/")
		baseName = strings.TrimPrefix(baseName, "/bin/")
		return ce.isCommandAllowed(baseName)
	}

	return false
}
