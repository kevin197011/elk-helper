// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package executor

import "errors"

// ErrSkippedInterval means the rule tick fired but the interval gate has not elapsed yet.
// This is not a failure; the scheduler must not log it as "executed successfully".
var ErrSkippedInterval = errors.New("execution skipped: interval not elapsed")
