/*
 *  Ants - Let the ants handle your serial communication.
 *  Copyright (C) 2015  Roland Singer <roland.singer[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package ants - Let the ants handle your serial communication.
// This implementation is thread-safe.
package ants

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

//#################//
//### Constants ###//
//#################//

const (
	readChanSize     = 25
	readBufferSize   = 512
	readWaitDuration = 50 * time.Millisecond

	maxMessageSize     = 2048 // In bytes.
	readMessageTimeout = 5 * time.Second

	readControlMessageChanSize = 3
	readDataChunkChanSize      = 5
	writeDataChunkChanSize     = 5

	// Protocol constants:
	dle  = 0x10
	umsn = 0 // Unknown message sequence number (UMSN)

	// Protocol control characters:
	stx = 0x02
	etx = 0x03
	ack = 0x06
	nak = 0x15
)

//#################//
//### Variables ###//
//#################//

// Errors:
var (
	// ErrTimeout is thrown if a timeout is reached.
	ErrTimeout = errors.New("timeout reached")

	// ErrClosed is thrown if the port is closed.
	ErrClosed = errors.New("port closed")
)

//#############################//
//### Control Message type ###//
//#############################//

type controlMessage struct {
	TypeCharacter byte
	MSN           byte // Message sequence number.
}

//#################//
//### Port type ###//
//#################//

// A Port is an open port which reads and writes from a source.
type Port struct {
	source io.ReadWriteCloser

	isClosed   bool
	closeChan  chan struct{}
	closeMutex sync.Mutex

	readChan               chan byte
	readBinaryDataBuffer   []byte
	readControlMessageChan chan controlMessage

	readDataChunkChan  chan []byte
	writeDataChunkChan chan []byte

	crc16Validator          crcValidator
	dataMessageCRCValidator crcValidator
	dataMessageCRCLength    int // Bytes counted.
}

// NewPort creates and returns a new ANTS port.
// This method expects an io.ReadWriteCloser interface as source.
// Optionally pass a configuration.
func NewPort(source io.ReadWriteCloser, config ...*Config) *Port {
	// Get the config.
	var c *Config
	if len(config) > 0 {
		c = config[0]
	} else {
		c = new(Config)
	}

	// Set the default config values for unset variables.
	c.setDefaults()

	// Create a new port.
	p := &Port{
		source:                 source,
		closeChan:              make(chan struct{}),
		readChan:               make(chan byte, readChanSize),
		readControlMessageChan: make(chan controlMessage, readControlMessageChanSize),
		readDataChunkChan:      make(chan []byte, readDataChunkChanSize),
		writeDataChunkChan:     make(chan []byte, writeDataChunkChanSize),
		crc16Validator:         getCRC16Validator(),
	}

	// Set the data message CRC length depending on the config CRC type.
	// Also set the CRC validator.
	if c.DataMessageCRC == CRC32 {
		p.dataMessageCRCValidator = getCRC32Validator()
		p.dataMessageCRCLength = 4
	} else {
		p.dataMessageCRCValidator = getCRC16Validator()
		p.dataMessageCRCLength = 2
	}

	// Start the loop goroutines.
	go p.readFromSourceLoop()
	go p.readMessagesLoop()
	go p.writeDataMessagesLoop()

	return p
}

// IsClosed returns a boolean whenever the port is closed.
func (p *Port) IsClosed() bool {
	return p.isClosed
}

// Close the serial port.
func (p *Port) Close() error {
	// Lock the mutex.
	p.closeMutex.Lock()
	defer p.closeMutex.Unlock()

	// Return if already closed.
	if p.isClosed {
		return nil
	}

	// Set the flag.
	p.isClosed = true

	// Close the close channel.
	close(p.closeChan)

	// Close the source
	err := p.source.Close()
	if err != nil {
		return fmt.Errorf("failed to close port's source: %v", err)
	}

	return nil
}

// Read a verified data chunk from the serial port.
// Optionally pass a timeout duration.
// If the timeout is reached, then ErrTimeout is returned.
// If the port is closed, then ErrClosed is returned.
func (p *Port) Read(timeout ...time.Duration) (data []byte, err error) {
	timeoutChan := make(chan (struct{}))

	// Create a timeout timer if a timeout is specified.
	if len(timeout) > 0 && timeout[0] > 0 {
		timer := time.AfterFunc(timeout[0], func() {
			// Trigger the timeout by closing the channel.
			close(timeoutChan)
		})

		// Always stop the timer on defer.
		defer timer.Stop()
	}

	// Read from the data channel or timeout.
	select {
	case <-p.closeChan:
		return nil, ErrClosed
	case <-timeoutChan:
		return nil, ErrTimeout
	case data = <-p.readDataChunkChan:
		return data, nil
	}
}

// Write a data chunk to the port.
// If the port is closed, then ErrClosed is returned.
func (p *Port) Write(data []byte) error {
	if p.isClosed {
		return ErrClosed
	}

	// Just write to the channel.
	p.writeDataChunkChan <- data

	return nil
}

//#######################//
//### Private methods ###//
//#######################//

func (p *Port) closeAndLogError() {
	err := p.Close()
	if err != nil {
		Log.Errorf("failed to close port: %v", err)
	}
}

func (p *Port) writeDataMessagesLoop() {
	for {
		select {
		case <-p.closeChan:
			// Just release this goroutine if the port is closed.
			return
		case data := <-p.writeDataChunkChan:
			// Escape the data.
			data = escapeDLE(data)

			// Prepend the escaped STX control character.
			data = append([]byte{dle, stx}, data...)

			// Calculate the CRC checksum.
			crc := p.dataMessageCRCValidator.Checksum(data)

			// Escape the CRC.
			crc = escapeDLE(crc)

			// Append the CRC.
			data = append(data, crc...)

			// Append the escaped ETX control character.
			data = append(data, []byte{dle, etx}...)

			// Resend the data until an acknowledge control character is received.
		ResendLoop:
			for {
				// Write the data to the source.
				err := p.writeToSource(data)
				if err != nil {
					// Log the error and close the port.
					Log.Errorf("failed to write data to the source: %v", err)
					p.closeAndLogError()
					return
				}

				// TODO: Add timeout.

				// Wait for a control character as response.
				select {
				case cm := <-p.readControlMessageChan:
					// Break the resend loop on a successful transmission.
					if cm.TypeCharacter == ack {
						break ResendLoop
					}

					// Otherwise resend the data.
					continue ResendLoop
				}
			}
		}
	}
}

func (p *Port) writeControlMessage(ctrlType byte, msn byte) {
	// TODO
}

// writeToSource writes the data bytes to the source.
func (p *Port) writeToSource(data []byte) (err error) {
	// Catch all panics, and return the error.
	// Panics could occur in the p.source.Write call, which is third-party code...
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: write data to source: %v", e)
		}
	}()

	// Write to the source.
	n, err := p.source.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to source: %v", err)
	}

	// Check if data was partially transmitted.
	if n != len(data) {
		// Send the escaped ETX control character and dismiss any write error.
		// Pretend as no error occurred. The peer will request a resend...
		_, _ = p.source.Write([]byte{dle, etx})

		// Log
		Log.Warningf("write data to source: failed to send complete data chunk: data was only transmitted partially")
	}

	return nil
}

func (p *Port) readFromSourceLoop() {
	// Catch all panics, log the error and close the port.
	// Panics could occur in the p.source.Read call, which is third-party code...
	defer func() {
		if e := recover(); e != nil {
			Log.Errorf("panic: read data from source: %v", e)
			p.closeAndLogError()
		}
	}()

	// The read buffer.
	buf := make([]byte, readBufferSize)

	// Read from the source as long as the port is open.
	for !p.isClosed {
		// Read data from the source.
		n, err := p.source.Read(buf)
		if err != nil && err != io.EOF {
			// Log the error and close the port.
			Log.Errorf("failed to read data from source: %v", err)
			p.closeAndLogError()
			return
		}

		// If nothing was received, then read again after a short timeout.
		if n == 0 {
			time.Sleep(readWaitDuration)
			continue
		}

		// Iterate through all received bytes and push them to the read channel.
		for _, b := range buf[:n] {
			p.readChan <- b
		}
	}
}

func (p *Port) readMessagesLoop() {
	var buf []byte
	var controlCharacter byte

	// Flags:
	isControlMessage := false
	startCharacterFound := false
	byteIsEscaped := false

	// Create a new timeout timer in a stopped state.
	timeoutTimer := time.NewTimer(readMessageTimeout)
	timeoutTimer.Stop()

	// Close the timeout always on exit.
	defer timeoutTimer.Stop()

	// Start the magic :P
	for {
		select {
		case <-p.closeChan:
			// The port was closed. Release this goroutine.
			return

		case <-timeoutTimer.C:
			// Timeout reached. Reset flags and clear message buffer.
			isControlMessage = false
			startCharacterFound = false
			byteIsEscaped = false

			controlCharacter = 0

			// Clear the buffer.
			buf = buf[:0]

			// Log
			Log.Warningf("read data: read message timeout reached: discarding data")

		case b := <-p.readChan:
			// Anonymous function for defers.
			func() {
				// Hint: This protocol uses the Data Link Escape (DLE) character to
				// differentiate between control characters and the binary data transmission.
				// Control characters are preceded with the DLE character.
				// Whenever the DLE character is encountered in the data, it is
				// sent twice to prevent the byte that follows from being interpreted
				// as a control character.
				//
				// Set the escaped flag.
				if !byteIsEscaped && b == dle {
					byteIsEscaped = true
					return
				}

				// Always reset the esape flag on defer.
				defer func() {
					byteIsEscaped = false
				}()

				// Check for control characters. They have to be escaped.
				if byteIsEscaped {
					// Check if the byte is a start character, if searching for it.
					if !startCharacterFound {
						if b == stx || b == ack || b == nak {
							// Set the flag.
							if b == stx {
								isControlMessage = false
							} else {
								isControlMessage = true

								// Save the control message character.
								controlCharacter = b
							}

							// Set the flag.
							startCharacterFound = true

							// Restart the timeout timer.
							timeoutTimer.Reset(readMessageTimeout)
						} else {
							// Discard the byte, but log this occurrence.
							Log.Warningf("read data: expected start character but got other byte: %v", b)
						}

						return
					}

					// If the byte is the end character, then handle the received message body
					// and clear the buffer for the next read procedure.
					if b == etx {
						// Stop the timeout timer.
						timeoutTimer.Stop()

						// Unescape the buffer.
						buf = unescapeDLE(buf)

						// Handle the message body in a new function to keep things clear.
						if isControlMessage {
							err := p.handleReceivedControlMessageBody(controlCharacter, buf)
							if err != nil {
								Log.Warningf("read data: handle control message body: %v", err)
							}
						} else {
							err := p.handleReceivedDataMessageBody(buf)
							if err != nil {
								Log.Warningf("read data: handle data message body: %v", err)
							}
						}

						// Clear the buffer.
						buf = buf[:0]

						return
					}
				}

				// Append the new byte to the message buffer.
				buf = append(buf, b)

				// Check if the maximum buffer size is reached.
				if len(buf) > maxMessageSize {
					// Discard the received bytes and start over again.
					buf = buf[:0]

					// Log this.
					Log.Warningf("read data: maximum message buffer size of %v bytes reached: discarding message", maxMessageSize)

					return
				}
			}()
		}
	}
}

func (p *Port) handleReceivedControlMessageBody(typeCharacter byte, body []byte) (err error) {
	// Check for the required body length.
	// Message sequence number and CRC checksum have to be contained.
	// 1 Byte + 2 Bytes
	if len(body) != 3 {
		return fmt.Errorf("invalid control message body")
	}

	// Extract the CRC checksum.
	pos := len(body) - 2
	crcChecksum := body[pos:]

	// Remove the CRC checksum from the body.
	body = body[:pos]

	// Validate the the message body with the checksum.
	if !p.crc16Validator.Validate(body, crcChecksum) {
		return fmt.Errorf("message body is corrupt: message CRC checksum is invalid")
	}

	// Extract the peer message sequence number (PMSN).
	pmsn := body[0]

	// Create a new control message value.
	cm := controlMessage{
		TypeCharacter: typeCharacter,
		MSN:           pmsn,
	}

	// Push it to the channel.
	p.readControlMessageChan <- cm

	return nil
}

func (p *Port) handleReceivedDataMessageBody(body []byte) (err error) {
	// Set the peer message sequence number to the initial unknown constant.
	var pmsn byte = umsn

	// Send a control message on defer.
	// Control messages have to be send as a reply for a data message.
	defer func() {
		// Send an Acknowledge or Negative Acknowledge Control Message.
		if err != nil {
			p.writeControlMessage(nak, pmsn)
		} else {
			p.writeControlMessage(ack, pmsn)
		}
	}()

	// Check for the required minimum body length.
	// Message sequence number, append data flag and CRC checksum have to be contained.
	// 1 Byte + 1 Byte + 2/4 Bytes
	if len(body) < 2+p.dataMessageCRCLength {
		return fmt.Errorf("invalid data message body: body is too short")
	}

	// Extract the CRC checksum.
	pos := len(body) - p.dataMessageCRCLength
	crcChecksum := body[pos:]

	// Remove the CRC checksum from the body.
	body = body[:pos]

	// Validate the the message body with the checksum.
	if !p.dataMessageCRCValidator.Validate(body, crcChecksum) {
		return fmt.Errorf("message body is corrupt: message CRC checksum is invalid")
	}

	// Extract the peer message sequence number (PMSN).
	pmsn = body[0]

	// Extract the append data flag.
	appendData := body[1]

	// Extract the binary data.
	binData := body[2:]

	// Check if the binary data is send in multiple messages.
	if appendData == 0 {
		// End of binary data transmission.
		// Obtain the complete data chunk.
		data := append(p.readBinaryDataBuffer, binData...)

		// Push the data chunk to the channel.
		p.readDataChunkChan <- data

		// Clear the binary data chunk buffer.
		p.readBinaryDataBuffer = p.readBinaryDataBuffer[:0]

		// Release memory if the capacity of the buffer is huge.
		if cap(p.readBinaryDataBuffer) > 10240 {
			p.readBinaryDataBuffer = nil
		}
	} else {
		// The data message transmission is not complete.
		// Push the received binary data to the buffer.
		p.readBinaryDataBuffer = append(p.readBinaryDataBuffer, binData...)
	}

	return nil
}

//###############//
//### Private ###//
//###############//

func escapeDLE(data []byte) []byte {
	escapedData := make([]byte, 0, len(data))

	for _, b := range data {
		if b == dle {
			escapedData = append(escapedData, dle, dle)
		} else {
			escapedData = append(escapedData, b)
		}
	}

	return escapedData
}

func unescapeDLE(data []byte) []byte {
	unescapedData := make([]byte, 0, len(data))
	isEscaped := false

	for _, b := range data {
		if !isEscaped && b == dle {
			isEscaped = true
			continue
		}

		isEscaped = false

		unescapedData = append(unescapedData, b)
	}

	return unescapedData
}
