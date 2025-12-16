/*
Copyright The Volcano Authors.

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

package controller

type Config struct {
	EnableLeaderElection bool
	Workers              int
	Kubeconfig           string
	MasterURL            string
	Controllers          []string
}

// shouldEnableController checks if a specific controller should be enabled
func (cc Config) shouldEnableController(controllerName string) bool {
	// If no controllers are specified, enable all by default
	if len(cc.Controllers) == 0 {
		return true
	}

	// Check if the specified controller is in the list
	for _, ctrl := range cc.Controllers {
		if ctrl == controllerName {
			return true
		}
	}
	return false
}

// shouldEnableAnyController checks if any of the controllers should be enabled
func (cc Config) shouldEnableAnyController() bool {
	return len(cc.Controllers) == 0 ||
		cc.shouldEnableController(ModelServingController) ||
		cc.shouldEnableController(ModelBoosterController) ||
		cc.shouldEnableController(AutoscalerController)
}
