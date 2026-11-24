package annotate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

// checkedJSON preserves unambiguous field identity in dataset inputs and model
// replies. Reject duplicate names before map decoding can overwrite a value.
// Decoding into a json.RawMessage first checks the syntax and encoding/json's
// nesting bound (10000) before the scan below, whose recursion has none of
// its own, and keeps the decoder's message ("invalid character '}' ..."),
// which is what a model is told when its reply is retried. This is an
// ingestion check, not semantic verification of an annotation.
func checkedJSON(raw []byte, dst any) error {
	if !utf8.Valid(raw) {
		return fmt.Errorf("JSON is not valid UTF-8")
	}
	if err := json.Unmarshal(raw, new(json.RawMessage)); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var value func() error
	value = func() error {
		token, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				token, err := d.Token()
				if err != nil {
					return err
				}
				key, ok := token.(string)
				if !ok {
					return fmt.Errorf("object member is not a string")
				}
				if seen[key] {
					return fmt.Errorf("duplicate JSON member %q", key)
				}
				seen[key] = true
				if err := value(); err != nil {
					return err
				}
			}
			_, err := d.Token()
			return err
		case '[':
			for d.More() {
				if err := value(); err != nil {
					return err
				}
			}
			_, err := d.Token()
			return err
		}
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	if err := value(); err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

// jsonFields lists the JSON member names of a struct's fields.
func jsonFields(v any) []string {
	var out []string
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" || !f.IsExported() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		out = append(out, name)
	}
	return out
}

// canonicalNames refuses an object with two members that spell one of the
// fields. encoding/json matches a struct's fields case-insensitively, so
// {"gold": "a.json", "Gold": ""} would decode as an empty gold: the last
// spelling wins, silently. One member is unambiguous however it is spelt;
// members that are no field's are left alone.
func canonicalNames(raw json.RawMessage, fields []string) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("not a JSON object: %v", err)
	}
	for _, f := range fields {
		var spelt []string
		for k := range m {
			if strings.EqualFold(k, f) {
				spelt = append(spelt, k)
			}
		}
		if len(spelt) < 2 {
			continue
		}
		sort.Strings(spelt)
		base := spelt[0]
		if slices.Contains(spelt, f) { // the field's own spelling, if given, is the base
			base = f
		}
		other := spelt[0]
		if other == base {
			other = spelt[1]
		}
		return fmt.Errorf("member %q is another spelling of %q", other, base)
	}
	return nil
}

// canonicalEach applies canonicalNames to every object of a JSON array;
// other elements are left to the caller.
func canonicalEach(raw json.RawMessage, fields []string) error {
	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		return err
	}
	for i, el := range list {
		if t := bytes.TrimSpace(el); len(t) > 0 && t[0] == '{' {
			if err := canonicalNames(el, fields); err != nil {
				return fmt.Errorf("entry %d: %v", i+1, err)
			}
		}
	}
	return nil
}
