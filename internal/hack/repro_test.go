package hack

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// reset clears the global type cache so each test starts clean.
func reset() { typeCache = map[reflect.Type]map[LoopHandlerPolicy]reflect.Type{} }

// catch runs f and returns any panic value.
func catch(f func()) (p any) {
	defer func() { p = recover() }()
	f()
	return nil
}

// H1: the direct Metadata form panics (or drops the field).
type valuesH1 struct {
	Metadata `bson:"values,omitempty"`
	SomeKey  string
}
type docH1 struct {
	Values valuesH1
	This   string
}

func TestH1_MetadataDirectForm(t *testing.T) {
	reset()
	var typ reflect.Type
	var err error
	if p := catch(func() { typ, err = Rewrite(reflect.TypeFor[docH1](), LoopHandlerPolicyBlock) }); p != nil {
		t.Fatalf("Rewrite panicked: %v", p)
	}
	if err != nil {
		t.Fatal(err)
	}
	if got := typ.Field(0).Tag.Get("bson"); got != "values,omitempty" {
		t.Fatalf("Values tag = %q, want %q", got, "values,omitempty")
	}
}

// H2: an embedded *T that carries Metadata panics.
type taggerH2 struct {
	Metadata `bson:"x"`
}
type withPtrH2 struct {
	*taggerH2
	Name string
}
type docH2 struct{ F withPtrH2 }

func TestH2_MetadataEmbeddedPointer(t *testing.T) {
	reset()
	var typ reflect.Type
	var err error
	if p := catch(func() { typ, err = Rewrite(reflect.TypeFor[docH2](), LoopHandlerPolicyBlock) }); p != nil {
		t.Fatalf("Rewrite panicked: %v", p)
	}
	if err != nil {
		t.Fatal(err)
	}
	if got := typ.Field(0).Tag.Get("bson"); got != "x" {
		t.Fatalf("F tag = %q, want %q", got, "x")
	}
}

// H3: types with methods (time.Time) are rewritten and lose their encoding.
type docH3 struct {
	At time.Time `bson:",opaque"`
}

func TestH3_TimeNotRewritten(t *testing.T) {
	reset()
	typ, err := Rewrite(reflect.TypeFor[docH3](), LoopHandlerPolicyBlock)
	if err != nil {
		t.Fatal(err)
	}
	if got := typ.Field(0).Type; got != reflect.TypeFor[time.Time]() {
		t.Fatalf("At rewritten to %s; time.Time must be left unchanged", got)
	}
}

// H4: the cache ignores the policy.
type nodeH4 struct{ Next *nodeH4 }

func TestH4_CacheIgnoresPolicy(t *testing.T) {
	reset()
	if _, err := Rewrite(reflect.TypeFor[nodeH4](), LoopHandlerPolicyIgnore); err != nil {
		t.Fatal(err)
	}
	if _, err := Rewrite(reflect.TypeFor[nodeH4](), LoopHandlerPolicyBlock); err == nil {
		t.Fatal("Block policy returned no error for a circular type (cached Ignore result reused)")
	}
}

// H5: a failed rewrite leaves half-built types in the cache.
type nodeH5 struct {
	Next *nodeH5
	V    int
}

func TestH5_FailedRewriteCached(t *testing.T) {
	reset()
	EnableAbiPatch(false)
	if _, err := Rewrite(reflect.TypeFor[nodeH5](), LoopHandlerPolicyHandle); err == nil {
		t.Fatal("expected an error with ABI patching disabled")
	}
	if cached, ok := typeCache[reflect.TypeFor[nodeH5]()]; ok {
		t.Fatalf("failed rewrite left %v in the cache", cached)
	}
}

// H6: results depend on which type is rewritten first.
type aH6 struct {
	B *bH6
	X int
}
type bH6 struct {
	A *aH6
	Y int
}

func TestH6_OrderDependentResults(t *testing.T) {
	reset()
	alone, err := Rewrite(reflect.TypeFor[bH6](), LoopHandlerPolicyIgnore)
	if err != nil {
		t.Fatal(err)
	}
	reset()
	if _, err := Rewrite(reflect.TypeFor[aH6](), LoopHandlerPolicyIgnore); err != nil {
		t.Fatal(err)
	}
	after, err := Rewrite(reflect.TypeFor[bH6](), LoopHandlerPolicyIgnore)
	if err != nil {
		t.Fatal(err)
	}
	if alone.String() != after.String() {
		t.Fatalf("Rewrite(bH6) depends on call order:\n alone: %s\n after: %s", alone, after)
	}
}

// H7: patching a slice-typed circular field breaks the struct layout.
type treeH7 struct {
	Children []treeH7
	V        int
}

func TestH7_PatchLayoutMismatch(t *testing.T) {
	reset()
	if err := EnableAbiPatch(true); err != nil {
		t.Skip("ABI patching unsupported:", err)
	}
	defer EnableAbiPatch(false)
	typ, err := Rewrite(reflect.TypeFor[treeH7](), LoopHandlerPolicyHandle)
	if err != nil {
		return // rejecting the field is an acceptable fix
	}
	if src := reflect.TypeFor[treeH7](); typ.Size() != src.Size() || typ.Field(1).Offset != src.Field(1).Offset {
		t.Fatalf("layout mismatch: rewritten size %d (V at %d), original size %d (V at %d)",
			typ.Size(), typ.Field(1).Offset, src.Size(), src.Field(1).Offset)
	}
}

