package node

import "sync"

var (
	managerLock *sync.RWMutex = &sync.RWMutex{}
	managers    []Manager
)

func RegisterManager(m Manager) {
	managerLock.Lock()
	defer managerLock.Unlock()
	managers = append(managers, m)
}

func Managers() []Manager {
	managerLock.RLock()
	defer managerLock.RUnlock()
	return managers
}

func Device(device string) Manager {
	managerLock.RLock()
	defer managerLock.RUnlock()
	for _, m := range managers {
		if m.Device() == device {
			return m
		}
	}
	return nil
}

func Devices() []string {
	managerLock.RLock()
	defer managerLock.RUnlock()
	devices := make([]string, 0, len(managers))
	for _, m := range managers {
		devices = append(devices, m.Device())
	}
	return devices
}
