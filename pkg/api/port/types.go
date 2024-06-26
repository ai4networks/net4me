package port

type PortListResponse struct {
	NodeID string `json:"node_id"`
	Ports  []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
	} `json:"ports"`
}

type PortInfoResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
