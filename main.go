package main

import (
	"fmt"
	"log"
	"strings"

	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	// "time"
)

var Cookie string = ""

var GbxDir string
// var uploadUrl string = "http://127.0.0.1:8080"
var uploadUrl string = "https://trackmania.exchange/upload/replay"

func chooseDirectory(w fyne.Window, h *widget.Label) {
	open := dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
		save_dir := "NoPathYet!"
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if dir != nil {
			log.Println(dir.Path())
			save_dir = dir.Path() // here value of save_dir shall be updated!
		}
		log.Println(save_dir)
		h.SetText(save_dir)
		GbxDir = save_dir
	}, w)

	// Default location
	uri, err := storage.ListerForURI(storage.NewFileURI("./"))
	if err != nil {
		log.Println(err)
		dialog.ShowError(err, w)
		return
	}

	open.SetLocation(uri)
	open.Show()

}

func main() {

	image := canvas.NewImageFromFile("./logo.png")
	image.FillMode = canvas.ImageFillOriginal

	emptyValidator := func(s string) error {
		if len(s) < 1 {
			return fmt.Errorf("field cannot be empty")
		}
		return nil
	}
	myApp := app.New()
	myWindow := myApp.NewWindow("Entry Widget")
	myWindow.SetTitle("TMX Uploader")
	myWindow.SetFixedSize(true)
	myWindow.Resize(fyne.NewSize(600, 400))

	input := widget.NewEntry()
	input.SetPlaceHolder("Paste Cookie here")
	input.Validate()

	progress := widget.NewProgressBar()
	progress.SetValue(0)

	input.Validator = emptyValidator

	selectedDir := widget.NewLabel("No Files Chosen")
	statusText := widget.NewLabel("Status")

	uploadButton := widget.NewButton("Upload", func() {

		log.Println("Cookie was:", input.Text)
		Cookie = input.Text

		log.Println("Reading files form", GbxDir)

		entries, err := os.ReadDir(GbxDir)
		if err != nil {
			log.Fatal(err)
		}

		validFiles := []string{}

		log.Println("Found", len(entries), "total files")
		for _, e := range entries {

			if strings.HasSuffix(e.Name(), ".Replay.Gbx") && !e.IsDir() {
				log.Println(e.Name())
				validFiles = append(validFiles, filepath.Join(GbxDir, e.Name()))
			}
		}
		log.Println("Found", len(validFiles), "valid files")

		for i, e := range validFiles {
			statusText.SetText(e)
			UploadReplay(e)
			// time.Sleep(time.Millisecond * 500)
			progress.SetValue(float64(i+1) / float64(len(validFiles)))
		}
	})

	input.SetOnValidationChanged(func(err error) {
		log.Println(err)
		if err != nil {
			uploadButton.Disable()
		} else {
			uploadButton.Enable()
		}
	})

	chooseButton := widget.NewButton("Choose Directory", func() {
		chooseDirectory(myWindow, selectedDir) // Text of selectedDir updated by return value
	})

	buttons := container.NewVBox(
		image,
		input,
		selectedDir,
		statusText,
		chooseButton,
		uploadButton,
	)

	myWindow.SetContent(container.NewVBox(buttons, progress))
	myWindow.ShowAndRun()

}

func UploadReplay(gbxPath string) error {

	filename := filepath.Base(gbxPath)

	log.Println("Uploading", filename, "(", gbxPath, ")")
	newLine := "\r\n"

	part1 := []byte(`-----------------------------86198276832215236822279235129` + newLine + `Content-Disposition: form-data; name="replay_file"; filename="`)
	part2 := []byte(filename)
	part3 := []byte(`"` + newLine + `Content-Type: application/octet-stream` + newLine + newLine)
	part5 := []byte(newLine + `-----------------------------86198276832215236822279235129--` + newLine)

	b, err := os.ReadFile(gbxPath)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", uploadUrl, bytes.NewBuffer(
		slices.Concat(part1, part2, part3, b, part5),
	))

	if err != nil {
		log.Println(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Content-Type", "multipart/form-data; boundary=---------------------------86198276832215236822279235129")
	req.Header.Set("Origin", "https://trackmania.exchange")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://trackmania.exchange/upload/replay")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")
	req.Header.Set("Cookie", Cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	// log.Println("response Body:", string(body))

	return nil

}
