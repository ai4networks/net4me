package control

import (
	"net/http"

	"github.com/ai4networks/net4me/pkg/api/node"
	"github.com/gorilla/mux"
)

var (
	router *mux.Router = mux.NewRouter()
)

func Serve(address string) error {
	return http.ListenAndServe(address, router)
}

func init() {
	// v0
	v0 := router.PathPrefix("/api/v0").Subrouter()

	v0.HandleFunc("/device/{dev}/nodes", node.Nodes).Methods(http.MethodGet)
	v0.HandleFunc("/device/{dev}/node/add", node.Add).Methods(http.MethodPost)
}
