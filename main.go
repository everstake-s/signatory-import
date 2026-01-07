package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/", postHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handle request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	script := fmt.Sprintf(
		`spawn sudo /usr/local/bin/signatory-cli --config /etc/signatory.yaml import --vault nitro; `+
			`expect "Enter the secret key:"; `+
			`send -- "%s\r"; `+
			`expect eof`,
		string(body),
	)

	cmd := exec.Command("expect", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Request error: %s, output: %s", err.Error(), string(out))
		http.Error(w, "Error running expect command: "+err.Error()+"; Output: "+string(out), http.StatusInternalServerError)
		return
	}

	log.Println("Request handled: output: %s", string(out))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"message": "OK"}`)
}