// H8: identical rewritten structs share one runtime type, so patches collide.
type aH8 struct{ Next *aH8 }
type bH8 struct{ Next *bH8 }
type pairH8 struct {
	A aH8
	B bH8
}

func TestH8_SharedRuntimeType(t *testing.T) {
	reset()
	if err := EnableAbiPatch(true); err != nil {
		t.Skip("ABI patching unsupported:", err)
	}
	defer EnableAbiPatch(false)
	typ, err := Rewrite(reflect.TypeFor[pairH8](), LoopHandlerPolicyHandle)
	if err != nil {
		t.Fatal(err)
	}
	a, b := typ.Field(0).Type, typ.Field(1).Type
	if a.Field(0).Type.Elem() != a {
		t.Errorf("A.Next points to %s, want *A'", a.Field(0).Type)
	}
	if b.Field(0).Type.Elem() != b {
		t.Errorf("B.Next points to %s, want *B'", b.Field(0).Type)
	}
	if a == b {
		t.Errorf("rewritten A and B are the same runtime type")
	}
}

// H9: EnableAbiPatch races with Rewrite. Run with: go test -race -run H9
type nodeH9 struct{ Next *nodeH9 }

func TestH9_EnableAbiPatchRace(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); EnableAbiPatch(i%2 == 0) }()
		go func() {
			defer wg.Done()
			Rewrite(reflect.TypeFor[nodeH9](), LoopHandlerPolicyHandle)
		}()
	}
	wg.Wait()
	EnableAbiPatch(false)
}

// H11: inlining two fields with the same BSON key is not detected.
type IDA struct {
	ID string `bson:"id"`
}
type IDB struct {
	ID string `bson:"id"`
}
type docH11 struct {
	IDA
	IDB
}

func TestH11_DuplicateKeys(t *testing.T) {
	reset()
	if _, err := Rewrite(reflect.TypeFor[docH11](), LoopHandlerPolicyBlock); err == nil {
		t.Fatal(`two inlined fields map to key "id"; Rewrite should return an error`)
	}
}

// H13: tags on embedded fields are replaced, dropping json/validate.
type BaseH13 struct{ X int }
type docH13 struct {
	BaseH13 `json:"base" validate:"required"`
}

func TestH13_TagsClobbered(t *testing.T) {
	reset()
	typ, err := Rewrite(reflect.TypeFor[docH13](), LoopHandlerPolicyBlock)
	if err != nil {
		t.Fatal(err)
	}
	tag := typ.Field(0).Tag
	if tag.Get("json") != "base" || tag.Get("validate") != "required" {
		t.Fatalf("tag = %q; json and validate keys were dropped", tag)
	}
}

// H14: an embedded circular pointer under Ignore.
type SelfH14 struct {
	*SelfH14
	V int
}

func TestH14_EmbeddedCircularIgnore(t *testing.T) {
	reset()
	var err error
	if p := catch(func() { _, err = Rewrite(reflect.TypeFor[SelfH14](), LoopHandlerPolicyIgnore) }); p != nil {
		t.Fatalf("Rewrite panicked: %v", p)
	}
	if err != nil {
		t.Fatal(err)
	}
}

// H15: bson:"opaque" (no comma) is accepted but makes the key "opaque".
type docH15 struct {
	At time.Time `bson:"opaque"`
}

func TestH15_OpaqueWithoutComma(t *testing.T) {
	reset()
	typ, err := Rewrite(reflect.TypeFor[docH15](), LoopHandlerPolicyBlock)
	if err != nil {
		return // rejecting the tag is an acceptable fix
	}
	if name, _, _ := strings.Cut(typ.Field(0).Tag.Get("bson"), ","); name == "opaque" {
		t.Fatal(`field At will be stored under the key "opaque"; use bson:",opaque" or reject the tag`)
	}
}

// H16: struct values inside maps are not rewritten.
type BaseH16 struct{ X int }
type InnerH16 struct{ BaseH16 }
type docH16 struct{ M map[string]InnerH16 }

func TestH16_MapsNotRewritten(t *testing.T) {
	reset()
	typ, err := Rewrite(reflect.TypeFor[docH16](), LoopHandlerPolicyBlock)
	if err != nil {
		t.Fatal(err)
	}
	if typ.Field(0).Type.Elem() == reflect.TypeFor[InnerH16]() {
		t.Fatal("map element type was not rewritten, so its embedded fields won't be inlined")
	}
}

// H17: self-referential wrapper types loop forever.
type sliceH17 []sliceH17

func TestH17_SelfReferentialWrapper(t *testing.T) {
	done := make(chan struct{})
	go func() { concreteType(reflect.TypeFor[sliceH17]()); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("concreteType did not return for type S []S (infinite loop)")
	}
}

// H18: errors don't identify the type or field.
type nodeH18 struct{ Next *nodeH18 }

func TestH18_ErrorContext(t *testing.T) {
	reset()
	_, err := Rewrite(reflect.TypeFor[nodeH18](), LoopHandlerPolicyBlock)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "nodeH18") || !strings.Contains(err.Error(), "Next") {
		t.Fatalf("error %q does not name the type and field", err)
	}
}
