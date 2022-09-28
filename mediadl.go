package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func init() {
	// Set verbose logging
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	// We need yt-dlp installed and working to continue
	out, err := exec.Command("yt-dlp", "--version").Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("using yt-dlp version %s", string(out))
	fileChannel := make(chan string, 1)

	go func() {
		for {
			select {
			case fileToDownload := <-fileChannel:
				response, err := downloadMedia(fileToDownload)
				if err != nil {
					log.Printf("error downloading media: %s\n", err.Error())
				}
				log.Printf("file %s downloaded\n", response)
			}
		}
	}()

	mux := http.NewServeMux()
	dlHandler := DLHandler{
		FileChannel: fileChannel,
	}
	staticHomePage := http.FileServer(http.Dir("./static"))
	mux.Handle("/", staticHomePage)
	mux.Handle("/dl", dlHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8988"
	}
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	log.Printf("starting server on http://%s:%s", host, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}

type DLHandler struct {
	FileChannel chan string
}

func (d DLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		if err := req.ParseForm(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error parsing form: %s", err.Error())))
		}
		fileToDownload := req.FormValue("file")
		d.FileChannel <- fileToDownload
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
	}
}

func downloadMedia(url string) ([]byte, error) {
	out, err := exec.Command("yt-dlp", url).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}
