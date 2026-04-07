package state

type Newborn struct {
	Hosts    []Host
	Setup    SetupOptions
	Software SoftwareOptions
}

type Host struct {
	Connect ConnectOptions
	Setup   HostSetup
}

type ConnectOptions struct {
	IP         string
	Password   string
	SSHPort    int
	SSHKeyPath string
}

type HostSetup struct {
	Hostname     string
	Name         string
	Username     string
	Password     string
	PasswordSalt string
	SSHPort      int
	SSHKeyPath   string
	SSHPublicKey string
}

type SetupOptions struct {
	Swap         string
	ReserveFile  bool
	FirewallHTTP string
}

type SoftwareOptions struct {
	OCIRuntime string
	OCICompose bool
	K8sRuntime string
	RemoveSnap bool
}

func New() *Newborn {
	return &Newborn{
		Hosts: []Host{},
		Setup: SetupOptions{
			Swap:         "",
			ReserveFile:  false,
			FirewallHTTP: "anywhere",
		},
		Software: SoftwareOptions{
			OCIRuntime: "",
			OCICompose: false,
			K8sRuntime: "",
			RemoveSnap: false,
		},
	}
}
