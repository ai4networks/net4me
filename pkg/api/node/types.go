package node

type NodeGetResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NodeListResponse struct {
	Nodes []NodeGetResponse `json:"nodes"`
}

type NodeAddRequest struct {
	Device string         `json:"device"`
	Name   string         `json:"name"`
	Config map[string]any `json:"config"`
}

type NodeAddResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type NodeInfoResponse struct {
}
