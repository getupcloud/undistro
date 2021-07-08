/*
Copyright 2020-2021 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package retry

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewBackoff() wait.Backoff {
	// Return a exponential backoff configuration which returns durations for a total time of ~40s.
	// Example: 0, .5s, 1.2s, 2.3s, 4s, 6s, 10s, 16s, 24s, 37s
	// Jitter is added as a random fraction of the duration multiplied by the jitter factor.
	return wait.Backoff{
		Duration: time.Second,
		Factor:   1.5,
		Steps:    10,
		Jitter:   0.5,
	}
}

// WithExponentialBackoff repeats an operation until it passes or the exponential backoff times out.
func WithExponentialBackoff(opts wait.Backoff, operation func() error) error {
	log := log.Log
	i := 0
	err := wait.ExponentialBackoff(opts, func() (bool, error) {
		i++
		if err := operation(); err != nil {
			if i < opts.Steps {
				log.V(5).Info("Operation failed, retrying with backoff", "Cause", err.Error())
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return errors.Wrapf(err, "action failed after %d attempts", i)
	}
	return nil
}
