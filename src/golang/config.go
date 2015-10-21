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

//################//
//### CRC type ###//
//################//

type CRCType int

const (
	CRC16 = 1 << iota
	CRC32 = 1 << iota
)

//###################//
//### Config type ###//
//###################//

// A Config represents the ANTS port configuration.
type Config struct {
	// DataMessageCRCType specifies the used CRC checksum for data messages.
	// The default is CRC16.
	DataMessageCRC CRCType
}

//###############//
//### Private ###//
//###############//

// setDefaults sets the default values for unset variables.
func (c *Config) setDefaults() {
	if c.DataMessageCRC != CRC16 && c.DataMessageCRC != CRC32 {
		c.DataMessageCRC = CRC16
	}
}
