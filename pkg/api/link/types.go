package link

type LinkCreateResponse struct {
	Ports []struct{
		ID string `json:"id"`
		Name string `json:"name"`
	} `json:"ports"`
}
