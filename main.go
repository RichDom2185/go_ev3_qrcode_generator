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
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/ev3go/ev3"
	"github.com/ev3go/ev3dev/fb"
	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	ev3.LCD.Init(true)
	defer ev3.LCD.Close()

	q, err := qrcode.New("Skne66[bbBj2Ss#qjW/", qrcode.Low) // 25 x 25 without border
	if err != nil {
		panic(err)
	}
	q.DisableBorder = true
	// s := q.ToString(false)
	s := q.ToString(true)
	// s := q.ToSmallString(true)
	// s := q.ToSmallString(false)
	b := []byte(s)

	var output []byte
	// output = append(output, 0x00)

	// empty row of 25 bytes
	empty_row := make([]byte, 24)

	// 128 rows = 25 * 5 + 3 blank rows
	// output = append(output, empty_row...)
	// output = append(output, empty_row...)
	// output = append(output, empty_row...)
	for i := 0; i < 3; i++ {
		output = append(output, empty_row...)
	}

	// instead of using 1 byte per square, we use 5 out of 8 bits
	// hence, we need to convert every 5 squares to 5 bytes aka 40 bits

	var bitarray []bool
	// buffer := make([]bool, 40)
	// buffer_index := 0

	var row []byte
	for i := 0; i < len(b); {
		if b[i] == 226 {
			// row = append(row, 0xFF)
			for j := 0; j < 5; j++ {
				bitarray = append(bitarray, true)
				// buffer[buffer_index] = true
				// buffer_index++
			}
			// if buffer_index == 40 {
			// 	for j := 0; j < 5; j++ {
			// 		byte_ := byte(0)
			// 		for k := 0; k < 8; k++ {
			// 			if buffer[j*8+k] {
			// 				byte_ |= 1 << uint(7-k)
			// 			}
			// 		}
			// 		row = append(row, byte_)
			// 	}
			// 	buffer_index = 0
			// }
			i += 6
		} else if b[i] == 32 {
			// row = append(row, 0x00)
			for j := 0; j < 5; j++ {
				bitarray = append(bitarray, false)
				// buffer[buffer_index] = false
				// buffer_index++
			}
			// if buffer_index == 40 {
			// 	for j := 0; j < 5; j++ {
			// 		byte_ := byte(0)
			// 		for k := 0; k < 8; k++ {
			// 			if buffer[j*8+k] {
			// 				byte_ |= 1 << uint(7-k)
			// 			}
			// 		}
			// 		row = append(row, byte_)
			// 	}
			// 	buffer_index = 0
			// }
			i += 2
		} else { // newline aka b[i] == 10
			// fmt.Println(len(output))
			// fmt.Println(len(row))
			// row = append(row, 0x00, 0x00)

			// for j := 0; j < 5; j++ {
			// 	byte_ := byte(0)
			// 	for k := 0; k < 8; k++ {
			// 		if buffer[j*8+k] {
			// 			byte_ |= 1 << uint(7-k)
			// 		}
			// 	}
			// 	row = append(row, byte_)
			// }
			// buffer_index = 0

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

			row = append(row, make([]byte, 8)...)
			// row = append(row, empty_row...)
			fmt.Println(len(row))
			// fmt.Println(len(bitarray))

			// 128 rows = 25 * 5 + 3 blank rows
			output = append(output, row...)
			output = append(output, row...)
			output = append(output, row...)
			output = append(output, row...)
			output = append(output, row...)
			row = nil
			bitarray = nil
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

	for i := 0; i < 2; i++ {
		// Render the gopher to the screen.
		draw.Draw(ev3.LCD, ev3.LCD.Bounds(), secret_image, secret_image.Bounds().Min, draw.Src)

		// Run medium motor on outA at speed 50, wait for 0.5 second and then brake.
		time.Sleep(time.Second / 2)

		// Run large motors on B+C at speed 70, wait for 2 second and then brake.
		time.Sleep(2 * time.Second)

		// Run medium motor on outA at speed -75, wait for 0.5 second and then brake.
		time.Sleep(time.Second / 2)

		// Run large motors on B at speed -50 and C at speed 50, wait for 1 second and then brake.
		time.Sleep(time.Second)
	}
}
