package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

// opts
var index = flag.String("l", "index.html", "Set the landing page")
var port = flag.String("p", ":8080", "Set the port")
var imgDir = flag.String("imgDir", "images", "Set the images directory")

var imgLogFile = flag.String("log", "log.json", "Set the log file")
var lf *logFile

var l = log.New(os.Stdout, "", 0)

// Loads a file as a string
//
// do not use on large files
func loadFileAsString(filename string) (string, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return "", errors.Wrap(err, "Could not load file")
	}
	defer file.Close()

	bites, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Wrap(err, "Could not read file")
	}

	return string(bites), nil
}

func serveLandingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	str, err := loadFileAsString(*index)
	if err != nil {
		l.Print(errors.Wrap(err, "Failed to serve landing page"))
	}
	if _, err := io.WriteString(w, str); err != nil {
		l.Print(errors.Wrap(err, "Failed to serve landing page"))
	}
}

func main() {
	flag.Parse()
	lf = newLogFile(*imgLogFile)
	if err := lf.get(); err != nil {
		l.Print(err)
	}

	http.HandleFunc("/", serveLandingPage)
	http.HandleFunc("/upload", getUploadedImage)
	http.HandleFunc("/img/", handleImages)
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatalln(err)
	}
}
