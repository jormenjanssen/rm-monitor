package main

import (
	"context"
	"io"
)

// NewAsioReader creates new asio reader
func NewAsioReader() (readCloser io.ReadCloser) {

	return &AsioReader{readCloser: readCloser,
		readCallChannel:     make(chan ReadCall),
		readCallbackChannel: make(chan ReadCall),
		context:             context.Background()}
}

// AsioReader structure
type AsioReader struct {
	readCloser          io.ReadCloser
	readCallChannel     chan ReadCall
	readCallbackChannel chan ReadCall
	context             context.Context
}

// ReadCall type
type ReadCall struct {
	data   []byte
	length int
	fr     bool
	err    error
}

// HasStarted check if the read call is finished
func (readCall *ReadCall) HasStarted() bool {
	return readCall.length != unreadMagic
}

// HasFinished check if the read call is finished
func (readCall *ReadCall) HasFinished() bool {
	return readCall.fr
}

// Finish the readcall
func (readCall *ReadCall) Finish(data []byte, l int, err error) {
	readCall.fr = true
}

// Close method
func (asioReader *AsioReader) Close() error {
	return asioReader.readCloser.Close()
}

var unreadMagic = -99997532
var readStartedMagic = -999971

// Read method
func (asioReader *AsioReader) Read(p []byte) (int, error) {

	// Push our read request to our channel
	rr := ReadCall{data: p,
		length: unreadMagic}
	asioReader.readCallChannel <- rr

	// Read message back from channel
	rc := <-asioReader.readCallbackChannel

	return rc.length, rc.err
}

func (asioReader *AsioReader) internalReadFunc(ctx context.Context) {

	for {

		activeCall := ReadCall{length: unreadMagic}

		select {

		case rr := <-asioReader.readCallChannel:

			// Finish our call early if we got one previous.
			if activeCall.HasStarted() && activeCall.HasFinished() {
				asioReader.readCallbackChannel <- activeCall
				return
				// Todo: try to complete the current within timeout range
			} else if activeCall.HasStarted() {
				// Todo: Discard the active call after we done
				// Todo: We are the active call
			} else {
				// Todo: try to complete us if within time else

				// Replace the active call with the one we got from the channel.
				activeCall = rr
				activeCall.length = readStartedMagic

				// Start the actual read
				l, err := asioReader.Read(rr.data)

				asioReader.readCallbackChannel <- ReadCall{
					data:   rr.data,
					length: l,
					err:    err}
			}

			// Todo case time
		case <-ctx.Done():
			return
		}
	}

}
