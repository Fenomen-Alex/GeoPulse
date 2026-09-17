// Command gencities builds the compact offline city-search index (cities.bin)
// embedded by server/internal/spatial.
//
// It reads a GeoNames cities dump (TSV, e.g. cities15000.txt) plus the
// countryInfo.txt name table, and writes a normalized, deduplicated binary
// index to server/internal/spatial/cities.bin. Run:
//
//	go run ./cmd/gencities -cities cities15000.txt -countries countryInfo.txt -out ../internal/spatial/cities.bin
//
// Data is from GeoNames (https://geonames.org), CC-BY 4.0. Attribute
// "GeoNames" in any app that ships the generated index.
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const magic = "GPCTY01"

// city holds one normalized entry written into the binary index.
type city struct {
	name    string // lowercased asciiname
	country string // ISO-3166-1 alpha-2 code
	ccName  string // English country name (label)
	lat     int32  // scaled by 1e5
	lng     int32  // scaled by 1e5
	pop     uint32
}

func main() {
	citiesPath := flag.String("cities", "", "path to GeoNames cities TSV (cities15000.txt)")
	countriesPath := flag.String("countries", "", "path to GeoNames countryInfo.txt")
	outPath := flag.String("out", "", "output path for cities.bin")
	flag.Parse()

	if *citiesPath == "" || *countriesPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "missing required flag(s); see -h")
		os.Exit(2)
	}

	ccName, err := readCountryNames(*countriesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "countries: %v\n", err)
		os.Exit(1)
	}

	cities, err := readCities(*citiesPath, ccName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cities: %v\n", err)
		os.Exit(1)
	}

	if err := writeIndex(*outPath, cities); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d cities to %s\n", len(cities), *outPath)
}

// readCountryNames maps ISO-3166-1 alpha-2 codes to English country names.
func readCountryNames(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	names := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}
		names[strings.ToUpper(fields[0])] = strings.TrimSpace(fields[4])
	}
	return names, sc.Err()
}

// readCities parses the GeoNames dump, keeping populated P-feature places and
// deduplicating by lowercased name + country (highest population wins).
func readCities(path string, ccName map[string]string) ([]city, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	seen := map[string]int{} // "name|cc" -> index in out
	out := []city{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 256*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		// geonameid name asciiname alternatnames lat lng featclass featcode cc ...
		if len(fields) < 15 {
			continue
		}
		featClass := fields[6]
		if featClass != "P" {
			continue
		}
		cc := strings.ToUpper(fields[8])
		name := strings.TrimSpace(fields[2]) // asciiname
		if name == "" {
			continue
		}
		var lat, lng float64
		var pop int64
		fmt.Sscanf(fields[4], "%f", &lat)
		fmt.Sscanf(fields[5], "%f", &lng)
		fmt.Sscanf(fields[14], "%d", &pop)
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			continue
		}
		key := strings.ToLower(name) + "|" + cc
		if idx, ok := seen[key]; ok {
			if pop > int64(out[idx].pop) {
				out[idx].pop = uint32(pop)
			}
			continue
		}
		cn := ccName[cc]
		seen[key] = len(out)
		out = append(out, city{
			name:    strings.ToLower(name),
			country: cc,
			ccName:  cn,
			lat:     int32(lat * 1e5),
			lng:     int32(lng * 1e5),
			pop:     uint32(pop),
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, nil
}

// writeIndex serializes cities to the versioned binary format:
//
//	magic[7] | uint32 count | cities...
//	city = uint16 nameLen | name | uint16 ccNameLen | ccName | byte ccLen | cc | int32 lat | int32 lng | uint32 pop
func writeIndex(path string, cities []city) error {
	var buf []byte
	buf = append(buf, magic...)
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], uint32(len(cities)))
	buf = append(buf, tmp[:]...)

	for _, c := range cities {
		buf = appendString(buf, c.name)
		buf = appendString(buf, c.ccName)
		buf = appendString(buf, c.country)

		binary.LittleEndian.PutUint32(tmp[:], uint32(c.lat))
		buf = append(buf, tmp[:]...)
		binary.LittleEndian.PutUint32(tmp[:], uint32(c.lng))
		buf = append(buf, tmp[:]...)
		binary.LittleEndian.PutUint32(tmp[:], c.pop)
		buf = append(buf, tmp[:]...)
	}

	if err := os.WriteFile(path, buf, 0o644); err != nil {
		return err
	}
	return nil
}

func appendString(buf []byte, s string) []byte {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], uint16(len(s)))
	buf = append(buf, tmp[:]...)
	return append(buf, s...)
}