package topology

const (
	HostStateUnknown HostState = "unknown"
	HostStateReady   HostState = "ready"
	HostStateRunning HostState = "running"
	HostStateError   HostState = "error"
	HostStateStopped HostState = "stopped"
	HostStateRemoved HostState = "removed"
)

// State returns a HostState representing the status of the underlying node. To
// do this, it uses the Info and Running functions of the node. If Info can not
// be collected, it is assumed that the node is not created or has been removed.
// However, if this is not the case, the state will be considered Ready.
// Furthermore, if Info determines the node to be ready and Running returns
// true, the node is considered to be running.
func (h *Host) State() HostState {
	state := HostStateUnknown
	if _, err := h.node.Info(); err == nil {
		state = HostStateReady
	} else {
		state = HostStateRemoved
	}
	if state == HostStateReady && h.node.Running() {
		state = HostStateRunning
	}
	return state
}
