package spatial

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// countingGeocoder records how many times it is invoked, so tests can assert
// that the cache absorbed repeat queries.
type countingGeocoder struct {
	calls    int
	fallback bool
}

func (c *countingGeocoder) Search(query string, limit int) ([]Place, error) {
	c.calls++
	if c.fallback {
		return nil, nil
	}
	return []Place{{Lat: 52.52, Lng: 13.405, Label: "Berlin, Germany"}}, nil
}

func TestCachingGeocoder_HitsAvoidUpstream(t *testing.T) {
	inner := &countingGeocoder{}
	c := NewCachingGeocoder(inner, 64, time.Hour)

	if _, err := c.Search("berlin", 5); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Search("berlin", 5); err != nil {
		t.Fatal(err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 upstream call after duplicates, got %d", inner.calls)
	}
}

func TestCachingGeocoder_DistinctQueriesBypass(t *testing.T) {
	inner := &countingGeocoder{}
	c := NewCachingGeocoder(inner, 64, time.Hour)

	for _, q := range []string{"berlin", "paris", "tokyo"} {
		if _, err := c.Search(q, 5); err != nil {
			t.Fatal(err)
		}
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 upstream calls for distinct queries, got %d", inner.calls)
	}
}

func TestCachingGeocoder_CachesEmptyResults(t *testing.T) {
	inner := &countingGeocoder{fallback: true}
	c := NewCachingGeocoder(inner, 64, time.Hour)

	if _, err := c.Search("xyz-not-a-place", 5); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Search("xyz-not-a-place", 5); err != nil {
		t.Fatal(err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected empty results to be cached (1 call), got %d", inner.calls)
	}
}

func TestCachingGeocoder_EvictsOldest(t *testing.T) {
	inner := &countingGeocoder{}
	c := NewCachingGeocoder(inner, 2, time.Hour)

	for _, q := range []string{"a", "b", "c", "a"} {
		if _, err := c.Search(q, 5); err != nil {
			t.Fatal(err)
		}
	}
	// a was evicted when c was inserted, so the final "a" hits upstream again.
	if inner.calls != 4 {
		t.Fatalf("expected LRU eviction (4 calls), got %d", inner.calls)
	}
}

func TestCachingGeocoder_ErrorNotCached(t *testing.T) {
	inner := &countingGeocoder{}
	c := NewCachingGeocoder(inner, 64, time.Hour)

	nerr := &failOnceGeocoder{c: inner}
	c.inner = nerr

	if _, err := c.Search("boom", 5); err == nil {
		t.Fatal("expected error from upstream")
	}
	if _, err := c.Search("boom", 5); err == nil {
		t.Fatal("errors must not be cached, expected second call")
	}
	if nerr.calls != 2 {
		t.Fatalf("expected 2 upstream calls for failing service, got %d", nerr.calls)
	}
}

type failOnceGeocoder struct {
	c     *countingGeocoder
	calls int
}

func (f *failOnceGeocoder) Search(q string, limit int) ([]Place, error) {
	f.calls++
	f.c.calls++
	return nil, errors.New(fmt.Sprintf("upstream error %d", f.calls))
}