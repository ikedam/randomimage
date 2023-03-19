package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	version = "dev"
	commit  = "none"
)

func checkModifiedSince(r *http.Request, modTime time.Time) bool {
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince == "" {
		return true
	}
	checkTime, err := http.ParseTime(ifModifiedSince)
	if err != nil {
		return true
	}
	return modTime.After(checkTime)
}

func isImage(name string) bool {
	name = strings.ToLower(name)
	return strings.HasSuffix(name, ".png") ||
		strings.HasSuffix(name, ".jpg") ||
		strings.HasSuffix(name, ".jpeg") ||
		strings.HasSuffix(name, ".bmp")
}

func main() {
	imageDir := "./images"
	if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stat, err := os.Stat(imageDir)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error reading directory: %+v", err)
			return
		}
		modTime := stat.ModTime()
		w.Header().Add("Last-Modified", modTime.Format(http.TimeFormat))
		if !checkModifiedSince(r, modTime) {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		files, err := ioutil.ReadDir(imageDir)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error reading directory: %+v", err)
			return
		}
		pathList := make([]string, 0, len(files))
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if !isImage(file.Name()) {
				continue
			}
			path, err := url.JoinPath(imageDir, file.Name())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error building path: %+v", err)
				return
			}
			if strings.HasPrefix(os.Getenv("PATH_INFO"), "/") {
				// CGI is called as `.../json.cgi/images.json`
				// Path should be fixed to `../images/..`.
				// url.JoinPath() removes prefix `..` specifications.
				path = "../" + path
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Error building path: %+v", err)
					return
				}
			}
			pathList = append(pathList, path)
		}
		body, err := json.Marshal(pathList)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error marshaling json: %+v", err)
			return
		}
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.Header().Add("Content-Length", fmt.Sprintf("%v", len(body)))
		w.Header().Add("Cache-Control", "max-age=86400")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})); err != nil {
		log.Printf("Error: %+v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
