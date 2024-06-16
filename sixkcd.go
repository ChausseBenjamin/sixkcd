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

var version string

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
    Version: version,
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

      encoder := sixel.NewEncoder(os.Stdout)

      switch ctx.Bool("raw") {
      case true:
        _ = encoder.Encode(img)
      default:
        fmt.Println("Title: ", comic.Title)
        _ = encoder.Encode(img)
        if comic.Alt != "" {
          fmt.Println(comic.Alt)
        }
      }


      return nil
    },
  }

  if err := app.Run(os.Args); err != nil {
    log.Fatal(err)
  }
}

func fetchComic(id string) Comic {
  var res *http.Response

  var err error

  if id == "latest" {
    res, err = http.Get(hostname+"/"+target)
  } else {
    res, err = http.Get(hostname+"/"+id+"/"+target)
  }
  defer res.Body.Close()

  if err != nil {
    log.Fatal(err)
  }

  content, err := io.ReadAll(res.Body)
  if err != nil {
    log.Fatal(err)
  }
  comic := Comic{}
  err = json.Unmarshal(content, &comic)
  if err != nil {
    log.Fatal(err)
  }

  return comic
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
