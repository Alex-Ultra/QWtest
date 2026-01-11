package models

type Agent struct {
	ID          string  `json:"id"`
	Token       string  `json:"-"`
	IP          string  `json:"ip"`
	Platform    string  `json:"platform"` // "windows", "linux"
	LastSeen    int64   `json:"last_seen"`
	Hashrate    float64 `json:"hashrate"`
	CPUUsage    float64 `json:"cpu_usage"`
	RAMUsage    float64 `json:"ram_usage"`
	TempCPU     int     `json:"temp_cpu"`
	Status      string  `json:"status"` // "online", "mining", "stopped", "offline"
	Version     string  `json:"version"`
}

type Command struct {
	ID       string      `json:"id"`
	AgentID  string      `json:"agent_id"`
	Action   string      `json:"action"` // "start_monitor", "stop_monitor", "update_config", "reboot"
	Payload  interface{} `json:"payload,omitempty"`
	Created  int64       `json:"created"`
	Sent     bool        `json:"sent"`
}

type Metrics struct {
	Hashrate float64 `json:"hashrate"`
	CPUUsage float64 `json:"cpu_usage"`
	RAMUsage float64 `json:"ram_usage"`
	TempCPU  int     `json:"temp_cpu"`
	Status   string  `json:"status"`
}

type CommandRequest struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload,omitempty"`
}