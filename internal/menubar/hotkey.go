//go:build darwin || windows

package menubar

import (
	"fmt"
	"strings"
	"unicode"
)

// splitHotkey parses a shortcut spec like "cmd+opt+m" into its modifier tokens
// and a single key token. Tokens are trimmed and lowercased. The last token is
// the key; every token before it is a modifier. At least one modifier and a
// non-empty key are required. Mapping the tokens to platform key codes happens
// in the platform-specific backend.
func splitHotkey(spec string) (mods []string, key string, err error) {
	parts := strings.Split(spec, "+")
	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.ToLower(strings.TrimSpace(p))
		if t == "" {
			return nil, "", fmt.Errorf("invalid hotkey %q: empty token", spec)
		}
		tokens = append(tokens, t)
	}
	if len(tokens) < 2 {
		return nil, "", fmt.Errorf("invalid hotkey %q: need at least one modifier and a key", spec)
	}
	return tokens[:len(tokens)-1], tokens[len(tokens)-1], nil
}

// shortcutLabel renders a spec for the menu, capitalizing each token, e.g.
// "cmd+opt+m" -> "Shortcut: Cmd+Opt+M".
func shortcutLabel(spec string) string {
	tokens := strings.Split(spec, "+")
	for i, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		r := []rune(t)
		r[0] = unicode.ToUpper(r[0])
		tokens[i] = string(r)
	}
	return "Shortcut: " + strings.Join(tokens, "+")
}
