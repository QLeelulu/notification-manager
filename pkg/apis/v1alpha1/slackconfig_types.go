/*


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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SlackConfigSpec defines the desired state of SlackConfig
type SlackConfigSpec struct {
	// The URL to send HTTP POST requests to.
	// SlackApiUrlSecret takes precedence over SlackApiUrl. One of SlackApiUrlSecret and SlackApiUrl should be defined.
	SlackApiUrlSecret *v1.SecretKeySelector `json:"slackApiUrlSecret,omitempty"`
	// The API URL to use for Slack notifications.
	SlackApiUrl *string `json:"slackApiUrl,omitempty"`
	// HttpConfig used for slack
	HttpConfigSelector *metav1.LabelSelector `json:"httpConfigSelector,omitempty"`
}

// SlackConfigStatus defines the observed state of SlackConfig
type SlackConfigStatus struct {
}

// +kubebuilder:object:root=true

// SlackConfig is the Schema for the slackconfigs API
type SlackConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlackConfigSpec   `json:"spec,omitempty"`
	Status SlackConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SlackConfigList contains a list of SlackConfig
type SlackConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SlackConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SlackConfig{}, &SlackConfigList{})
}