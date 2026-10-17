package manager

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// ClientInfo gets the desired information to connect to the API. IsLCU determines which API we're connected to
type ClientInfo struct {
	Port     string
	Password string
	IsLCU    bool
}

// Manager maintains desired state and websocket connection to API to monitor changes
type Manager struct {
	mu           sync.Mutex
	desiredState string
	activeClient *ClientInfo
	wsConn       *websocket.Conn
}

// Creates new manager instance and returns pointer to it
func NewManager() *Manager {
	return &Manager{
		desiredState: "chat",
	}
}

// DetectActiveClient checks if LCU or Riot Client is running
func (m *Manager) DetectActiveClient() (*ClientInfo, error) {
	// Lock so no other process can touch
	m.mu.Lock()
	// defer unlock so we unlock once function returns
	defer m.mu.Unlock()

	// Check LCU first
	if info, err := scanProcess("LeagueClientUx", true); err == nil {
		m.activeClient = info
		return info, nil
	}

	// Check Riot Client second
	if info, err := scanProcess("Riot Client", false); err == nil {
		m.activeClient = info
		return info, nil
	}

	m.activeClient = nil
	return nil, fmt.Errorf("no active client found")
}

// Scans for running api and gets the port and password. Returns pointer to ClientInfo
// Got the process from this artciel, can refer to it here: https://hextechdocs.dev/getting-started-with-the-lcu-api/
func scanProcess(processName string, isLCU bool) (*ClientInfo, error) {
	// Check running processes on Mac
	out, err := exec.Command("ps", "-A").Output()
	if err != nil {
		return nil, err
	}
	// Filters running processes for the process we want
	for _, line := range strings.Split(string(out), "\n") {
		// Check if string contains our process name
		if strings.Contains(line, processName) {
			// Define the regex for the port and password
			portRegex := regexp.MustCompile(`--app-port=([0-9]+)`)
			passwordRegex := regexp.MustCompile(`--remoting-auth-token=([\w-]+)`)

			portMatches := portRegex.FindStringSubmatch(line)
			passwordMatches := passwordRegex.FindStringSubmatch(line)
			// Since FindStringSubmatch returns an array of matches with index 0 being the overall match and index 1 being first submatch we make sure that the len of this slice is > 1.
			if len(portMatches) > 1 && len(passwordMatches) > 1 {
				return &ClientInfo{
					Port:     portMatches[1],
					Password: passwordMatches[1],
					IsLCU:    isLCU,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("%s not running", processName)
}
