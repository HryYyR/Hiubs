package handler

import (
	"fmt"
	"log"
	"net/http"
)

func RegisterHandler(httpPort int) {
	http.HandleFunc("/commit", commitService)
	http.HandleFunc("/all", allStateService)

	fmt.Println("http server listening at ", httpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
}
