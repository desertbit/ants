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

// Package serial is a small wrapper to provide an io.ReadWriteCloser interface
// for serial port communication for the ANTS library.
package serial

import (
	"fmt"
	"io"

	"github.com/tarm/serial"
)

// OpenPort opens a serial port with the config and
// returns an io.ReadWriteCloser interface.
func OpenPort(config *Config) (io.ReadWriteCloser, error) {
	// Set the default config values for unset values.
	config.setDefaults()

	// Create the serial configuration.
	c := &serial.Config{
		Name:        config.Name,
		Baud:        config.Baud,
		ReadTimeout: config.ReadTimeout,
	}

	// Open the serial port.
	serialPort, err := serial.OpenPort(c)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %v", err)
	}

	return serialPort, nil
}
