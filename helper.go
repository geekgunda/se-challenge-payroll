package payrollapi

import (
	"fmt"
	"log"
	"net/http"
)

func writeErrResponse(w http.ResponseWriter, statusCode int, err error, msg string) {
	if err != nil {
		log.Println("Error in handler: ", err)
	}
	w.WriteHeader(statusCode)
	if msg != "" {
		fmt.Fprint(w, msg)
	}
	return
}
