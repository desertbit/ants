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

package loopback

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoopback(t *testing.T) {
	for x := 0; x < 100; x++ {
		l := New()
		testCounts := 1000
		data := []byte("Hello World\n")
		dataLen := len(data)

		var wg sync.WaitGroup
		wg.Add(2)

		// Write loop.
		go func() {
			defer wg.Done()

			for i := 0; i < testCounts; i++ {
				n, err := l.Write(data)
				require.NoError(t, err, "(write error)")
				require.Equal(t, n, dataLen)
			}
		}()

		// Read loop.
		go func() {
			defer wg.Done()

			counter := 0

			timeout := time.NewTimer(3 * time.Second)
			defer timeout.Stop()

			var buffer []byte

			for {
				if counter == testCounts {
					break
				}

				// Timeout:
				select {
				case <-timeout.C:
					t.Fatalf("timeout reached: counter: %v", counter)
				default:
				}

				buf := make([]byte, 512)
				n, err := l.Read(buf)
				require.NoError(t, err, "(read error)")

				if n == 0 {
					time.Sleep(10 * time.Millisecond)
					continue
				}

				buffer = append(buffer, buf[:n]...)

				for len(buffer) >= dataLen {
					d := buffer[:dataLen]
					buffer = buffer[dataLen:]

					require.True(t, string(d) == string(data), "read data invalid: counter=%v %v != %v", counter, d, data)
					counter++
				}
			}
		}()

		// Wait for finished.
		wg.Wait()

		err := l.Close()
		require.NoError(t, err)

	}
}
