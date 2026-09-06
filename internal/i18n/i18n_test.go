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
