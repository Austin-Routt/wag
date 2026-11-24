package annotate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// checkedJSON preserves unambiguous field identity in dataset inputs and model
// replies. Reject duplicate names before map decoding can overwrite a value.
// json.Valid checks syntax and encoding/json's nesting bound before the scan.
// This is an ingestion check, not semantic verification of an annotation.
func checkedJSON(raw []byte, dst any) error {
	if !utf8.Valid(raw) {
		return fmt.Errorf("JSON is not valid UTF-8")
	}
	if !json.Valid(raw) {
		return fmt.Errorf("invalid JSON or nesting exceeds encoding/json's limit")
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
