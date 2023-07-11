// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"time"

	"github.com/briandowns/spinner"
)

var DefaultSpinner = spinner.New(spinner.CharSets[9], 200*time.Millisecond)
