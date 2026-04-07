package config

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

type Defaults struct {
	Name       string
	IP         string
	SSHKeyPath string
}

var (
	loadDefaultsOnce sync.Once
	defaults         Defaults
)

func LoadDefaults() Defaults {
	loadDefaultsOnce.Do(func() {
		for _, path := range []string{".env", "../.env"} {
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				key, value, ok := strings.Cut(line, "=")
				if !ok {
					continue
				}

				switch strings.TrimSpace(key) {
				case "NAME":
					defaults.Name = strings.TrimSpace(value)
				case "IP":
					defaults.IP = strings.TrimSpace(value)
				case "SSH_KEY_PATH":
					defaults.SSHKeyPath = strings.TrimSpace(value)
				}
			}

			break
		}
	})

	return defaults
}
