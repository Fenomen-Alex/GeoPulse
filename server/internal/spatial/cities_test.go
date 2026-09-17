package spatial

import (
	"strings"
	"testing"
)

func TestCityIndex_SearchByPrefix(t *testing.T) {
	idx := NewCityIndex()
	if len(idx.cities) == 0 {
		t.Fatal("expected embedded city index to be loaded")
	}

	places, err := idx.Search("berlin", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(places) == 0 {
		t.Fatal("expected a match for berlin")
	}
	if places[0].Label != "Berlin, Germany" {
		t.Errorf("expected Berlin, Germany as first match, got %q", places[0].Label)
	}
	if places[0].Lat == 0 && places[0].Lng == 0 {
		t.Error("expected coordinates for Berlin")
	}
}

func TestCityIndex_ExactRanking(t *testing.T) {
	idx := NewCityIndex()

	places, _ := idx.Search("washington", 3)
	if len(places) == 0 {
		t.Fatal("expected matches for washington")
	}
	// The exact-name city (Washington, D.C.) should outrank state towns with
	// the same prefix thanks to the exact-name weight.
	if !strings.Contains(places[0].Label, "Washington") {
		t.Errorf("expected a Washington match first, got %q", places[0].Label)
	}
}

func TestCityIndex_NoMatch(t *testing.T) {
	idx := NewCityIndex()
	places, err := idx.Search("xxqqzznonexistentplace", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(places) != 0 {
		t.Errorf("expected no matches, got %d", len(places))
	}
}

func TestCityIndex_ShortQuery(t *testing.T) {
	idx := NewCityIndex()
	places, err := idx.Search("  a ", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	// Single-char query is allowed by the local index (the handler enforces
	// the minimum), but it should return bounded, sane results.
	if len(places) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(places))
	}
}

func TestChainGeocoder_PrimaryHitSkipsFallback(t *testing.T) {
	primary := &stubGeocoderStatic{
		places: []Place{{Lat: 52.52, Lng: 13.405, Label: "Berlin, Germany"}},
	}
	fallback := &stubGeocoderStatic{err: errUnreachable}

	c := NewChainGeocoder(primary, fallback)
	places, err := c.Search("berlin", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(places) != 1 {
		t.Fatalf("expected 1 result, got %d", len(places))
	}
	if fallback.called {
		t.Error("fallback should not be called when primary has a match")
	}
}

func TestChainGeocoder_FallbackOnMiss(t *testing.T) {
	primary := &stubGeocoderStatic{places: nil}
	fallback := &stubGeocoderStatic{
		places: []Place{{Lat: 52.52, Lng: 13.405, Label: "Berlin, Germany"}},
	}

	c := NewChainGeocoder(primary, fallback)
	places, err := c.Search("berlin", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(places) != 1 {
		t.Fatalf("expected 1 result from fallback, got %d", len(places))
	}
	if !fallback.called {
		t.Error("expected fallback to be consulted on primary miss")
	}
}

var errUnreachable = &chainupError{}

type chainupError struct{}

func (e *chainupError) Error() string { return "upstream unavailable" }

type stubGeocoderStatic struct {
	places []Place
	err    error
	called bool
}

func (s *stubGeocoderStatic) Search(_ string, limit int) ([]Place, error) {
	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	if len(s.places) > limit {
		return s.places[:limit], nil
	}
	return s.places, nil
}

// TestParseCityIndex_CorruptData ensures a broken index degrades gracefully
// instead of crashing the server.
func TestParseCityIndex_CorruptData(t *testing.T) {
	if _, err := parseCityIndex([]byte("nope")); err == nil {
		t.Error("expected an error for corrupt index")
	}
	if _, err := parseCityIndex([]byte(cityIndexMagic + "notenoughbytes")); err == nil {
		t.Error("expected an error for truncated index")
	}
}

func BenchmarkCityIndex_Search(b *testing.B) {
	idx := NewCityIndex()
	queries := []string{"berlin", "washington", "new york", "paris", "tokyo"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := idx.Search(queries[i%len(queries)], 8); err != nil {
			b.Fatal(err)
		}
	}
}