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

package serial

import (
	"time"
)

// A Config represents the serial port configuration.
type Config struct {
	// Name specifies the port name or path.
	Name string

	// Baud specifies the Baudrate.
	Baud int

	// The total read timeout of one data chunk.
	// The default value is 5 Seconds.
	ReadTimeout time.Duration
}

//###############//
//### Private ###//
//###############//

// setDefaults sets the default values for unset variables.
func (c *Config) setDefaults() {
	// Set the read timeout to the default value if not set.
	if int64(c.ReadTimeout) <= 0 {
		c.ReadTimeout = 1 * time.Second
	}
}
