package main

import (
	"fmt"
	"os"
	image "image"
	_ "image/png"
	_ "image/jpeg"
	"log"
	_ "image/jpeg"
	"io/ioutil"
	"github.com/otiai10/gosseract"
	"image/color"
	"image/png"
	"strconv"
	"github.com/llgcode/draw2d"
	"golang.org/x/image/draw"
	"github.com/llgcode/draw2d/draw2dimg"
	"strings"
	"math"
)

type SubImageRect struct {
	image.Image

	source image.Image
	bounds image.Rectangle
	offset image.Point
}

func (s *SubImageRect) ColorModel() color.Model {
	return s.source.ColorModel()
}

func (s *SubImageRect) Bounds() image.Rectangle {
	return s.bounds
}

func (s *SubImageRect) At(x, y int) color.Color {
	p := image.Point{
		X:s.offset.X + x,
		Y:s.offset.Y + y,
	}
	return s.source.At(p.X, p.Y)
}

func HasWhiteBorders(img image.Image) bool {
	const DEBUG = false
	bounds := img.Bounds()
	min := bounds.Min
	max := bounds.Max
	white := color.RGBA{255,255,255, 255}

	if DEBUG {
		log.Print("min/max X: ", min.X, max.X)
	}

	for _, x := range []int{min.X, max.X} {
		for y := 0; y < max.Y; y++ {
			c := img.At(x, y)
			if DEBUG {
				log.Print(x, y, c)
			}
			if c != white {
				if DEBUG {
					log.Println("Border at ", x, y, c, "should be:", white)
				}
				return false
			}
		}
	}

	if DEBUG {
		log.Print("min/max Y:", min.Y, max.Y)
	}

	for _, y := range []int{min.Y, max.Y} {
		for x := 0; x < max.X; x ++ {
			c := img.At(x, y)
			if DEBUG {
				log.Print(x, y, c)
			}
			if c != white {
				if DEBUG {
					log.Println("Border at ", x, y, c, "should be:", white)
				}
				return false
			}
		}
	}
	return true
}

func NewSubImageRect(img image.Image, offset image.Point, size image.Point) *SubImageRect {

	if img == nil {
		return nil
	}

	if offset.X + size.X > img.Bounds().Max.X {
		return nil
	}
	if offset.Y + size.Y > img.Bounds().Max.Y {
		return nil
	}

	i := new(SubImageRect)
	i.source = img
	i.offset = offset
	i.bounds = image.Rect(0, 0, size.X, size.Y)

	return i
}

func HasContent(img image.Image) bool {
	bounds := img.Bounds()
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			if img.At(x, y) != color.White {
				return true
			}
		}
	}
	return false
}


func splitCharacters(img image.Image) (rects []*SubImageRect) {
        const size = 34
	const DEBUG = false
	white := color.RGBA{255,255,255, 255}

	bounds := img.Bounds()

        for x := 0; x < bounds.Max.X - size; x++ {
                for y := 0; y < bounds.Max.Y - size; y++ {
			log.Println(image.Pt(x, y))
                        rect := NewSubImageRect(img, image.Pt(x, y), image.Pt(size, size))
			if rect == nil {
				fmt.Println("Failed to create subrect.")
				os.Exit(1)
			}

			if HasWhiteBorders(rect) {
				if HasContent(rect) {
					if DEBUG {
						fmt.Println("Found Characters at", x, y)
					}

					if DEBUG {
						fn := fmt.Sprintf("%s-%s.png", strconv.Itoa(x), strconv.Itoa(y))
						if fh, err := os.Create(fn); err != nil {
							log.Fatal(err)
						} else {
							fmt.Println(fn)
							png.Encode(fh, rect)
							fh.Close()
						}
					}

					// figure out how wide this character really is
					var dx int = 0
					for dx = rect.Bounds().Max.X / 2; dx < rect.Bounds().Max.X -1; dx++ {
						var s = 0
						for dy := rect.Bounds().Min.Y; dy < rect.Bounds().Max.Y; dy++ {
							if rect.At(dx, dy) != white {
								s += 1
								break
							}
						}
						if s == 0 {
							// after the first white column we assume the others are empty as well
							break

						}
					}

					var ax int = 0
					for ax = rect.Bounds().Max.X / 2; ax > rect.Bounds().Min.X; ax-- {
						var s = 0
						for dy := rect.Bounds().Min.Y; dy < rect.Bounds().Max.Y; dy++ {
							if rect.At(ax, dy) != white {
								s += 1
								break
							}
						}
						if s == 0 {
							// after the first white column we assume the others are empty as well
							break

						}
					}

					log.Println("dx", dx)
					x += dx


					rect = NewSubImageRect(rect, image.Pt(ax, 0), image.Pt(dx-ax, rect.Bounds().Max.Y))
					if rects == nil {
						rects = []*SubImageRect{}
					}

					rects = append(rects, rect)
					y = 0
					break
				} else {
					if DEBUG {
						fmt.Println("Content :(")
					}
				}
			} else {
				if DEBUG {
					fmt.Println("Border :(")
				}
			}

                }
        }
	return rects
}

