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
	// LoopDetector records the struct types currently being
	// rewritten. Each entry maps a source type to the location where its
	// rewritten type will be stored once it is built.
	LoopDetector map[reflect.Type]*reflect.Type

	// LoopHandler holds deferred patches for circular fields,
	// applied after all types have been built.
	LoopHandler []func() error
)

// track marks typ as in progress. finalType points to where its rewritten
// type will be stored.
func (ld LoopDetector) track(typ reflect.Type, finalType *reflect.Type) {
	if _, ok := ld[typ]; ok {
		return
	}
	ld[typ] = finalType
}

// has reports whether typ is in progress and, if so, returns the location
// of its rewritten type.
func (ld LoopDetector) has(typ reflect.Type) (*reflect.Type, bool) {
	out, ok := ld[typ]
	return out, ok
}

// release marks typ as no longer in progress.
func (ld LoopDetector) release(typ reflect.Type) {
	delete(ld, typ)
}

// inject runs each handler in order and stops at the first error.
func (lh LoopHandler) inject() error {
	for _, fn := range lh {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
