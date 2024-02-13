package dims

import "net/http"

func HandleDimsStatus(config Config, debug bool, dev bool, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ALIVE"))
}
