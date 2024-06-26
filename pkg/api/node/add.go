package node

import (
	"encoding/json"
	"net/http"

	n "github.com/ai4networks/net4me/pkg/node"
	"github.com/gorilla/mux"
)

func Add(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	device := params["dev"]
	if device == "" {
		http.Error(w, "device is required", http.StatusBadRequest)
		return
	}
	deviceManager := n.Device(device)
	if deviceManager == nil {
		http.Error(w, "device manager not found", http.StatusInternalServerError)
		return
	}

	var requestBody NodeAddRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	node, err := deviceManager.Add(requestBody.Name, make(map[string]string), requestBody.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nodeName, err := node.Name()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := NodeAddResponse{
		ID:   node.ID(),
		Name: nodeName,
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
