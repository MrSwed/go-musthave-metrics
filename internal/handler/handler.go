package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"net/http"
)

func Handler() {
	mux := http.NewServeMux()
	mux.HandleFunc(constants.UpdateRoute, UpdateMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
