package scanner

import "fmt"

type Pattern string

const (
	PatternCVC   Pattern = "CVC"
	PatternVCV   Pattern = "VCV"
	PatternCVCV  Pattern = "CVCV"
	PatternCVCVC Pattern = "CVCVC"
	PatternAll   Pattern = "ALL"
)

var (
	vowels     = []byte("aeiou")
	consonants = []byte("bcdfghjklmnprstvwxz") // no q, y — better pronounceability
)

// GenerateDomains produces pronounceable domain names for the given config.
func GenerateDomains(length int, pattern Pattern, tlds []string) []string {
	names := Generate(length, string(pattern))
	domains := make([]string, 0, len(names)*len(tlds))
	for _, tld := range tlds {
		for _, name := range names {
			domains = append(domains, fmt.Sprintf("%s.%s", name, tld))
		}
	}
	return domains
}

// Generate produces name strings (without TLD) for given length and pattern.
func Generate(length int, pattern string) []string {
	p := Pattern(pattern)
	return generateNames(length, p)
}

func generateNames(length int, pattern Pattern) []string {
	switch length {
	case 3:
		return gen3(pattern)
	case 4:
		return gen4(pattern)
	case 5:
		return gen5(pattern)
	}
	return nil
}

func gen3(pattern Pattern) []string {
	var out []string
	if pattern == PatternCVC || pattern == PatternAll {
		for _, c1 := range consonants {
			for _, v := range vowels {
				for _, c2 := range consonants {
					out = append(out, string([]byte{c1, v, c2}))
				}
			}
		}
	}
	if pattern == PatternVCV || pattern == PatternAll {
		for _, v1 := range vowels {
			for _, c := range consonants {
				for _, v2 := range vowels {
					out = append(out, string([]byte{v1, c, v2}))
				}
			}
		}
	}
	return out
}

func gen4(pattern Pattern) []string {
	var out []string
	if pattern == PatternCVCV || pattern == PatternAll {
		for _, c1 := range consonants {
			for _, v1 := range vowels {
				for _, c2 := range consonants {
					for _, v2 := range vowels {
						out = append(out, string([]byte{c1, v1, c2, v2}))
					}
				}
			}
		}
	}
	return out
}

func gen5(pattern Pattern) []string {
	var out []string
	if pattern == PatternCVCVC || pattern == PatternAll {
		for _, c1 := range consonants {
			for _, v1 := range vowels {
				for _, c2 := range consonants {
					for _, v2 := range vowels {
						for _, c3 := range consonants {
							out = append(out, string([]byte{c1, v1, c2, v2, c3}))
						}
					}
				}
			}
		}
	}
	return out
}
