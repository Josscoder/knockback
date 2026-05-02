package knockback

import "sync"

var (
	managerMu sync.RWMutex
	manager   = NewManager(Settings{})
)

func setManager(m *Manager) {
	managerMu.Lock()
	defer managerMu.Unlock()
	manager = m
}

func getManager() *Manager {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager
}
