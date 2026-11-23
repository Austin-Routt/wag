package evidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Decode rejects malformed, duplicate-key, unknown-field and trailing JSON.
// Canonical field spelling is enforced by comparing the generic input's keys
// with the decoded model's keys, without folding user-supplied map keys.
func Decode(raw []byte) (*Bundle, error) {
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("evidence JSON is not valid UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := uniqueValue(d); err != nil {
		return nil, err
	}
	if err := unicodeEscapes(raw); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing JSON after evidence bundle")
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	var b Bundle
	if err := d.Decode(&b); err != nil {
		return nil, err
	}
	var original, canonical interface{}
	if err := json.Unmarshal(raw, &original); err != nil {
		return nil, err
	}
	encoded, _ := json.Marshal(b)
	_ = json.Unmarshal(encoded, &canonical)
	if err := canonicalKeys(original, canonical, "$", false); err != nil {
		return nil, err
	}
	return &b, nil
}

func uniqueValue(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	switch t {
	case json.Delim('{'):
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not text")
			}
			if seen[s] {
				return fmt.Errorf("duplicate JSON key %q", s)
			}
			seen[s] = true
			if err := uniqueValue(d); err != nil {
				return err
			}
		}
		_, err = d.Token()
		return err
	case json.Delim('['):
		for d.More() {
			if err := uniqueValue(d); err != nil {
				return err
			}
		}
		_, err = d.Token()
		return err
	}
	return nil
}

func canonicalKeys(a, b interface{}, path string, freeMap bool) error {
	if a == nil {
		return fmt.Errorf("%s: null is not a record value; use explicit unknown statuses or empty lists", path)
	}
	switch x := a.(type) {
	case map[string]interface{}:
		y, ok := b.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s: expected a record", path)
		}
		var required []string
		if strings.HasSuffix(path, ".payload") {
			required = []string{"path", "sha256", "bytes", "media_type"}
		}
		if strings.HasSuffix(path, ".selector") {
			required = []string{"kind"}
			switch x["kind"] {
			case "text":
				required = append(required, "start", "end", "quote")
			case "page":
				required = append(required, "page")
			case "time":
				required = append(required, "start_ms", "end_ms")
			case "region":
				required = append(required, "page", "x", "y", "width", "height")
			}
		}
		for _, key := range required {
			if _, present := x[key]; !present {
				return fmt.Errorf("%s: missing required field %q", path, key)
			}
		}
		for key, v := range x {
			other, ok := y[key]
			if !ok && !freeMap {
				return fmt.Errorf("%s: noncanonical field %q", path, key)
			}
			if err := canonicalKeys(v, other, path+"."+key, key == "dimensions"); err != nil {
				return err
			}
		}
	case []interface{}:
		y, ok := b.([]interface{})
		if !ok && len(x) != 0 {
			return fmt.Errorf("%s: expected list", path)
		}
		for i, v := range x {
			if err := canonicalKeys(v, y[i], fmt.Sprintf("%s[%d]", path, i), false); err != nil {
				return err
			}
		}
	case string:
		if strings.IndexByte(x, 0) >= 0 {
			return fmt.Errorf("%s: NUL in text", path)
		}
	}
	return nil
}

// encoding/json accepts isolated UTF-16 surrogates by replacing them with
// U+FFFD. Evidence text must not be silently changed. The JSON token pass has
// already checked syntax, so this pass only enforces paired Unicode escapes.
func unicodeEscapes(raw []byte) error {
	inside := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			inside = !inside
			continue
		}
		if !inside || raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) || raw[i] != 'u' {
			continue
		}
		if i+4 >= len(raw) {
			return fmt.Errorf("truncated Unicode escape")
		}
		n, err := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if err != nil {
			return err
		}
		i += 4
		if n >= 0xDC00 && n <= 0xDFFF {
			return fmt.Errorf("unpaired low Unicode surrogate")
		}
		if n >= 0xD800 && n <= 0xDBFF {
			if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
				return fmt.Errorf("unpaired high Unicode surrogate")
			}
			low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
			if err != nil || low < 0xDC00 || low > 0xDFFF {
				return fmt.Errorf("unpaired high Unicode surrogate")
			}
			i += 6
		}
	}
	return nil
}
