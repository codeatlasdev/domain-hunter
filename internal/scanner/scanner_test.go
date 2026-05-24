package scanner

import (
	"testing"
)

func TestGenerateCVC3(t *testing.T) {
	names := Generate(3, "CVC")
	if len(names) == 0 {
		t.Fatal("expected CVC names, got 0")
	}
	// 19 consonants * 5 vowels * 19 consonants = 1805
	expected := 19 * 5 * 19
	if len(names) != expected {
		t.Errorf("expected %d CVC names, got %d", expected, len(names))
	}
	// Verify pattern: consonant-vowel-consonant
	for _, name := range names[:10] {
		if len(name) != 3 {
			t.Errorf("expected 3 chars, got %d: %s", len(name), name)
		}
		if isVowel(name[0]) {
			t.Errorf("first char should be consonant: %s", name)
		}
		if !isVowel(name[1]) {
			t.Errorf("second char should be vowel: %s", name)
		}
		if isVowel(name[2]) {
			t.Errorf("third char should be consonant: %s", name)
		}
	}
}

func TestGenerateVCV3(t *testing.T) {
	names := Generate(3, "VCV")
	// 5 vowels * 19 consonants * 5 vowels = 475
	expected := 5 * 19 * 5
	if len(names) != expected {
		t.Errorf("expected %d VCV names, got %d", expected, len(names))
	}
}

func TestGenerateCVCV4(t *testing.T) {
	names := Generate(4, "CVCV")
	// 19 * 5 * 19 * 5 = 9025
	expected := 19 * 5 * 19 * 5
	if len(names) != expected {
		t.Errorf("expected %d CVCV names, got %d", expected, len(names))
	}
	// Verify pattern
	for _, name := range names[:10] {
		if len(name) != 4 {
			t.Errorf("expected 4 chars, got %d: %s", len(name), name)
		}
		if isVowel(name[0]) || !isVowel(name[1]) || isVowel(name[2]) || !isVowel(name[3]) {
			t.Errorf("invalid CVCV pattern: %s", name)
		}
	}
}

func TestGenerateAll3(t *testing.T) {
	names := Generate(3, "ALL")
	cvc := 19 * 5 * 19  // 1805
	vcv := 5 * 19 * 5   // 475
	expected := cvc + vcv // 2280
	if len(names) != expected {
		t.Errorf("expected %d ALL names, got %d", expected, len(names))
	}
}

func TestGenerateAll4(t *testing.T) {
	names := Generate(4, "ALL")
	// Only CVCV for 4 letters
	expected := 19 * 5 * 19 * 5
	if len(names) != expected {
		t.Errorf("expected %d names, got %d", expected, len(names))
	}
}

func TestGenerateDomains(t *testing.T) {
	domains := GenerateDomains(3, PatternCVC, []string{"com", "dev"})
	cvc := 19 * 5 * 19
	expected := cvc * 2 // 2 TLDs
	if len(domains) != expected {
		t.Errorf("expected %d domains, got %d", expected, len(domains))
	}
	// Check format
	for _, d := range domains[:5] {
		if !contains(d, ".com") && !contains(d, ".dev") {
			t.Errorf("domain missing TLD: %s", d)
		}
	}
}

func TestGenerateAllDomains(t *testing.T) {
	domains := GenerateAll([]string{"com"}, 3, "CVC")
	expected := 19 * 5 * 19
	if len(domains) != expected {
		t.Errorf("expected %d, got %d", expected, len(domains))
	}
}

func TestProviders(t *testing.T) {
	for _, tld := range SupportedTLDs {
		if _, ok := Providers[tld]; !ok {
			t.Errorf("missing provider for TLD: %s", tld)
		}
	}
}

func TestProvidersURLFormat(t *testing.T) {
	for tld, url := range Providers {
		if !contains(url, "https://") {
			t.Errorf("provider %s URL not HTTPS: %s", tld, url)
		}
		if !contains(url, "domain/") {
			t.Errorf("provider %s URL missing /domain/ path: %s", tld, url)
		}
	}
}

func isVowel(b byte) bool {
	return b == 'a' || b == 'e' || b == 'i' || b == 'o' || b == 'u'
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
