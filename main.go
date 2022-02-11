package main

import (
	"bytes"
	"image"
	"image/draw"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/ev3go/ev3"
	"github.com/ev3go/ev3dev/fb"
	qrcode "github.com/skip2/go-qrcode"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ev3.LCD.Init(true)
	defer ev3.LCD.Close()

	// Read secret from system
	data, err := ioutil.ReadFile("/var/lib/sling/secret")
	check(err)
	secret := string(data)
	secret = strings.TrimSpace(secret)
	secret = strings.ReplaceAll(secret, "-", "")
	secret_int := new(big.Int)
	secret_int.SetString(secret, 16)
	secret = secret_int.Text(62)

	// Create QR code
	q, err := qrcode.New(secret, qrcode.Low) // 25 x 25 without border
	check(err)
	q.DisableBorder = true
	s := q.ToString(true)
	b := []byte(s)

	// EV3 display resolution is 128 rows x 178 columns
	const ROW_NUM_BYTES int = 24
	const QR_CODE_SIZE int = 25

	// each pixel on the EV3 is represented by 1 bit
	// instead of using 1 byte (8 screen pixels) per QR code pixel,
	// we use a scale factor due to the limited display resolution,
	// so each QR code pixel only occupies a certain number of bits.
	const SCALE_FACTOR int = 5

	const VERTICAL_MARGIN int = 128 - (QR_CODE_SIZE * SCALE_FACTOR) // DO NOT CHANGE
	const BOTTOM_MARGIN int = VERTICAL_MARGIN / 3                   // DO NOT CHANGE
	const TOP_MARGIN int = VERTICAL_MARGIN - BOTTOM_MARGIN          // DO NOT CHANGE
	// Top margin is always >= bottom margin for visibility as
	// the top of the EV3 has more shadow, reducing contrast
	const ROW_BUFFER_PADDING int = 8 - (QR_CODE_SIZE*SCALE_FACTOR)%8                                         // DO NOT CHANGE
	const HORIZONTAL_MARGIN int = (ROW_NUM_BYTES*8 - (QR_CODE_SIZE * SCALE_FACTOR) - ROW_BUFFER_PADDING) / 8 // DO NOT CHANGE

	EMPTY_ROW := make([]byte, ROW_NUM_BYTES) // DO NOT CHANGE

	var row_buffer []bool
	var output []byte

	for i := 0; i < TOP_MARGIN; i++ {
		output = append(output, EMPTY_ROW...)
	}

	var row []byte

	for pointer := 0; pointer < len(b); { // iterate through entire QR code data
		if b[pointer] == 226 {
			// a "black" QR code sub-pixel (█). 2 such sub-pixels are
			// needed to form a single square QR code pixel.
			// b[i:i+6] == [226, 150, 136, 266, 150, 136]
			for j := 0; j < SCALE_FACTOR; j++ {
				row_buffer = append(row_buffer, true)
			}
			pointer += 6 // next QR code pixel
		} else if b[pointer] == 32 {
			// a "white" QR code pixel -- a simple space character ( ).
			// 2 such characters are needed to form a single square QR code pixel.
			// b[i:i+2] == [32, 32]
			for j := 0; j < SCALE_FACTOR; j++ {
				row_buffer = append(row_buffer, false)
			}
			pointer += 2 // next QR code pixel
		} else {
			// newline, end of row.
			// b[i] == 10

			// append the remaining bits to the row buffer
			// to complete a whole byte
			row_buffer = append(row_buffer, make([]bool, ROW_BUFFER_PADDING)...)
			var byte_ byte
			for j := 0; j < len(row_buffer); j++ {
				if row_buffer[j] {
					// little endian - order of bits is reversed
					byte_ |= 1 << uint(j%8)
				}
				if (j % 8) == 7 {
					row = append(row, byte_)
					byte_ = byte(0)
				}
			}

			row = append(row, make([]byte, HORIZONTAL_MARGIN)...) // complete row

			// paste the row SCALE_FACTOR times to preserve aspect ratio
			// for j := 0; j < SCALE_FACTOR; j++ {
			// 	output = append(output, row...)
			// }
			output = append(output, bytes.Repeat(row, SCALE_FACTOR)...)

			row = nil        // start off a new row
			row_buffer = nil // reset row buffer
			pointer += 1     // next QR code pixel
		}

	}
	output = output[:len(output)-1]

	var secret_image = &fb.Monochrome{
		Pix:    output,
		Stride: ROW_NUM_BYTES,
		Rect: image.Rectangle{
			Min: image.Point{
				X: 0,
				Y: 0,
			},
			Max: image.Point{
				X: 178,
				Y: 128,
			},
		},
	}

	// Render the secret image to the LCD
	draw.Draw(ev3.LCD, ev3.LCD.Bounds(), secret_image, secret_image.Bounds().Min, draw.Src)
	time.Sleep(5 * time.Second)

}
