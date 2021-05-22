package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
	"strings"
)

var (
	version = "dev"
	commit  = "none"
)

func detectContentType(name string) string {
	name = strings.ToLower(name)
	if strings.HasSuffix(name, ".png") {
		return "image/png"
	} else if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(name, ".bmp") {
		return "image/bmp"
	}
	return ""
}

func main() {

	imageDir := "./images"
	if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := ioutil.ReadDir(imageDir)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error reading directory: %+v", err)
			return
		}
		for i := 0; i < 100; i++ {
			pos, err := rand.Int(rand.Reader, big.NewInt(int64(len(files))))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error generating random value: %+v", err)
				return
			}
			file := files[pos.Int64()]
			if file.IsDir() {
				continue
			}
			contentType := detectContentType(file.Name())
			if contentType == "" {
				continue
			}
			fh, err := os.Open(filepath.Join(imageDir, file.Name()))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Failed open file %v: %+v", file.Name(), err)
				return
			}
			w.WriteHeader(http.StatusOK)
			io.Copy(w, fh)
			return
		}
	})); err != nil {
		log.Printf("Error: %+v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
