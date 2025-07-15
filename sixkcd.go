package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-sixel"
	"github.com/urfave/cli/v2"
)

const (
	hostname = "https://xkcd.com"
	target   = "info.0.json"
)

var version string

type Comic struct {
	Month      string
	Num        int
	Link       string
	Year       string
	News       string
	Safe_title string
	Transcript string
	Alt        string
	Img        string
	Title      string
	Day        string
}

func main() {
	app := &cli.App{
		Name:    "siXKCD",
		Usage:   "Sixel viewer/fetcher for XKCD Comics",
		Version: version,
		Authors: []*cli.Author{
			{
				Name:  "Benjamin Chausse",
				Email: "benjamin@chausse.xyz",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "id",
				Aliases: []string{"i"},
				Usage:   "Specify which comic `ID` to fetch.\nA value of 0 will fetch the latest comic (default: random)",
			},
			&cli.BoolFlag{
				Name:    "raw",
				Aliases: []string{"r"},
				Usage:   "Output only the sixel image without the title or the alternate caption.",
			},
			&cli.StringFlag{
				Name:    "renderer",
				Aliases: []string{"R"},
				Usage:   "Specify the renderer to use: sixel, kitty",
				Value:   "sixel",
				EnvVars: []string{"SIXKCD_RENDERER"},
			},
		},
		Action: appAction,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func appAction(ctx *cli.Context) error {
	var comic *Comic
	var fetchErr error

	comic, fetchErr = fetchComic("latest")
	if fetchErr != nil {
		return fetchErr
	}

	switch ctx.String("id") {
	case "0":
		// Latest is already fetched
		break
	case "":
		// Random
		rand := rand.Intn(comic.Num + 1)
		comic, fetchErr = fetchComic(strconv.Itoa(rand))
	default:
		comic, fetchErr = fetchComic(ctx.String("id"))
	}

	if fetchErr != nil {
		return fetchErr
	}

	img, err := fetchImg(comic.Img)
	if err != nil {
		return err
	}

	isRaw := ctx.Bool("raw")
	if !isRaw {
		fmt.Println("Title: ", comic.Title)
	}

	renderer := ctx.String("renderer")
	switch strings.ToLower(renderer) {
	case "kitty":
		if err := renderKittyImage(img); err != nil {
			return err
		}
	case "sixel":
		if err := sixel.NewEncoder(os.Stdout).Encode(img); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported renderer: %s (supported: sixel, kitty)", renderer)
	}

	if !isRaw && comic.Alt != "" {
		fmt.Println(comic.Alt)
	}

	return nil
}

func fetchComic(id string) (*Comic, error) {
	comicUrl := hostname + "/" + id + "/" + target
	if id == "latest" {
		comicUrl = hostname + "/" + target
	}

	res, err := http.Get(comicUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s status code: %d", comicUrl, res.StatusCode)
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	comic := Comic{}
	if err := json.Unmarshal(content, &comic); err != nil {
		return nil, err
	}

	return &comic, nil
}

func fetchImg(url string) (image.Image, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func renderKittyImage(img image.Image) error {
	// Convert image to PNG format
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	// Base64 encode the PNG data
	data := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Split data into chunks (kitty protocol requires 4096 bytes max per chunk)
	// and chunks must be multiples of 4 bytes except the last one
	const chunkSize = 4096
	pos := 0

	for pos < len(data) {
		// Calculate chunk size - ensure it's a multiple of 4 except for the last chunk
		remaining := len(data) - pos
		currentChunkSize := chunkSize
		if remaining < chunkSize {
			currentChunkSize = remaining
		} else {
			// Ensure chunk size is multiple of 4
			currentChunkSize = (currentChunkSize / 4) * 4
		}

		chunk := data[pos : pos+currentChunkSize]
		pos += currentChunkSize

		// Start escape sequence
		fmt.Print("\033_G")

		// Add control data for first chunk
		if pos == currentChunkSize {
			fmt.Print("a=T,f=100")
		}

		// Add more data flag for all but last chunk
		if pos < len(data) {
			if pos == currentChunkSize {
				fmt.Print(",m=1")
			} else {
				fmt.Print("m=1")
			}
		} else {
			if pos != currentChunkSize {
				fmt.Print("m=0")
			}
		}

		// Add payload separator and data
		fmt.Print(";", chunk)

		// End escape sequence
		fmt.Print("\033\\")
	}

	return nil
}
