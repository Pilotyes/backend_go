package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Handler struct {
}

type UploadHandler struct {
	HostAddr  string
	UploadDir string
}

type ListHandler struct {
	FilesDir string
}

type Employee struct {
	Name   string  `json:"name"`
	Age    int     `json:"age"`
	Salary float32 `json:"salary"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		name := r.FormValue("name")
		fmt.Fprintf(w, "Parsed query-param with key \"name\": %s", name)
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to parse request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var employee Employee
		err = json.Unmarshal(body, &employee)
		if err != nil {
			http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Got a new employee!\nName: %s\nAge: %dy.o.\nSalary %0.2f\n",
			employee.Name,
			employee.Age,
			employee.Salary,
		)
	}

}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusBadRequest)
			return
		}

		filePath := h.UploadDir + "/" + header.Filename

		err = ioutil.WriteFile(filePath, data, 0777)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}

		fileLink := h.HostAddr + "/" + header.Filename

		cli := &http.Client{}

		req, err := http.NewRequest(http.MethodHead, fileLink, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to check file", http.StatusInternalServerError)
			return
		}

		resp, err := cli.Do(req)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to check file", http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != http.StatusOK {
			log.Println(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, fileLink)
	}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ext := r.FormValue("extension")

	files, err := ioutil.ReadDir(h.FilesDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if ext != "" && !strings.HasPrefix(ext, ".") {
		http.Error(w, "Bad extension", http.StatusBadRequest)
		return
	}

	for _, file := range files {
		if ext != "" {
			if fileName := file.Name(); strings.HasSuffix(fileName, ext) {
				fmt.Fprintf(w, "Filename: %s Size: %d\n", fileName, file.Size())
			}
		} else {
			fmt.Fprintf(w, "Filename: %s Size: %d\n", file.Name(), file.Size())
		}
	}
}

func main() {
	uploadHandler := &UploadHandler{
		HostAddr:  "http://localhost/",
		UploadDir: "upload",
	}

	handler := &Handler{}

	listHandler := &ListHandler{
		FilesDir: "upload",
	}

	http.Handle("/upload", uploadHandler)
	http.Handle("/", handler)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	http.Handle("/list", listHandler)

	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		srv := &http.Server{
			Addr:         ":80",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		srv.ListenAndServe()
		wg.Done()
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		dirToServe := http.Dir(uploadHandler.UploadDir)

		handler := http.FileServer(dirToServe)

		fs := &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		fs.ListenAndServe()
	}(&wg)

	wg.Wait()
}
