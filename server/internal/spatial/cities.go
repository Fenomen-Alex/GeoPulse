package spatial

import (
	_ "embed"

	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
)

//go:embed cities.bin
var cityIndexData []byte

const cityIndexMagic = "GPCTY01"

// cityEntry is one city loaded from the embedded binary index.
type cityEntry struct {
	name    string
	ccName  string
	country string
	lat     float64
	lng     float64
	pop     uint32
}

// CityIndex is an offline, embedded geocoder that resolves city-level queries
// from a GeoNames-derived index. It never touches the network.
type CityIndex struct {
	cities []cityEntry
}

// NewCityIndex loads the embedded index. If the index is missing or corrupt
// the returned index performs no matches (search degrades to the remote
// fallback) rather than failing to boot the server.
func NewCityIndex() *CityIndex {
	cities, err := parseCityIndex(cityIndexData)
	if err != nil {
		return &CityIndex{}
	}
	return &CityIndex{cities: cities}
}

func parseCityIndex(data []byte) ([]cityEntry, error) {
	if len(data) < len(cityIndexMagic)+4 {
		return nil, fmt.Errorf("city index too small")
	}
	if string(data[:len(cityIndexMagic)]) != cityIndexMagic {
		return nil, fmt.Errorf("bad city index magic")
	}
	count := int(binary.LittleEndian.Uint32(data[len(cityIndexMagic):]))
	off := len(cityIndexMagic) + 4

	cities := make([]cityEntry, 0, count)
	for i := 0; i < count; i++ {
		var c cityEntry
		var err error
		if c.name, off, err = readIndexString(data, off); err != nil {
			return nil, err
		}
		if c.ccName, off, err = readIndexString(data, off); err != nil {
			return nil, err
		}
		if c.country, off, err = readIndexString(data, off); err != nil {
			return nil, err
		}
		if off+12 > len(data) {
			return nil, fmt.Errorf("truncated city index")
		}
		c.lat = float64(int32(binary.LittleEndian.Uint32(data[off:]))) / 1e5
		c.lng = float64(int32(binary.LittleEndian.Uint32(data[off+4:]))) / 1e5
		c.pop = binary.LittleEndian.Uint32(data[off+8:])
		off += 12
		cities = append(cities, c)
	}
	return cities, nil
}

func readIndexString(data []byte, off int) (string, int, error) {
	if off+2 > len(data) {
		return "", off, fmt.Errorf("truncated city index string")
	}
	n := int(binary.LittleEndian.Uint16(data[off:]))
	off += 2
	if off+n > len(data) {
		return "", off, fmt.Errorf("truncated city index string body")
	}
	s := string(data[off : off+n])
	return s, off + n, nil
}

// Search returns up to limit city matches for the query. Exact-name matches
// rank first, then prefix matches weighted by population, then any substring
// matches. The index is small enough that a full pass is fast.
func (c *CityIndex) Search(query string, limit int) ([]Place, error) {
	q := normalizeQuery(query)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	// exact: cases like "washington -> Washington, D.C."
	// prefix: name starts with the query (bounded collection)
	// substring: fuzzy mid-name matches (only when few prefix hits)
	exact := make([]cityEntry, 0, 4)
	prefix := make([]cityEntry, 0, 200)
	sub := make([]cityEntry, 0, 8)

	for _, e := range c.cities {
		switch {
		case e.name == q:
			exact = append(exact, e)
		case strings.HasPrefix(e.name, q):
			if len(prefix) < 200 {
				prefix = append(prefix, e)
			}
		case len(exact)+len(prefix) < limit && strings.Contains(e.name, q):
			sub = append(sub, e)
		}
	}

	if len(exact) > 1 {
		sort.SliceStable(exact, func(i, j int) bool { return exact[i].pop > exact[j].pop })
	}
	if len(prefix) > 1 {
		sort.SliceStable(prefix, func(i, j int) bool {
			si := matchScore(prefix[i].name, q, prefix[i].pop)
			sj := matchScore(prefix[j].name, q, prefix[j].pop)
			if si == sj {
				return prefix[i].name < prefix[j].name
			}
			return si > sj
		})
	}
	if len(sub) > 1 {
		sort.SliceStable(sub, func(i, j int) bool {
			return sub[i].pop > sub[j].pop
		})
	}

	results := make([]Place, 0, limit)
	seen := map[string]bool{}
	add := func(e cityEntry) {
		if len(results) >= limit {
			return
		}
		key := e.name + "|" + e.country
		if seen[key] {
			return
		}
		seen[key] = true
		results = append(results, Place{Lat: e.lat, Lng: e.lng, Label: labelFor(e)})
	}
	for _, e := range exact {
		add(e)
	}
	for _, e := range prefix {
		add(e)
	}
	for _, e := range sub {
		add(e)
	}
	return results, nil
}

func matchScore(name, q string, pop uint32) int {
	switch {
	case name == q:
		return 1_000_000
	case strings.HasPrefix(name, q):
		return 100_000 + int(math.Log2(float64(pop+1)))*10
	default:
		return int(math.Log2(float64(pop+1))) * 10
	}
}

func labelFor(e cityEntry) string {
	title := e.name
	if len(title) > 0 {
		title = string(unicode.ToUpper(rune(title[0]))) + title[1:]
	}
	if e.ccName != "" {
		return title + ", " + e.ccName
	}
	return title
}

func normalizeQuery(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	return strings.Map(func(r rune) rune {
		if r < 32 || r > 0x7f {
			return -1
		}
		return r
	}, q)
}

// ChainGeocoder tries a primary offline geocoder first and falls back to a
// remote service (Nominatim) only when the local index has no matches, cutting
// upstream dependence for the common city-search case.
type ChainGeocoder struct {
	primary  Geocoder
	fallback Geocoder
}

func NewChainGeocoder(primary, fallback Geocoder) *ChainGeocoder {
	return &ChainGeocoder{primary: primary, fallback: fallback}
}

func (c *ChainGeocoder) Search(query string, limit int) ([]Place, error) {
	places, err := c.primary.Search(query, limit)
	if err == nil && len(places) > 0 {
		return places, nil
	}
	if c.fallback == nil {
		return places, err
	}
	return c.fallback.Search(query, limit)
}

var _ Geocoder = (*CityIndex)(nil)
var _ Geocoder = (*ChainGeocoder)(nil)