func Max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func caluclateStraightnessthingy(img image.Image) (sum int64) {
	bounds := img.Bounds()
	width := bounds.Max.X

	firstLine := ((width* 100)/3)/100
	secondLine := ((width*200)/3)/100

	firstLineRange := []int{Max(int(firstLine-1), int(0)), int(firstLine), Min(int(firstLine+1), int(width))}
	secondLineRange := []int{Max(int(secondLine-1), int(0)), int(secondLine), Min(int(secondLine+1), int(width))}

	ranges := append(firstLineRange, secondLineRange...)

	sum = 0
	for _, x := range ranges {
		for y := 0; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			r,g,b,_ := c.RGBA()
			sum += int64(255-r + 255-g+ 255-b)
		}
	}

	sum *= -1
	return
}


func FindBestRotation(img image.Image) (int, image.Image) {
	const maxRotation = 27

	// Just rotate the image until we have a max count of black pixles in a semi-vertial line
	// in the 1/3 and 2/3 line
	//
	// +-----------------+
	// |    |       |    |
	// |    |       |    |
	// |    |       |    |
	// +----+-------+----+

	type result struct {
		img image.Image
		degree int
		score int64
	}


	bestResult :=  result{
		img: img,
		degree: 0,
		score: -math.MaxInt64,
	}

	for x := -maxRotation; x < maxRotation; x++ {
		i := image.NewRGBA(image.Rect(0, 0, 350, 350))
		res := result{
			degree: x,
			img: i,
		}

		if x == 0 {
			res.img = img
		} else {
			rotationMatrix := draw2d.NewRotationMatrix(float64(x))
			draw2dimg.DrawImage(img, i, rotationMatrix, draw.Src, draw2dimg.LinearFilter)
		}

		res.score = caluclateStraightnessthingy(res.img)
		log.Println(res.degree, res.score)
		if res.score >= bestResult.score {
			bestResult = res
		}
	}


	return bestResult.degree, bestResult.img
}


func main() {

	const thresholdPercent = 90

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
						rgbaImage.Set(x, y, image.Black)
					}
				}
			}


			var results []string
			for idx, character := range splitCharacters(rgbaImage) {

				_, img := FindBestRotation(character)

				//log.Print(r)
				fn := fmt.Sprintf("%s.png", strconv.Itoa(idx))

				if fh, err := os.Create(fn); err != nil {
					log.Fatal(err)
				} else {
					fmt.Println(fn)
					png.Encode(fh, img)
					fh.Close()
				}

				if fh, err := ioutil.TempFile("", "gocaptcha-solver"); err != nil {
					log.Fatal(err)
					os.Exit(3)
				} else {

					png.Encode(fh, img)

					out := gosseract.Must(gosseract.Params{
						Src: fh.Name(),
						Languages: "eng",
						Whitelist: "bcdefghkmnpqrstuvwxyz2356789",
					})
					os.Remove(fh.Name())
					results = append(results, strings.TrimSpace(out))

				}

			}

			fmt.Println(strings.Join(results, ""))

		}
	}
}
