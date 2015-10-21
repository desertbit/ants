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

package ants

import (
	"encoding/binary"
	"hash/crc32"
	"sync"

	"github.com/howeyc/crc16"
)

const (
	crc16Polynomial = 0x8408
	crc32Polynomial = 0xeb31d82e
)

var (
	_crc16Validator *crc16Validator
	_crc32Validator *crc32Validator
	crcMutex        sync.Mutex
)

//##############################//
//### crcValidator interface ###//
//##############################//

type crcValidator interface {
	Validate(data []byte, rawCRC []byte) bool
	Checksum(data []byte) (rawCRC []byte)
}

//#############################//
//### CRC-16 implementation ###//
//#############################//

type crc16Validator struct {
	table *crc16.Table
}

func getCRC16Validator() *crc16Validator {
	// Lock the mutex.
	crcMutex.Lock()
	defer crcMutex.Unlock()

	// If already created, return it.
	if _crc16Validator != nil {
		return _crc16Validator
	}

	// Create a new validator.
	_crc16Validator = &crc16Validator{
		table: crc16.MakeTable(crc16Polynomial),
	}

	return _crc16Validator
}

func (c *crc16Validator) Validate(data []byte, rawCRC []byte) bool {
	// Convert the raw CRC byte slice.
	origCRC := binary.LittleEndian.Uint16(rawCRC)

	// Calculate the CRC checksum of data.
	crc := crc16.Checksum(data, c.table)

	// Compare the checksums.
	return crc == origCRC
}

func (c *crc16Validator) Checksum(data []byte) (rawCRC []byte) {
	// Calculate the CRC checksum of data.
	crc := crc16.Checksum(data, c.table)

	// Transform to a byte slice.
	rawCRC = make([]byte, 2)
	binary.LittleEndian.PutUint16(rawCRC, crc)

	return rawCRC
}

//#############################//
//### CRC-32 implementation ###//
//#############################//

type crc32Validator struct {
	table *crc32.Table
}

func getCRC32Validator() *crc32Validator {
	// Lock the mutex.
	crcMutex.Lock()
	defer crcMutex.Unlock()

	// If already created, return it.
	if _crc32Validator != nil {
		return _crc32Validator
	}

	// Create a new validator.
	_crc32Validator = &crc32Validator{
		table: crc32.MakeTable(crc32Polynomial),
	}

	return _crc32Validator
}

func (c *crc32Validator) Validate(data []byte, rawCRC []byte) bool {
	// Convert the raw CRC byte slice.
	origCRC := binary.LittleEndian.Uint32(rawCRC)

	// Calculate the CRC checksum of data.
	crc := crc32.Checksum(data, c.table)

	// Compare the checksums.
	return crc == origCRC
}

func (c *crc32Validator) Checksum(data []byte) (rawCRC []byte) {
	// Calculate the CRC checksum of data.
	crc := crc32.Checksum(data, c.table)

	// Transform to a byte slice.
	rawCRC = make([]byte, 4)
	binary.LittleEndian.PutUint32(rawCRC, crc)

	return rawCRC
}
