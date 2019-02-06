package util

import (
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	//fmt.Println(payload)
	response, _ := json.Marshal(payload)
	//fmt.Println(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithJSONFromBytes(w http.ResponseWriter, code int, dataInbytes []byte) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dataInbytes)
}
