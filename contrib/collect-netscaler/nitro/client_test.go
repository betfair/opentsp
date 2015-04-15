// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package nitro

import (
	"testing"
)

func TestConnLimit(t *testing.T) {
	limit := make(connLimit, 2)

	limit.Wait()
	limit.Wait()
	limit.Done()
	limit.Done()

	limit.Wait()
	limit.Wait()
	limit.Done()
	limit.Done()

	limit.Wait()
	limit.Done()
	limit.Wait()
	limit.Done()
}
