package proxmox

// APIResponse represents a standard Proxmox API response
type APIResponse struct {
	Data interface{} `json:"data"`
}

// Node represents a Proxmox node
type Node struct {
	Node      string  `json:"node"`
	Status    string  `json:"status"`
	Uptime    int64   `json:"uptime"`
	CPU       float64 `json:"cpu,omitempty"`
	MaxCPU    int     `json:"maxcpu,omitempty"`
	Memory    int64   `json:"memory,omitempty"`
	MaxMemory int64   `json:"maxmemory,omitempty"`
	Disk      int64   `json:"disk,omitempty"`
	MaxDisk   int64   `json:"maxdisk,omitempty"`
}

// VM represents a virtual machine
type VM struct {
	VMID    int    `json:"vmid"`
	Name    string `json:"name"`
	Node    string `json:"node"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	CPU     int    `json:"cpus,omitempty"`
	Memory  int64  `json:"memory,omitempty"`
	MaxDisk int64  `json:"maxdisk,omitempty"`
	Uptime  int64  `json:"uptime,omitempty"`
	PID     int    `json:"pid,omitempty"`
}

// Container represents an LXC container
type Container struct {
	VMID    int    `json:"vmid"`
	Name    string `json:"name"`
	Node    string `json:"node"`
	Status  string `json:"status"`
	Type    string `json:"type"`
	CPU     int    `json:"cpus,omitempty"`
	Memory  int64  `json:"memory,omitempty"`
	MaxDisk int64  `json:"maxdisk,omitempty"`
	Uptime  int64  `json:"uptime,omitempty"`
}

// Storage represents a storage device
type Storage struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Enabled int    `json:"enabled,omitempty"`
	Used    int64  `json:"used,omitempty"`
	Total   int64  `json:"total,omitempty"`
}

// MemoryInfo represents memory statistics
type MemoryInfo struct {
	Used      int64 `json:"used,omitempty"`
	Available int64 `json:"available,omitempty"`
	Total     int64 `json:"total,omitempty"`
	Free      int64 `json:"free,omitempty"`
}

// SwapInfo represents swap statistics
type SwapInfo struct {
	Used  int64 `json:"used,omitempty"`
	Total int64 `json:"total,omitempty"`
	Free  int64 `json:"free,omitempty"`
}

// RootfsInfo represents rootfs statistics
type RootfsInfo struct {
	Used  int64 `json:"used,omitempty"`
	Total int64 `json:"total,omitempty"`
	Free  int64 `json:"free,omitempty"`
	Avail int64 `json:"avail,omitempty"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Cores   int    `json:"cores,omitempty"`
	CPUs    int    `json:"cpus,omitempty"`
	Sockets int    `json:"sockets,omitempty"`
	MHz     string `json:"mhz,omitempty"`
	Model   string `json:"model,omitempty"`
}

// NodeStatus represents detailed node status from /nodes/{node}/status API
type NodeStatus struct {
	Uptime     int64      `json:"uptime,omitempty"`
	CPU        float64    `json:"cpu,omitempty"`
	Idle       float64    `json:"idle,omitempty"`
	Wait       float64    `json:"wait,omitempty"`
	Memory     MemoryInfo `json:"memory,omitempty"`
	Swap       SwapInfo   `json:"swap,omitempty"`
	Rootfs     RootfsInfo `json:"rootfs,omitempty"`
	CPUInfo    CPUInfo    `json:"cpuinfo,omitempty"`
	PVEVersion string     `json:"pveversion,omitempty"`
	KVersion   string     `json:"kversion,omitempty"`
	LoadAvg    []string   `json:"loadavg,omitempty"`
}

// Task represents a background task
type Task struct {
	ID        string `json:"id"`
	Node      string `json:"node"`
	PID       int    `json:"pid,omitempty"`
	PPID      int    `json:"ppid,omitempty"`
	Starttime int64  `json:"starttime,omitempty"`
	Type      string `json:"type,omitempty"`
	User      string `json:"user,omitempty"`
	Status    string `json:"status,omitempty"`
}

// Cluster represents cluster information
type Cluster struct {
	Name       string `json:"name"`
	NodesCount int    `json:"nodes,omitempty"`
	Version    string `json:"version,omitempty"`
}

// User represents a Proxmox user
type User struct {
	UserID    string   `json:"userid"`
	Enable    int      `json:"enable,omitempty"`
	Expire    int64    `json:"expire,omitempty"`
	FirstName string   `json:"firstname,omitempty"`
	LastName  string   `json:"lastname,omitempty"`
	Email     string   `json:"email,omitempty"`
	Comment   string   `json:"comment,omitempty"`
	Groups    []string `json:"groups,omitempty"`
}

// Group represents a Proxmox user group
type Group struct {
	GroupID string `json:"groupid"`
	Comment string `json:"comment,omitempty"`
}

// Role represents a Proxmox role
type Role struct {
	RoleID    string   `json:"roleid"`
	Privs     []string `json:"privs,omitempty"`
	Special   bool     `json:"special,omitempty"`
	Comment   string   `json:"comment,omitempty"`
	AutoGenID bool     `json:"autogenid,omitempty"`
}

// APIToken represents an API token
type APIToken struct {
	TokenID     string `json:"tokenid"`
	Expire      int64  `json:"expire,omitempty"`
	Comment     string `json:"comment,omitempty"`
	LastUsed    int64  `json:"lastused,omitempty"`
	Value       string `json:"value,omitempty"`
	FullTokenID string `json:"full-tokenid,omitempty"`
}

// ACLEntry represents an access control list entry
type ACLEntry struct {
	Path        string `json:"path"`
	RoleID      string `json:"roleid"`
	UserID      string `json:"user,omitempty"`
	GroupID     string `json:"group,omitempty"`
	TokenID     string `json:"token,omitempty"`
	Propagate   int    `json:"propagate,omitempty"`
	Type        string `json:"type,omitempty"`
	IssueTime   int64  `json:"issuetime,omitempty"`
	ExpiresTime int64  `json:"expirestime,omitempty"`
}
