package handler

import (
	"github.com/MrSwed/go-musthave-metrics/internal/constants"
	"github.com/MrSwed/go-musthave-metrics/internal/storage"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" ||
		r.Header.Get("Content-Length") != "0" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	params := strings.Split(
		strings.Trim(
			strings.TrimPrefix(r.URL.Path, constants.UpdateRoute),
			"/"),
		"/")

	log.Println(params)

	if len(params) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch params[0] {
	case constants.MetricTypeGauge:
		if v, err := strconv.ParseFloat(params[2], 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			storage.Store.Gauge(params[1], v)
		}
	case constants.MetricTypeCounter:
		if v, err := strconv.ParseInt(params[2], 10, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			storage.Store.Counter(params[1], v)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// test
	//ct, cl := r.Header.Get("Content-Type"), r.Header.Get("Content-Length")
	//fmt.Println(ct, cl)

	// if ok
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//w.Header().Set("Content-Length", "11")

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Saved: Ok")); err != nil {
		log.Println("error write response", err)
	}

	log.Println(storage.Store)
}

func Handler() {
	mux := http.NewServeMux()
	mux.HandleFunc(constants.UpdateRoute, UpdateMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}

}
