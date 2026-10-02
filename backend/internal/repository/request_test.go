package repository

import "testing"

func TestEscapeLikePatternTreatsWildcardsAndBackslashLiterally(t *testing.T) {
	got := escapeLikePattern(`50%_d'água\final`)
	want := `50\%\_d'água\\final`
	if got != want {
		t.Fatalf("escapeLikePattern() = %q, want %q", got, want)
	}
}

func TestRequestFilterArgsWrapsEscapedTitleAsSubstring(t *testing.T) {
	args := requestFilterArgs(RequestFilters{TitleQuery: `á_%`})
	if got, ok := args[4].(string); !ok || got != `%á\_\%%` {
		t.Fatalf("title pattern = %#v", args[4])
	}
}
