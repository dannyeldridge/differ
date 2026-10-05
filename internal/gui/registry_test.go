package gui

import (
	"reflect"
	"testing"
)

func TestRegistryAddDedupesAndKeepsOrder(t *testing.T) {
	r := newRegistry()

	if _, added := r.add("/a"); !added {
		t.Fatal("first add of /a should report added")
	}
	if _, added := r.add("/b"); !added {
		t.Fatal("first add of /b should report added")
	}
	if _, added := r.add("/a"); added {
		t.Fatal("second add of /a should not report added")
	}

	want := []string{"/a", "/b"}
	if got := r.list(); !reflect.DeepEqual(got, want) {
		t.Fatalf("list() = %v, want %v", got, want)
	}
}

func TestRegistryAddReturnsSameStopChannelForExistingRepo(t *testing.T) {
	r := newRegistry()
	first, _ := r.add("/a")
	second, _ := r.add("/a")
	if first != second {
		t.Fatal("re-adding an open repo must not replace its stop channel")
	}
}

func TestRegistryRemoveClosesStopChannel(t *testing.T) {
	r := newRegistry()
	stop, _ := r.add("/a")
	r.add("/b")

	if !r.remove("/a") {
		t.Fatal("remove of an open repo should report true")
	}

	select {
	case <-stop:
	default:
		t.Fatal("remove should close the repo's stop channel")
	}

	want := []string{"/b"}
	if got := r.list(); !reflect.DeepEqual(got, want) {
		t.Fatalf("list() = %v, want %v", got, want)
	}
}

func TestRegistryRemoveUnknownRepo(t *testing.T) {
	r := newRegistry()
	if r.remove("/nope") {
		t.Fatal("remove of an unknown repo should report false")
	}
}

func TestRegistryReAddAfterRemoveGetsFreshStopChannel(t *testing.T) {
	r := newRegistry()
	old, _ := r.add("/a")
	r.remove("/a")

	fresh, added := r.add("/a")
	if !added {
		t.Fatal("re-adding a removed repo should report added")
	}
	if fresh == old {
		t.Fatal("re-added repo must get a new stop channel")
	}
	select {
	case <-fresh:
		t.Fatal("fresh stop channel must be open")
	default:
	}
}

func TestRegistryListReturnsCopy(t *testing.T) {
	r := newRegistry()
	r.add("/a")
	got := r.list()
	got[0] = "/mutated"
	if r.list()[0] != "/a" {
		t.Fatal("list() must not expose internal slice")
	}
}
