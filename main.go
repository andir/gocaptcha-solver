package main

import (
	"fmt"
	"os"
	image "image"
	_ "image/png"
	_ "image/jpeg"
	"log"
	"image/jpeg"
	"io/ioutil"
	"github.com/otiai10/gosseract"
)


func main() {

	const thresholdPercent = 85

	log.SetOutput(os.Stderr)

	if len(os.Args) < 2 {
		fmt.Println("Please provide an image")
		os.Exit(1)
	}

	if file, err := os.Open(os.Args[1]); err != nil {
		fmt.Println("Failed to open image")
		log.Fatal(err)
		os.Exit(1)
	} else {
		if decodedImage, format, err := image.Decode(file); err != nil {
			fmt.Println("Failed to decode image")
			log.Fatal(err)
			os.Exit(1)
		} else {
			log.Println(format)
			rgbaImage := image.NewRGBA(decodedImage.Bounds())
			bounds := rgbaImage.Bounds()
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					color := decodedImage.At(x, y)
					r, g, b, _ := color.RGBA()
					if r + g + b > (255 * 3) * thresholdPercent {
						rgbaImage.Set(x, y, image.White)
					} else {
						rgbaImage.Set(x, y, color)
					}
				}
			}

			if fh, err := ioutil.TempFile("", "gocaptcha-solver"); err != nil {
				log.Fatal(err)
				os.Exit(3)
			} else {
				defer os.Remove(fh.Name())
				jpeg.Encode(fh, rgbaImage, nil)

				out := gosseract.Must(gosseract.Params{
					Src: fh.Name(),
					Languages: "eng",
					Whitelist:"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
				})
				fmt.Println(out)

				os.Exit(0)
			}
		}
	}
}
