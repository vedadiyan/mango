package hack

import "reflect"

type (
	CircularReferenceDetector map[reflect.Type]*reflect.Type
	CircularReferenceHandlers []func() error
)

func (crd CircularReferenceDetector) Track(typ reflect.Type, finalType *reflect.Type) {
	crd[typ] = finalType
}

func (crd CircularReferenceDetector) Has(typ reflect.Type) (*reflect.Type, bool) {
	out, ok := crd[typ]
	return out, ok
}

func (crd CircularReferenceDetector) Release(typ reflect.Type) {
	delete(crd, typ)
}

func (crh CircularReferenceHandlers) inject() error {
	for _, fn := range crh {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
