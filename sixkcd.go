package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	_ "image/jpeg"
	"io"
	"log"
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

type Xkcd struct {
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
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name: "id",
        Aliases: []string{"i"},
        Usage: "Specify which comic `ID` to fetch.\nA value of 0 will fetch the latest comic (default: random)",
      },
    },
    Action: func(ctx *cli.Context) error {
      comic := fetchComic("latest")
      switch ctx.String("id") {
      case "0":
        // Latest is already fetched
        break
      case "":
        // Random
        rand := rand.Intn(comic.Num+1)
        comic = fetchComic(strconv.Itoa(rand))
      default:
        comic = fetchComic(ctx.String("id"))
      }
      var img image.Image = fetchImg(comic.Img)

      fmt.Println("Title: ", comic.Title)
      encoder := sixel.NewEncoder(os.Stdout)
      _ = encoder.Encode(img)
      if comic.Alt != "" {
        fmt.Println(comic.Alt)
      }


      return nil
    },
  }

  if err := app.Run(os.Args); err != nil {
    log.Fatal(err)
  }
}

func fetchComic(id string) Xkcd {
  var res *http.Response

  var err error

  if id == "latest" {
    res, err = http.Get(hostname+"/"+target)
  } else {
    res, err = http.Get(hostname+"/"+id+"/"+target)
  }

  if err != nil {
    log.Fatal(err)
  }

  content, err := io.ReadAll(res.Body)
  res.Body.Close()
  if err != nil {
    log.Fatal(err)
  }
  today := Xkcd{}
  err = json.Unmarshal(content, &today)
  if err != nil {
    log.Fatal(err)
  }

  return today
}

func fetchImg(url string) image.Image {
  res, err := http.Get(url)
  if err != nil {
    log.Fatal(err)
  }
  img, _, err := image.Decode(res.Body)
  if err != nil {
    log.Fatal(err)
  }
  res.Body.Close()

  return img
}
