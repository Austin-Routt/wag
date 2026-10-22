package catalogue

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

// Corpus is the book's text, normalized, for the copy check: our own
// writing must not reproduce the book's sentences (plan §1). The text stays
// outside the repository; the checker is given it (--book).
type Corpus struct {
	sentences [][]string
	grams     map[int]map[string]bool
}

var wordRE = regexp.MustCompile(`[a-z0-9']+`)
var varRE = regexp.MustCompile(`^[a-z][0-9]*$`)

// words normalizes a text to words: lower case, and every one-letter word or
// letter with digits (the schemes' variables: E, A, S1, but also "a") as
// "x", so that a sentence copied with its variables renamed still matches.
func words(s string) []string {
	ws := wordRE.FindAllString(strings.ToLower(strings.NewReplacer("’", "'", "{", " ", "}", " ").Replace(s)), -1)
	for i, w := range ws {
		if varRE.MatchString(w) {
			ws[i] = "x"
		}
	}
	return ws
}

// NewCorpus builds a corpus from texts (one sentence or line each).
func NewCorpus(texts []string) *Corpus {
	c := &Corpus{grams: map[int]map[string]bool{}}
	for _, t := range texts {
		for _, line := range strings.Split(t, "\n") {
			ws := words(line)
			if len(ws) > 0 {
				c.sentences = append(c.sentences, ws)
			}
		}
	}
	return c
}

// LoadCorpus reads book texts: a JSON file contributes the strings under
// its "text" and "book_text" keys (docs/book-catalogue/catalogue.json);
// any other file, its lines.
func LoadCorpus(paths []string) (*Corpus, error) {
	var texts []string
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(strings.ToLower(p), ".json") {
			var v any
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			collectText(v, &texts)
			continue
		}
		texts = append(texts, string(b))
	}
	return NewCorpus(texts), nil
}

func collectText(v any, out *[]string) {
	switch x := v.(type) {
	case map[string]any:
		for k, y := range x {
			if s, ok := y.(string); ok && (k == "text" || k == "book_text") {
				*out = append(*out, s)
				continue
			}
			collectText(y, out)
		}
	case []any:
		for _, y := range x {
			collectText(y, out)
		}
	}
}

// Sentences is how many book sentences the corpus holds.
func (c *Corpus) Sentences() int { return len(c.sentences) }

func (c *Corpus) gramSet(n int) map[string]bool {
	if s, ok := c.grams[n]; ok {
		return s
	}
	s := map[string]bool{}
	for _, ws := range c.sentences {
		for i := 0; i+n <= len(ws); i++ {
			s[strings.Join(ws[i:i+n], " ")] = true
		}
	}
	c.grams[n] = s
	return s
}

// Shares returns a run of n words that the text shares with a book
// sentence (normalized), or "".
func (c *Corpus) Shares(text string, n int) string {
	ws := words(text)
	set := c.gramSet(n)
	for i := 0; i+n <= len(ws); i++ {
		if g := strings.Join(ws[i:i+n], " "); set[g] {
			return g
		}
	}
	return ""
}

// Closest returns, for a text of at most maxWords words, its similarity to
// the closest book sentence (2·LCS/(m+n) over normalized words) and that
// sentence; for a longer text, 0.
func (c *Corpus) Closest(text string, maxWords int) (float64, string) {
	ws := words(text)
	if len(ws) == 0 || len(ws) > maxWords {
		return 0, ""
	}
	best, sent := 0.0, ""
	for _, s := range c.sentences {
		if len(s) > 2*len(ws)+4 || 2*len(s)+4 < len(ws) {
			continue
		}
		if r := similarity(ws, s); r > best {
			best, sent = r, strings.Join(s, " ")
		}
	}
	return best, sent
}

// similarity is 2·LCS(a,b)/(len(a)+len(b)) over words.
func similarity(a, b []string) float64 {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			switch {
			case a[i-1] == b[j-1]:
				cur[j] = prev[j-1] + 1
			case prev[j] >= cur[j-1]:
				cur[j] = prev[j]
			default:
				cur[j] = cur[j-1]
			}
		}
		prev, cur = cur, prev
	}
	return 2 * float64(prev[len(b)]) / float64(len(a)+len(b))
}
