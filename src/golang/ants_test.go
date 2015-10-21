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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDLEEscaping(t *testing.T) {
	data := []byte{dle, dle, dle, 0, dle, dle, 0, dle, 0, 0, dle, 0, dle, dle, dle, dle, dle, 0, dle, dle, dle, dle, 0, 0, dle, dle, 0, dle, dle, 0, dle, dle, dle}

	d := escapeDLE(data)
	d = unescapeDLE(d)

	require.True(t, len(d) == len(data))

	for i, b := range data {
		require.True(t, b == d[i])
	}
}
