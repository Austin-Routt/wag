// Package language implements the string-only template dialect shared by the
// catalogue, game cards, board renderer and their validation.
package language

import (
	"fmt"
	"strconv"
	"strings"
)

// Render supports %s, Go's one-based %[n]s indices, and literal %%. Indices
// may repeat, but every predicate argument must occur at least once. Values
// are inserted once and are never parsed as another format string.
func Render(tmpl string, args []string) (string, error) {
	return render(tmpl, args, len(args))
}

// Validate checks syntax, index bounds and argument coverage without rendering.
func Validate(tmpl string, arity int) error {
	_, err := render(tmpl, nil, arity)
	return err
}

// StartsWithArgument identifies a leading string placeholder, including an
// indexed one. A leading literal percent (%%) is not an argument.
func StartsWithArgument(tmpl string) bool {
	if strings.HasPrefix(tmpl, "%s") {
		return true
	}
	if !strings.HasPrefix(tmpl, "%[") {
		return false
	}
	i := strings.IndexByte(tmpl, ']')
	return i > 2 && i+1 < len(tmpl) && tmpl[i+1] == 's'
}

func render(tmpl string, args []string, arity int) (string, error) {
	if arity < 0 {
		return "", fmt.Errorf("negative template arity")
	}
	// Arity can come from a file's predicate/arity key. Keep validation
	// proportional to the template length, even for a malformed huge arity.
	used := make(map[int]bool)
	next := 0
	var out strings.Builder
	for i := 0; i < len(tmpl); {
		if tmpl[i] != '%' {
			out.WriteByte(tmpl[i])
			i++
			continue
		}
		start := i
		i++
		if i >= len(tmpl) {
			return "", fmt.Errorf("unterminated template escape at %d", start)
		}
		if tmpl[i] == '%' {
			out.WriteByte('%')
			i++
			continue
		}
		index := next
		if tmpl[i] == '[' {
			i++
			begin := i
			for i < len(tmpl) && tmpl[i] >= '0' && tmpl[i] <= '9' {
				i++
			}
			if begin == i || i >= len(tmpl) || tmpl[i] != ']' {
				return "", fmt.Errorf("invalid template index at %d", start)
			}
			n, err := strconv.Atoi(tmpl[begin:i])
			if err != nil || n < 1 || n > arity {
				return "", fmt.Errorf("template index outside 1..%d at %d", arity, start)
			}
			index = n - 1
			i++
		}
		if i >= len(tmpl) || tmpl[i] != 's' {
			return "", fmt.Errorf("unsupported template verb at %d; use %%s or %%[n]s", start)
		}
		if index < 0 || index >= arity {
			return "", fmt.Errorf("template argument outside 1..%d at %d", arity, start)
		}
		used[index] = true
		if args != nil {
			out.WriteString(args[index])
		}
		next = index + 1
		i++
	}
	if len(used) != arity {
		// If k indices occur, one of the first k+1 must be missing.
		for i := 0; i <= len(used); i++ {
			if !used[i] {
				return "", fmt.Errorf("template omits argument %d", i+1)
			}
		}
	}
	return out.String(), nil
}
