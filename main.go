// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The Go gopher was designed by Renee French and is
// licensed under the Creative Commons Attributions 3.0.

// demo is a reimplementation of the Demo program loaded on new ev3 bricks,
// without sound. It demonstrates the use of the ev3dev Go API.
// The control does not make full use of the ev3dev API where it could.
package main

import (
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

	var output []byte

	// empty row of 24 bytes
	empty_row := make([]byte, 24)

	// 128 rows = 25 * 5 + 3 blank rows
	for i := 0; i < 3; i++ {
		output = append(output, empty_row...)
	}

	// instead of using 1 byte per square, we use 5 out of 8 bits
	// hence, we need to convert every 5 squares to 5 bytes aka 40 bits

	var bitarray []bool

	var row []byte
	for i := 0; i < len(b); {
		if b[i] == 226 {
			for j := 0; j < 5; j++ {
				bitarray = append(bitarray, true)
			}
			i += 6
		} else if b[i] == 32 {
			for j := 0; j < 5; j++ {
				bitarray = append(bitarray, false)
			}
			i += 2
		} else { // newline aka b[i] == 10
			// append the remaining 3 bits after 25 * 5 = 125 bits to get 16 whole bytes
			bitarray = append(bitarray, false, false, false)
			var byte_ byte
			for j := 0; j < 16*8; j++ {
				if bitarray[j] {
					byte_ |= 1 << uint(j%8)
				}
				if (j % 8) == 7 {
					row = append(row, byte_)
					byte_ = byte(0)
				}
			}

			row = append(row, make([]byte, 8)...) // complete row with 8 bytes

			// 128 rows = 25 * 5 + 3 blank rows
			output = append(output, row...)
			output = append(output, row...)
			output = append(output, row...)
			output = append(output, row...)
			output = append(output, row...)

			row = nil      // start off a new row
			bitarray = nil // reset bitarray
			i += 1
		}

	}
	output = output[:len(output)-1]

	var secret_image = &fb.Monochrome{
		Pix:    output,
		Stride: 24,
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
