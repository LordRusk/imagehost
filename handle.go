// for handling image requests
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	// for image support
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type imgRecord struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Format   string    `json:"format,omitempty"`
	Path     string    `json:"path,omitempty"`
	Added    time.Time `json:"added,omitempty"`
	Saved    bool      `json:"saved,omitempty"`
	contents []byte    // stored here temporarally
}

// saves the image
func (i *imgRecord) save() error {
	if i.Saved {
		return errors.New("Image has already been saved!")
	}

	if err := os.MkdirAll(*imgDir, os.ModePerm); err != nil {
		return errors.Wrap(err, "Failed to save image")
	}

	file, err := os.Create(i.Path)
	if err != nil {
		return errors.Wrap(err, "Failed to save image")
	}
	defer file.Close()

	bitesWritten, err := file.Write(i.contents)
	l.Printf("Wrote %d bytes to disk as '%s'", bitesWritten, i.Name)
	if err != nil {
		return errors.Wrap(err, "Failed to save image")
	}

	if err := file.Sync(); err != nil {
		return errors.Wrap(err, "Failed to sync file")
	}

	i.Saved = true
	return nil
}

// returns the image config
//
// because image.Config contains an
// interface, it cannot be stored in
// a json file. This is to get around that.
func (i *imgRecord) getImageConf() (image.Config, error) {
	bites, err := os.ReadFile(i.Path)
	if err != nil {
		return image.Config{}, errors.Wrap(err, "Unable get image config")
	}

	conf, _, err := image.DecodeConfig(bytes.NewReader(bites))
	if err != nil {
		return image.Config{}, errors.Wrap(err, "Unable get image config")
	}
	return conf, nil
}

type logFile struct {
	path    string
	records map[string]imgRecord
	// *********ID*****Record
}

// returns a new log file
func newLogFile(path string) *logFile {
	return &logFile{path: path, records: make(map[string]imgRecord)}
}

// adds a record to memory
func (l *logFile) add(r imgRecord) error {
	if _, ok := l.records[r.ID]; ok {
		return errors.New(fmt.Sprintf("'%s' already exists", r.Name))
	}

	l.records[r.ID] = r
	return nil
}

// saves the records to the disk
func (l *logFile) save() error {
	jsonBites, err := json.Marshal(l.records)
	if err != nil {
		return errors.Wrap(err, "Could not save records")
	}

	if err := os.WriteFile(l.path, jsonBites, 0666); err != nil {
		return errors.Wrap(err, "Could not save records")
	}
	return nil
}

// gets the records from the disk
//
// if no records exit, it will return
// an error
func (l *logFile) get() error {
	bites, err := os.ReadFile(l.path)
	if err != nil {
		return errors.Wrap(err, "Could not get records")
	}

	if err := json.Unmarshal(bites, &l.records); err != nil {
		return errors.Wrap(err, "Could not get records")
	}
	return nil
}

func getUploadedImage(w http.ResponseWriter, r *http.Request) {
	// handle the POST request
	//
	// don't reinvent the wheel
	if r.ContentLength > 0 && r.ContentLength < 1<<29 {
		r.ParseMultipartForm(r.ContentLength)
	} else {
		r.ParseMultipartForm(1 << 28)
	}

	// check if there is an uploaded image
	if r.MultipartForm == nil {
		w.Write([]byte("No file uploaded!"))
		return
	} else if len(r.MultipartForm.File["filename"]) < 1 {
		w.Write([]byte("No file uploaded!"))
		return
	}

	// get the uploaded image
	imgReader, err := r.MultipartForm.File["filename"][0].Open()
	if err != nil {
		l.Print(err)
	}

	imgBites, err := io.ReadAll(imgReader)
	if err != nil {
		l.Print(err)
	}

	// SHA256 Checksum servres as unique identiier for this file
	shaSum := sha256.Sum256(imgBites)
	_, format, err := image.DecodeConfig(bytes.NewReader(imgBites))
	if err != nil {
		l.Print(err)
	}
	img := imgRecord{
		ID:       base64.URLEncoding.EncodeToString(shaSum[:]),
		Added:    time.Now(),
		Name:     r.MultipartForm.File["filename"][0].Filename,
		Path:     *imgDir + "/" + r.MultipartForm.File["filename"][0].Filename,
		Format:   format,
		contents: imgBites,
	}

	// handle image
	if err := lf.add(img); err != nil {
		// redirect to image if it has
		// already been uploaded
		u := r.URL
		u.Path = "/img/" + img.ID
		http.Redirect(w, r, u.String(), 303)
		return
	}
	if err := img.save(); err != nil {
		l.Print(err)
	}
	if err := lf.save(); err != nil {
		l.Print(err)
	}

	// redirect to image after uploading
	u := r.URL
	u.Path = "/img/" + img.ID
	http.Redirect(w, r, u.String(), 303)
}

func handleImages(w http.ResponseWriter, r *http.Request) {
	strs := strings.Split(r.URL.Path, "/")
	record, ok := lf.records[strs[len(strs)-1]]
	if !ok { // check if the image exists
		w.Write([]byte("ID does not match any stored image"))
		return
	}

	bites, err := os.ReadFile(record.Path)
	if err != nil {
		l.Printf("Failed to read file %s: %s\n", record.Name, err)
		w.Write([]byte("Failed to read file, check console."))
		return
	}

	w.Header().Set("Content-Type", "image/"+record.Format+"; charset=utf-8")
	if _, err := w.Write(bites); err != nil {
		l.Printf("Failed to return image to user: %s\n", err)
	}
}
