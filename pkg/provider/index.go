package provider

type HostProvider interface {
	Init(map[string]interface{}) error
	// GetAllHosts() ([]string, error)
	GetActiveHosts() ([]string, error)
	GetBestHost() (string, error)
	RandomHost() (string, error)
	// GetDownHosts() ([]string, error)
}
