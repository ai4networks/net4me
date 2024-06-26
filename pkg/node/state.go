package node

type NodeState string

const (
	NodeStateUnknown NodeState = "unknown"
	NodeStateReady   NodeState = "ready"
	NodeStateError   NodeState = "error"
	NodeStateRunning NodeState = "running"
)
