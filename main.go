package main

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"os"
	"strings"
	"time"

	"github.com/ev3go/ev3"
	"github.com/ev3go/ev3dev/fb"
	qrcode "github.com/skip2/go-qrcode"
)

const (
	// Input file containing the string to generate a QR code from
	inputFilePath = "/var/lib/sling/secret_b62"
	// Output file for PNG image of QR code
	outputImagePath = "/var/lib/sling/secret_image.png"

	// Allows for adjustments to cater to varying display resolutions
	// and QR code sizes. In the EV3, each pixel on the display is
	// represented by 1 bit. Hence, displayScaleFactor corresponds to
	// how much each QR code pixel is scaled to match the display pixels.
	//
	// For example: a displayScaleFactor of 5 means that each small black/white
	// square in the QR code measures 5x5 pixels in the actual display.
	displayScaleFactor = 4
	// The corresponding scale factor for the output PNG image
	imageScaleFactor = 8

	// How long to display the QR code before the program exits
	sleepDuration = 10 * time.Second

	// DO NOT CHANGE THE FOLLOWING VALUES

	// EV3 display resolution is 128 rows x 178 columns
	ROW_NUM_BYTES = 24
	QR_CODE_SIZE  = 25

	VERTICAL_MARGIN = 128 - (QR_CODE_SIZE * displayScaleFactor)
	BOTTOM_MARGIN   = VERTICAL_MARGIN/2 - 1
	// Top margin is always >= bottom margin for visibility as
	// the top of the EV3 has more shadow, reducing contrast
	TOP_MARGIN = VERTICAL_MARGIN - BOTTOM_MARGIN

	ROW_BUFFER_PADDING = 8 - (QR_CODE_SIZE*displayScaleFactor)%8
	HORIZONTAL_MARGIN  = (ROW_NUM_BYTES*8 - (QR_CODE_SIZE * displayScaleFactor) - ROW_BUFFER_PADDING) / 8
	LEFT_MARGIN        = HORIZONTAL_MARGIN / 2
	RIGHT_MARGIN       = HORIZONTAL_MARGIN - LEFT_MARGIN
)

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
	data, err := os.ReadFile(inputFilePath)
	if err != nil {
		// No printing of error message as logs won't be seen through EV3 GUI
		panic(err)
	}
	secret := strings.TrimSpace(string(data))

	// Create QR code
	qr, err := qrcode.New(secret, qrcode.Low) // 25 x 25 without border for a 22-character input
	if err != nil {
		// No printing of error message as logs won't be seen through EV3 GUI
		panic(err)
	}

	if _, err := os.Stat(outputImagePath); errors.Is(err, os.ErrNotExist) {
		// Image does not exist, save QR code to disk
		// Ignore any errors
		qr.WriteFile(-imageScaleFactor, outputImagePath)
	}

	qr.DisableBorder = true

	s := qr.ToString(true)
	pixels := toPixels([]rune(s))

	EMPTY_ROW := make([]byte, ROW_NUM_BYTES) // DO NOT CHANGE

	var row_buffer []bool
	var output []byte

	output = append(output, bytes.Repeat(EMPTY_ROW, TOP_MARGIN)...)

	var row []byte = make([]byte, LEFT_MARGIN)

	for _, pixel := range pixels { // iterate through entire QR code data
		switch pixel {
		case 9608, 32: // pixel ('█' or ' ' respectively)
			value := pixel == 9608 // true if pixel is '█'
			for i := 0; i < displayScaleFactor; i++ {
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
			output = append(output, bytes.Repeat(row, displayScaleFactor)...)

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
	time.Sleep(sleepDuration)

}
