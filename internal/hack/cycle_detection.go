// Copyright 2026 Pouya Vedadiyan. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hack

import "reflect"

type (
	// circularReferenceDetector records the struct types currently being
	// rewritten. Each entry maps a source type to the location where its
	// rewritten type will be stored once it is built.
	circularReferenceDetector map[reflect.Type]*reflect.Type

	// circularReferenceHandlers holds deferred patches for circular fields,
	// applied after all types have been built.
	circularReferenceHandlers []func() error
)

// track marks typ as in progress. finalType points to where its rewritten
// type will be stored.
func (crd circularReferenceDetector) track(typ reflect.Type, finalType *reflect.Type) {
	crd[typ] = finalType
}

// has reports whether typ is in progress and, if so, returns the location
// of its rewritten type.
func (crd circularReferenceDetector) has(typ reflect.Type) (*reflect.Type, bool) {
	out, ok := crd[typ]
	return out, ok
}

// release marks typ as no longer in progress.
func (crd circularReferenceDetector) release(typ reflect.Type) {
	delete(crd, typ)
}

// inject runs each handler in order and stops at the first error.
func (crh circularReferenceHandlers) inject() error {
	for _, fn := range crh {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
