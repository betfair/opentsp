// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package restart

import (
	"os"
)

// BUG(masiulaniecj): On Windows, Do is equivalent to os.Exit(0).

func Do() { os.Exit(0) }
