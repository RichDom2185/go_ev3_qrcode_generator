# QR Code Generator for ev3dev

This program uses the [ev3dev Go library](https://github.com/ev3go/ev3dev) to generate and display QR codes on the Lego EV3 screen given an input string.

Used in conjunction with [ev3-source](https://github.com/source-academy/ev3-source) to print the device secret located at `/var/lib/sling/secret`.

The decision to hardcode the file path is done as the binary will be run directly from the EV3 GUI, hence it would not be feasible to require any command-line arguments or piping to specify the file path to read. Should the file path change in the future, simply update the value of the `inputFilePath` constant in `main.go`.

## Cross-Compiling for EV3

Golang provides an excellent cross-compiler, allowing you to compile for any platform, from any platform, greatly easing the development process.

### Flags and Environment Variables

The following values must be set in the environment variables before building the program:

| Environment Variable | Value   | Description                          |
| -------------------- | ------- | ------------------------------------ |
| `GOOS`               | `linux` | The operating system to compile for. |
| `GOARCH`             | `arm`   | The architecture to compile for.     |
| `GOARM`              | `5`     | The ARM architecture to compile for. |

**Step 1: Clone the `ev3dev` repository:**

```bash
git clone https://github.com/RichDom2185/go_ev3_qrcode_generator.git
cd go_ev3_qrcode_generator
```

**Step 2a: From Windows command line:**

```cmd
set GOOS=linux
set GOARCH=arm
set GOARM=5
go build main.go
```

**Step 2b: From Linux terminal:**

```bash
$ GOOS=linux GOARCH=arm GOARM=5 go build main.go
```
