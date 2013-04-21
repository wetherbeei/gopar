/*
 * Copyright © 2012 Paul Sbarra
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/tones111/go-opencl/cl"
	"github.com/tones111/raw"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
)

var (
	angle       *float64 = flag.Float64("a", 0, "Rotation Angle, CW (Degrees)")
	inFilename  *string  = flag.String("i", "", "Input Filename")
	outFilename *string  = flag.String("o", "", "Output Filename")
	help        *bool    = flag.Bool("h", false, "Display Usage")
)

type customError struct {
	mesg string
}

func (e *customError) Error() string {
	return e.mesg
}

func init() {
	flag.Parse()
}

func fatalError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func imagePixels(img image.Image) ([]uint32, error) {
	bounds := img.Bounds()
	size := bounds.Size()
	if size.X <= 0 || size.Y <= 0 {
		return nil, &customError{fmt.Sprint("Invalid image size: ", size)}
	}

	pixels := make([]uint32, 4*size.X*size.Y)
	index := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixels[index], pixels[index+1], pixels[index+2], pixels[index+3] = img.At(x, y).RGBA()
			index += 4
		}
	}
	return pixels, nil
}

func main() {
	var err error

	if *help {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	var input io.Reader
	if len(*inFilename) == 0 {
		input = os.Stdin
	} else {
		if input, err = os.Open(*inFilename); err != nil {
			fatalError(err)
		}
	}

	var output io.Writer
	if len(*outFilename) == 0 {
		output = os.Stdout
	} else {
		if output, err = os.Create(*outFilename); err != nil {
			fatalError(err)
		}
	}

	var inImage image.Image
	var inFormat string
	if inImage, inFormat, err = image.Decode(input); err != nil {
		fatalError(err)
	}

	size := inImage.Bounds().Size()
	var pixels []uint32
	if pixels, err = imagePixels(inImage); err != nil {
		fatalError(err)
	}
	_ = pixels

	for _, platform := range cl.Platforms {
		for _, dev := range platform.Devices {
			var context *cl.Context
			var queue *cl.CommandQueue
			var sourceImage, destImage *cl.Image
			var sampler *cl.Sampler
			var program *cl.Program
			var kernel *cl.Kernel
			var outPixels []byte

			if context, err = cl.NewContextOfDevices(map[cl.ContextParameter]interface{}{cl.CONTEXT_PLATFORM: platform}, []cl.Device{dev}); err != nil {
				fatalError(err)
			}

			format := cl.ImageFormat{ChannelOrder: cl.RGBA, ChannelDataType: cl.UNSIGNED_INT32}
			if sourceImage, err = context.NewImage2D(cl.MEM_READ_ONLY, format, uint32(size.X), uint32(size.Y), 0, nil); err != nil {
				fatalError(err)
			}

			if destImage, err = context.NewImage2D(cl.MEM_WRITE_ONLY, format, uint32(size.X), uint32(size.Y), 0, nil); err != nil {
				fatalError(err)
			}

			if sampler, err = context.NewSampler(false, cl.ADDRESS_CLAMP, cl.FILTER_NEAREST); err != nil {
				fatalError(err)
			}

			if queue, err = context.NewCommandQueue(dev, cl.QUEUE_NIL); err != nil {
				fatalError(err)
			}

			if program, err = context.NewProgramFromFile("rotate.cl"); err != nil {
				fatalError(err)
			}

			if err = program.Build(nil, ""); err != nil {
				if status := program.BuildStatus(dev); status != cl.BUILD_SUCCESS {
					fatalError(&customError{fmt.Sprintf("Build Error:\n%s\n", program.Property(dev, cl.BUILD_LOG))})
				}
				fatalError(err)
			}

			if kernel, err = program.NewKernelNamed("imageRotate"); err != nil {
				fatalError(err)
			}

			if err = queue.EnqueueWriteImage(sourceImage, true, [3]cl.Size{0, 0, 0}, [3]cl.Size{cl.Size(size.X), cl.Size(size.Y), 1}, 0, 0, raw.ByteSlice(pixels)); err != nil {
				fatalError(err)
			}

			if err = kernel.SetArgs(0, []interface{}{
				sourceImage, destImage,
				float32(math.Sin(*angle * math.Pi / 180)),
				float32(math.Cos(*angle * math.Pi / 180)),
				sampler}); err != nil {
				fatalError(err)
			}

			if err = queue.EnqueueKernel(kernel, []cl.Size{0, 0, 0}, []cl.Size{cl.Size(size.X), cl.Size(size.Y), 1}, []cl.Size{1, 1, 1}); err != nil {
				fatalError(err)
			}

			if outPixels, err = queue.EnqueueReadImage(destImage, true, [3]cl.Size{0, 0, 0}, [3]cl.Size{cl.Size(size.X), cl.Size(size.Y), 1}, 0, 0); err != nil {
				fatalError(err)
			}

			outImage := image.NewNRGBA(inImage.Bounds())
			for i := 0; i < len(outImage.Pix); i++ {
				outImage.Pix[i] = uint8(outPixels[4*i])
			}

			switch inFormat {
			case "jpeg":
				if err = jpeg.Encode(output, outImage, nil); err != nil {
					fatalError(err)
				}
			case "png":
				if err = png.Encode(output, outImage); err != nil {
					fatalError(err)
				}
			default:
				fatalError(&customError{fmt.Sprintf("Unknown Format: %s", inFormat)})
			}
			return
		}
	}

}
