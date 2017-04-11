package harbor

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	ConfigKey = `harbor`
)

type Config struct {
	PortRange      PortRange `json:"portsAvailable"`
	PortsAllocated []int     `json:"portsAllocated"`
}

func (m *Config) Load(db *leveldb.DB) error {
	j, err := db.Get([]byte(ConfigKey), nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(j, m); err != nil {
		return err
	}
	return err
}

func (m *Config) Save(db *leveldb.DB) error {
	j, err := m.ToJson()
	if err != nil {
		return err
	}
	if err := db.Put([]byte(ConfigKey), j, nil); err != nil {
		return err
	}
	return nil
}

func (m Config) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

func (m Config) CanAllocPorts(portRange PortRange) bool {

	if !m.PortRange.Contains(portRange) {
		return false
	}

	ports := portRange.AsInt()

	for _, port := range m.PortsAllocated {
		for i := ports[0]; i <= ports[1]; i++ {
			if port == i {
				return false
			}
		}
	}

	return true
}

func (m *Config) AllocPorts() ([]int, error) {

	availPorts := m.PortRange.AsInt()

	for from := availPorts[0]; from < availPorts[1]; from += 5 {
		to := from + 4
		if m.CanAllocPorts(PortRange(fmt.Sprintf("%d:%d", from, to))) {
			m.PortsAllocated = append(m.PortsAllocated, from, to)
			ports := make([]int, 0, 5)
			for i := from; i <= to; i++ {
				ports = append(ports, i)
			}
			return ports, nil
		}
	}
	return []int{}, errors.New("err: all ports are in use")
}

func (m *Config) ReleasePort(port int) {

	for idx, allocPort := range m.PortsAllocated {
		if allocPort != port {
			continue
		}
		tmp := m.PortsAllocated[:idx]
		m.PortsAllocated = append(tmp, m.PortsAllocated[idx+1:]...)
		break
	}
}

type PortRange string

func (m PortRange) IsValid() bool {
	ports := m.AsInt()
	return len(ports) == 2
}

func (m PortRange) AsInt() []int {
	ports := strings.Split(string(m), ":")
	if len(ports) < 2 {
		return []int{}
	}
	pFrom, err := strconv.Atoi(ports[0])
	if err != nil {
		return []int{}
	}
	pTo, err := strconv.Atoi(ports[1])
	if err != nil {
		return []int{}
	}
	return []int{pFrom, pTo}
}

func (m PortRange) Contains(portRange PortRange) bool {
	ports := portRange.AsInt()
	availPorts := m.AsInt()

	if ports[0] < availPorts[0] || availPorts[1] < ports[0] {
		return false
	}

	if ports[1] < availPorts[0] || availPorts[1] < ports[1] {
		return false
	}
	return true
}
