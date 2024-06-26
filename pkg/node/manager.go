package node

type Manager interface {
	// Device returns the name of the device type for the node. For example, "ovs"
	// or "dind" can be returned.
	Device() string

	// Setup configures the manager and performs any actions needed to prepare the
	// manager for delivering node functions. There is no guarantee that the setup
	// is run only once, so the manager should be prepared to handle multiple
	// calls. A configuration map is provided to the manager to allow for
	// additional setup options. If the setup can not be completed, an error will
	Setup(map[string]any) error

	// Info returns information about the manager and connections to services
	// required for the manager to deliver on the node operations.
	Info() (map[string]any, error)

	// Icon returns the name of the icon that should be used when visually
	// representing nodes managed by the manager. Icons should be from the Grafana
	// Icon set:
	// https://developers.grafana.com/ui/latest/index.html?path=/story/docs-overview-icon--icons-overview
	Icon() string

	// Color returns the HTML color string that should be used when visually
	// representing nodes managed by the manager.
	Color() string

	// Nodes returns a list of nodes that are managed by the manager. If the list
	// can not be determined, an error will be returned.
	Nodes(...NodeFilter) ([]Node, error)

	// Add adds a node to the manager. If the node can not be added, an error will
	Add(name string, labels map[string]string, config map[string]any) (Node, error)

	// Remove removes a node from the manager. If the node can not be removed, an
	// error will be returned.
	Remove(Node) error
}
