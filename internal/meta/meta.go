// Copyright 2021 Upbound Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package meta contains constants and types that are used across the bootstrapper.
package meta

const (
	// LabelKeyManagedBy is the key for the label indicating resource is managed by bootstrapper.
	LabelKeyManagedBy = "upbound.io/managed-by"
	// LabelValueManagedBy is the value for the label indicating resource is managed by bootstrapper.
	LabelValueManagedBy = "bootstrapper"
	// SecretNameEntitlement is the name of the Secret that contains the tokens
	// stored for entitlement of usage of Universal Crossplane.
	SecretNameEntitlement = "upbound-entitlement"
)
