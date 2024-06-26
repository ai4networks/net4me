package node

import (
	"encoding/json"
	"net/http"

	n "github.com/ai4networks/net4me/pkg/node"
	"github.com/gorilla/mux"
)

func Nodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dev := params["dev"]
	if dev == "" {
		http.Error(w, "device is required", http.StatusBadRequest)
		return
	}
	deviceManager := n.Device(dev)
	if deviceManager == nil {
		http.Error(w, "device manager not found", http.StatusInternalServerError)
		return
	}

	nodes, err := deviceManager.Nodes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := NodeListResponse{
		Nodes: make([]NodeGetResponse, len(nodes)),
	}
	for i, node := range nodes {
		nodeName, err := node.Name()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response.Nodes[i] = NodeGetResponse{
			ID:   node.ID(),
			Name: nodeName,
		}
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
