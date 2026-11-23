package hack

import "reflect"

type (
	circularReferenceDetector map[reflect.Type]*reflect.Type
	circularReferenceHandlers []func() error
)

func (crd circularReferenceDetector) track(typ reflect.Type, finalType *reflect.Type) {
	crd[typ] = finalType
}

func (crd circularReferenceDetector) has(typ reflect.Type) (*reflect.Type, bool) {
	out, ok := crd[typ]
	return out, ok
}

func (crd circularReferenceDetector) release(typ reflect.Type) {
	delete(crd, typ)
}

func (crh circularReferenceHandlers) inject() error {
	for _, fn := range crh {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
