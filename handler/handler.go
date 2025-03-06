package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func RegisterHandler(httpPort int) {
	http.HandleFunc("/commit", commitService)
	http.HandleFunc("/all", allStateService)

	fmt.Println("http server listening at ", httpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
}

func CheckCommandValid(command string) (string, bool) {
	result := strings.ReplaceAll(command, "  ", " ")
	// 连续替换多余的空格
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	result = strings.ToLower(result)

	return result, true
}
