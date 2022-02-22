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

func toPixels(chars []rune) [QR_CODE_SIZE][QR_CODE_SIZE]bool {
	// Convert the sequence of UTF-8 characters from the QR code into
	// a simpler format by removing the 2nd character of each pixel.
	// Normally, 2 identical characters are needed ("██" or "  ")
	// in order to form a square shape.
	var pixels [QR_CODE_SIZE][QR_CODE_SIZE]bool
	row := 0
	col := 0
	for i := 0; i < len(chars); col = (col + 1) % QR_CODE_SIZE {
		if chars[i] == 10 { // newline (1 char)
			row++
			col-- // newline is not a pixel
			i++
		} else { // pixel (2 chars)
			pixels[row][col] = chars[i] == '█'
			i += 2
		}
	}
	return pixels
}

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
const LEFT_MARGIN_SIZE int = HORIZONTAL_MARGIN / 2                                                       // DO NOT CHANGE
const ROW_BUFFER_SIZE = QR_CODE_SIZE * SCALE_FACTOR                                                      // DO NOT CHANGE
const ROW_BODY_BYTES = ROW_NUM_BYTES - HORIZONTAL_MARGIN                                                 // DO NOT CHANGE

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

	EMPTY_ROW := make([]byte, ROW_NUM_BYTES) // DO NOT CHANGE

	var row_buffer []bool = make([]bool, ROW_BUFFER_SIZE)
	var output []byte

	output = append(output, bytes.Repeat(EMPTY_ROW, TOP_MARGIN)...)

	var row []byte = make([]byte, ROW_NUM_BYTES)

	for _, pixel_row := range pixels {
		for i := 0; i < ROW_BODY_BYTES; i++ {
			row[LEFT_MARGIN_SIZE+i] = 0 // clear row
		}
		for i := range row_buffer {
			row_buffer[i] = pixel_row[i/SCALE_FACTOR]
		}
		// var byte_ byte
		for index, bit := range row_buffer {
			if bit {
				// little endian - order of bits is reversed
				row[LEFT_MARGIN_SIZE+index/8] |= 1 << uint(index%8)
				// byte_ |= 1 << uint(index%8)
			}
		}
		// paste the row SCALE_FACTOR times to preserve aspect ratio
		output = append(output, bytes.Repeat(row, SCALE_FACTOR)...)
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
