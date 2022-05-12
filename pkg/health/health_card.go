/*
Copyright AppsCode Inc. and Contributors

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

package health

import "k8s.io/klog/v2"

type HealthCard struct {
	lastFailure  HealthCheckFailureLabel
	totalFailure int32
	threshold    int32
	clientCount  int32
}

func newHealthCard(threshold int32) *HealthCard {
	return &HealthCard{
		threshold: threshold,
	}
}

// HasFailed returns true or false based on the threshold.
// Update the health check condition if this function returns true.
func (hcf *HealthCard) HasFailed() bool {
	klog.V(5).Infof("failure = %s, total = %d", hcf.lastFailure, hcf.totalFailure)
	return hcf.totalFailure >= hcf.threshold
}

// Register is used to register a specific failure.
// Call this method with specific label when an error is received in the health check.
func (hcf *HealthCard) Register(label HealthCheckFailureLabel) {
	if hcf.lastFailure == label {
		hcf.totalFailure++
	} else {
		hcf.totalFailure = 1
	}
	hcf.lastFailure = label
}

// Clear is used to reset the error counter.
// Call this method after each successful health check.
func (hcf *HealthCard) Clear() {
	hcf.totalFailure = 0
	hcf.lastFailure = ""
}

// ClientCreated is used to track the client which are created on the health check.
// Call this method after a client is successfully created in the health check.
func (hcf *HealthCard) ClientCreated() {
	hcf.clientCount++
}

// ClientClosed is used to track the client which are closed on the health check.
// Call this method after a client is successfully closed in the health check.
func (hcf *HealthCard) ClientClosed() {
	hcf.clientCount--
}

// GetClientCount is used to get the current open client count.
// This should always be 0.
func (hcf *HealthCard) GetClientCount() int32 {
	return hcf.clientCount
}
