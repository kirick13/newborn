package state

// import (
// 	"crypto/rand"
// 	"fmt"
// 	"math/big"
// )

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
			FirewallHTTP: "nowhere",
		},
		Software: SoftwareOptions{
			OCIRuntime: "",
			OCICompose: false,
			K8sRuntime: "",
			RemoveSnap: false,
		},
	}
}

// func (n *Newborn) AddRandomHost() Host {
// 	host := Host{
// 		Connect: ConnectOptions{
// 			IP:         randomIPv4(),
// 			SSHPort:    22,
// 			Password:   randomString(24, alphaNumeric),
// 			SSHKeyPath: "",
// 		},
// 		Setup: HostSetup{
// 			Hostname:     "host-" + randomString(8, lowerAlphaNumeric),
// 			Name:         "host-" + randomString(5, lowerAlphaNumeric),
// 			Username:     randomString(7, lowerAlphaNumeric),
// 			Password:     randomString(32, alphaNumeric),
// 			PasswordSalt: randomString(16, alphaNumeric),
// 			SSHPort:      randomInt(1025, 65535),
// 			SSHKeyPath:   "",
// 			SSHPublicKey: "",
// 		},
// 	}

// 	n.Hosts = append(n.Hosts, host)
// 	return host
// }

func (n *Newborn) DeleteHost(index int) bool {
	if index < 0 || index >= len(n.Hosts) {
		return false
	}

	n.Hosts = append(n.Hosts[:index], n.Hosts[index+1:]...)
	return true
}

func (n *Newborn) UpsertHost(index int, host Host) int {
	if index >= 0 && index < len(n.Hosts) {
		n.Hosts[index] = host
		return index
	}

	n.Hosts = append(n.Hosts, host)
	return len(n.Hosts) - 1
}

// const (
// 	alphaNumeric     = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
// 	lowerAlphaNumeric = "1234567890abcdefghijklmnopqrstuvwxyz"
// )

// func randomIPv4() string {
// 	return fmt.Sprintf("192.168.%d.%d", randomInt(0, 254), randomInt(1, 254))
// }

// func randomString(length int, alphabet string) string {
// 	if length <= 0 || len(alphabet) == 0 {
// 		return ""
// 	}

// 	bytes := make([]byte, length)
// 	for i := range bytes {
// 		bytes[i] = alphabet[randomInt(0, len(alphabet)-1)]
// 	}

// 	return string(bytes)
// }

// func randomInt(minValue, maxValue int) int {
// 	if maxValue <= minValue {
// 		return minValue
// 	}

// 	size := big.NewInt(int64(maxValue - minValue + 1))
// 	value, err := rand.Int(rand.Reader, size)
// 	if err != nil {
// 		return minValue
// 	}

// 	return minValue + int(value.Int64())
// }
