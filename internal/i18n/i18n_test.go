package i18n

import "testing"

func TestDetect(t *testing.T) {
	tests := map[string]Language{
		"pt_BR.UTF-8": PortugueseBrazil,
		"pt-PT":       PortugueseBrazil,
		"en_US.UTF-8": English,
		"":            English,
	}
	for input, want := range tests {
		if got := Detect(input); got != want {
			t.Errorf("Detect(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestParse(t *testing.T) {
	if language, ok := Parse("pt-BR"); !ok || language != PortugueseBrazil {
		t.Fatalf("Parse(pt-BR) = %q, %t", language, ok)
	}
	if _, ok := Parse("fr"); ok {
		t.Fatal("fr should not be accepted")
	}
}

func TestText(t *testing.T) {
	if got := Text(PortugueseBrazil, KeyAboutCoffee); got == "" || got == Text(English, KeyAboutCoffee) {
		t.Errorf("expected distinct Portuguese translation for KeyAboutCoffee, got %q", got)
	}
	if got := Text(English, KeyAboutCoffee); got == "" {
		t.Errorf("expected English translation for KeyAboutCoffee, got empty")
	}
	if got := Text(PortugueseBrazil, KeyAbout); got != "Sobre" {
		t.Errorf("expected 'Sobre', got %q", got)
	}
	if got := Text(English, KeyAbout); got != "About" {
		t.Errorf("expected 'About', got %q", got)
	}
}
