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

func toPixels(chars []rune) []rune {
	// Convert the sequence of UTF-8 characters from the QR code into
	// a simpler format by removing the 2nd character of each pixel.
	// Normally, 2 identical characters are needed ("██" or "  ")
	// in order to form a square shape.
	pointer := 0
	for i := 0; i < len(chars); pointer++ {
		chars[pointer] = chars[i]
		if chars[i] == 10 { // newline (1 char)
			i++
		} else { // pixel (2 chars)
			i += 2
		}
	}
	return chars[:pointer]
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
	secret_int, _ := new(big.Int).SetString(secret, 16)
	secret = secret_int.Text(62)

	// Create QR code
	qr, err := qrcode.New(secret, qrcode.Low) // 25 x 25 without border
	check(err)
	qr.DisableBorder = true
	s := qr.ToString(true)
	pixels := toPixels([]rune(s))

	// EV3 display resolution is 128 rows x 178 columns
	const ROW_NUM_BYTES int = 24
	const QR_CODE_SIZE int = 25
	const SLEEP_DURATION = 5 * time.Second

	// each pixel on the EV3 is represented by 1 bit
	// instead of using 1 byte (8 screen pixels) per QR code pixel,
	// we use a scale factor due to the limited display resolution,
	// so each QR code pixel only occupies a certain number of bits.
	const SCALE_FACTOR int = 4

	const VERTICAL_MARGIN int = 128 - (QR_CODE_SIZE * SCALE_FACTOR) // DO NOT CHANGE
	const BOTTOM_MARGIN int = VERTICAL_MARGIN/2 - 1                 // DO NOT CHANGE
	const TOP_MARGIN int = VERTICAL_MARGIN - BOTTOM_MARGIN          // DO NOT CHANGE
	// Top margin is always >= bottom margin for visibility as
	// the top of the EV3 has more shadow, reducing contrast
	const ROW_BUFFER_PADDING int = 8 - (QR_CODE_SIZE*SCALE_FACTOR)%8                                         // DO NOT CHANGE
	const HORIZONTAL_MARGIN int = (ROW_NUM_BYTES*8 - (QR_CODE_SIZE * SCALE_FACTOR) - ROW_BUFFER_PADDING) / 8 // DO NOT CHANGE
	const LEFT_MARGIN int = HORIZONTAL_MARGIN / 2                                                            // DO NOT CHANGE
	const RIGHT_MARGIN int = HORIZONTAL_MARGIN - LEFT_MARGIN                                                 // DO NOT CHANGE

	EMPTY_ROW := make([]byte, ROW_NUM_BYTES) // DO NOT CHANGE

	var row_buffer []bool
	var output []byte

	output = append(output, bytes.Repeat(EMPTY_ROW, TOP_MARGIN)...)

	var row []byte = make([]byte, LEFT_MARGIN)

	for _, pixel := range pixels { // iterate through entire QR code data
		switch pixel {
		case 9608, 32: // pixel ('█' or ' ' respectively)
			value := pixel == 9608 // true if pixel is '█'
			for i := 0; i < SCALE_FACTOR; i++ {
				row_buffer = append(row_buffer, value)
			}
		case 10: // newline, \n
			// append the remaining bits to the row buffer
			// to complete a whole byte
			row_buffer = append(row_buffer, make([]bool, ROW_BUFFER_PADDING)...)
			var byte_ byte
			for index, bit := range row_buffer {
				if bit {
					// little endian - order of bits is reversed
					byte_ |= 1 << uint(index%8)
				}
				if (index % 8) == 7 { // close byte
					row = append(row, byte_)
					byte_ = byte(0)
				}
			}

			row = append(row, make([]byte, RIGHT_MARGIN)...) // complete row

			// paste the row SCALE_FACTOR times to preserve aspect ratio
			output = append(output, bytes.Repeat(row, SCALE_FACTOR)...)

			row = make([]byte, LEFT_MARGIN) // start off a new row
			row_buffer = nil                // reset row buffer
		}
	}

	output = append(output, bytes.Repeat(EMPTY_ROW, BOTTOM_MARGIN)...)
	output = output[:len(output)-1] // removes the last newline character

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
	time.Sleep(SLEEP_DURATION)

}
