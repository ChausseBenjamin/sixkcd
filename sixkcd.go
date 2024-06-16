package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	_ "image/jpeg"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/mattn/go-sixel"
	"github.com/urfave/cli/v2"
)

const (
  hostname = "https://xkcd.com"
  target = "info.0.json"
)

type Comic struct {
  Month string
  Num int
  Link string
  Year string
  News string
  Safe_title string
  Transcript string
  Alt string
  Img string
  Title string
  Day string
}

func main() {
  app := &cli.App{
    Name: "siXKCD",
    Usage: "Sixel viewer/fetcher for XKCD Comics",
    Authors: []*cli.Author{
      {
        Name: "Benjamin Chausse",
        Email: "benjamin@chausse.xyz",
      },
    },
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name: "id",
        Aliases: []string{"i"},
        Usage: "Specify which comic `ID` to fetch.\nA value of 0 will fetch the latest comic (default: random)",
      },
      &cli.BoolFlag{
        Name: "raw",
        Aliases: []string{"r"},
        Usage: "Output only the sixel image without the title or the alternate caption.",
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
    rand := rand.Intn(comic.Num+1)
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

  if err := sixel.NewEncoder(os.Stdout).Encode(img); err != nil {
    return err
  }

  if !isRaw && comic.Alt != "" {
    fmt.Println(comic.Alt)
  }

  return nil
}

func fetchComic(id string) (*Comic, error){
  comicUrl := hostname+"/"+id+"/"+target
  if id == "latest" {
    comicUrl = hostname+"/"+target
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
