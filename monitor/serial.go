package main

import (
	"errors"
	"time"
)

// DefaultSize e for Config.Size
const DefaultSize = 8

// StopBits type
type StopBits byte

// Parity type
type Parity byte

const (
	// Stop1 1 stop bits
	Stop1 StopBits = 1
	// Stop1Half 1.5 stop bits
	Stop1Half StopBits = 15
	// Stop2 2 stop bits
	Stop2 StopBits = 2
)

const (
	// ParityNone no parity
	ParityNone Parity = 'N'
	// ParityOdd odd parity
	ParityOdd Parity = 'O'
	// ParityEven even parity
	ParityEven Parity = 'E'
	// ParityMark mark parity
	ParityMark Parity = 'M' // parity bit is always 1
	// ParitySpace space parity
	ParitySpace Parity = 'S' // parity bit is always 0
)

// Config structure for modem
type Config struct {
	Name        string
	Baud        int
	ReadTimeout time.Duration // Total timeout

	// Size is the number of data bits. If 0, DefaultSize is used.
	Size byte

	// Parity is the bit to use and defaults to ParityNone (no parity bit).
	Parity Parity

	// Number of stop bits to use. Default is 1 (1 stop bit).
	StopBits StopBits

	// RTSFlowControl bool

	// DTRFlowControl bool

	// XONFlowControl bool

	// CRLFTranslate bool

}

// ErrBadSize is returned if Size is not supported.
var ErrBadSize = errors.New("unsupported serial data size")

// ErrBadStopBits is returned if the specified StopBits setting not supported.
var ErrBadStopBits = errors.New("unsupported stop bit setting")

// ErrBadParity is returned if the parity is not supported.
var ErrBadParity = errors.New("unsupported parity setting")

// OpenPort opens a serial port with the specified configuration
func OpenPort(c *Config) (*Port, error) {

	size, par, stop := c.Size, c.Parity, c.StopBits

	if size == 0 {
		size = DefaultSize
	}

	if par == 0 {
		par = ParityNone
	}

	if stop == 0 {
		stop = Stop1
	}

	return openPort(c.Name, c.Baud, size, par, stop, c.ReadTimeout)

}

// Converts the timeout values for Linux / POSIX systems

func posixTimeoutValues(readTimeout time.Duration) (vmin uint8, vtime uint8) {

	const MAXUINT8 = 1<<8 - 1 // 255

	// set blocking / non-blocking read

	var minBytesToRead uint8 = 1

	var readTimeoutInDeci int64

	if readTimeout > 0 {

		// EOF on zero read

		minBytesToRead = 0

		// convert timeout to deciseconds as expected by VTIME

		readTimeoutInDeci = (readTimeout.Nanoseconds() / 1e6 / 100)

		// capping the timeout

		if readTimeoutInDeci < 1 {

			// min possible timeout 1 Deciseconds (0.1s)

			readTimeoutInDeci = 1

		} else if readTimeoutInDeci > MAXUINT8 {

			// max possible timeout is 255 deciseconds (25.5s)

			readTimeoutInDeci = MAXUINT8

		}

	}

	return minBytesToRead, uint8(readTimeoutInDeci)

}
