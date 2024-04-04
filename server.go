package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type File struct {
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	WordsPointers []float64 `json:"wordspointers"`
	SliceWords    []string  `json:"slicewords"`
}

var templates = template.Must(template.ParseFiles("index.html"))

const downloadedFilesDir = "downloaded_files"
const resultFilesDir = "result_files"
const saveFileDir = "save_files"

func indexHandle(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Println("[ERROR] ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func receivedFileHandle(w http.ResponseWriter, r *http.Request) {
	responseChannel := make(chan struct{})
	decoderStopper := make(chan struct{})

	go func() {
		// ParseMultipartForm parses a request body as multipart/form-data
		r.ParseMultipartForm(32 << 20)

		file, handler, err := r.FormFile("receivedFile") // Retrieve the file from form data
		if err != nil {
			log.Println("[ERROR] ", err)
			return
		}
		defer file.Close() // Close the file when we finish

		languageFlag := r.FormValue("lang")

		// часть с созданием унимальных имен файлов

		// This is path which we want to store the file
		downloadedFile, err := os.OpenFile(path.Join(downloadedFilesDir, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println("[ERROR]  open file in", downloadedFilesDir, err)
			return
		} else {
			log.Printf("[INFO]   The file \"%v\" was received successfully\n", handler.Filename)
		}

		// Copy the file to the destination path
		io.Copy(downloadedFile, file)
		downloadedFile.Close()

		// decoder_output
		err = runDecodeScript(decoderStopper, path.Join(downloadedFilesDir, handler.Filename), resultFilesDir, languageFlag)
		if err != nil {
			log.Println("[ERROR]  decode.py", err)
			return
		}

		fileName := handler.Filename[:len(handler.Filename)-3] + "txt"
		resultFile, err := os.Open(path.Join(resultFilesDir, fileName))
		if err != nil {
			log.Println("[ERROR]  open file in", resultFilesDir, err)
			return
		} else {
			log.Printf("[INFO]   The file \"%v\" was decoded successfully\n", handler.Filename)
		}
		defer resultFile.Close()

		data, err := io.ReadAll(resultFile)
		if err != nil {
			log.Println("[ERROR]  read file in", resultFilesDir, err)
			return
		}

		decodeString := string(data)

		var wordsPointers []float64
		var sliceWords []string
		bodyString := ""
		parseDecodeFile(&wordsPointers, &sliceWords, &bodyString, decodeString)

		editFile := File{
			Title:         fileName,
			Body:          bodyString, //[:len(bodyString)-1],
			WordsPointers: wordsPointers,
			SliceWords:    sliceWords,
		}

		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(editFile)

		responseChannel <- struct{}{}
	}()

	select {
	case <-r.Context().Done():
		decoderStopper <- struct{}{}
		log.Println("[ERROR]  the process was interrupted before it was executed")
		return
	case <-responseChannel:
		log.Println("[INFO]   the process was completed successfully")
		return
	}
}

func runDecodeScript(decoderStopper <-chan struct{}, fullPathToFile string, outputPath string, languageFlag string) error {
	decoderOut := make(chan error)

	Cmd := exec.Command("python3.11", "decode.py", fullPathToFile, outputPath, languageFlag)

	go func() {
		err := Cmd.Start()
		if err != nil {
			decoderOut <- err
		}
		err = Cmd.Wait()
		if err != nil {
			decoderOut <- err
		}
		decoderOut <- nil
	}()

	select {
	case result := <-decoderOut:
		return result
	case <-decoderStopper:
		err := Cmd.Process.Kill()
		if err != nil {
			return err
		}
		return errors.New("pre-stop the decoder")
	}
}

func parseDecodeFile(wordsPointers *[]float64, sliceWords *[]string, bodyString *string, outputString string) {
	values := strings.Split(outputString, "|")

	*bodyString += values[0]
	rightValuesSplitted := strings.Split(values[1], ",")
	lenRightValuesSplitted := len(rightValuesSplitted)

	*wordsPointers = make([]float64, lenRightValuesSplitted/2)
	*sliceWords = make([]string, lenRightValuesSplitted/2)

	for i := 0; i < lenRightValuesSplitted; i += 2 {
		val, _ := strconv.ParseFloat(rightValuesSplitted[i], 32)
		(*wordsPointers)[i/2] = val
		(*sliceWords)[i/2] = rightValuesSplitted[i+1]
	}
}

func main() {
	IP := "0.0.0.0"
	PORT := "3000"

	arguments := os.Args
	if len(arguments) == 1 {
		log.Println("[INFO]   Using default port number:", PORT)
	} else {
		PORT = arguments[1]
		log.Println("[INFO]   Using port number:", PORT)
	}

	TCPIP := IP + ":" + PORT
	log.Printf("[INFO]   to go to the homepage, use \"%v\"\n", TCPIP+"/serve")
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /serve", indexHandle)
	mux.HandleFunc("POST /serve", receivedFileHandle)

	if err := http.ListenAndServe(TCPIP, mux); err != nil {
		fmt.Println(err)
	}
}
