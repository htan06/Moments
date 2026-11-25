// Package content implements the Content value object: an immutable AST
// whose root is a `paragraph` node and whose children are `text`,
// `mention` or `hashtag` leaves. Only `paragraph` has children.
//
// The zero Content is null (a Post's content may be null).
package content

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"

	"github.com/htan06/Moments/post/internal/domain"
)

type NodeType string

const (
	Paragraph NodeType = "paragraph"
	Text      NodeType = "text"
	Mention   NodeType = "mention"
	Hashtag   NodeType = "hashtag"
)

const (
	AttrUserID    = "userId"
	AttrHashtagID = "hashtagId"
)

func (t NodeType) Valid() bool {
	switch t {
	case Paragraph, Text, Mention, Hashtag:
		return true
	}
	return false
}

func (t NodeType) String() string { return string(t) }

// HashtagName is a normalized tag name: lowercase, without '#'.
type HashtagName string

func NewHashtagName(raw string) (HashtagName, error) {
	s := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(raw), "#"))
	if s == "" {
		return "", domain.ErrNodeInvalid
	}
	for _, r := range s {
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", domain.ErrNodeInvalid
		}
	}
	return HashtagName(s), nil
}

func (h HashtagName) String() string { return string(h) }

// Node is a leaf of the Content tree.
type Node struct {
	typ  NodeType
	text string
	// refID is the userId of a mention or the hashtagId of a hashtag.
	refID int64
}

// NewTextNode keeps the text as is: whitespace-only text (e.g. " "
// between a mention and a hashtag) is valid, only empty text is not.
func NewTextNode(text string) (Node, error) {
	if text == "" {
		return Node{}, domain.ErrNodeInvalid
	}
	return Node{typ: Text, text: text}, nil
}

func NewMentionNode(text string, userID int64) (Node, error) {
	if strings.TrimSpace(text) == "" || userID <= 0 {
		return Node{}, domain.ErrNodeInvalid
	}
	return Node{typ: Mention, text: text, refID: userID}, nil
}

func NewHashtagNode(text string, hashtagID int64) (Node, error) {
	if hashtagID <= 0 {
		return Node{}, domain.ErrNodeInvalid
	}
	if _, err := NewHashtagName(text); err != nil {
		return Node{}, err
	}
	return Node{typ: Hashtag, text: text, refID: hashtagID}, nil
}

func (n Node) Type() NodeType { return n.typ }
func (n Node) Text() string   { return n.text }

// Attributes returns the node attributes as defined in post.md:
// `mention` has `userId`, `hashtag` has `hashtagId`.
func (n Node) Attributes() map[string]string {
	switch n.typ {
	case Mention:
		return map[string]string{AttrUserID: strconv.FormatInt(n.refID, 10)}
	case Hashtag:
		return map[string]string{AttrHashtagID: strconv.FormatInt(n.refID, 10)}
	}
	return nil
}

type Content struct {
	children []Node
}

// New builds a non-null Content from its leaves.
func New(children ...Node) (Content, error) {
	if len(children) == 0 {
		return Content{}, domain.ErrContentInvalid
	}
	for _, c := range children {
		if c.typ != Text && c.typ != Mention && c.typ != Hashtag {
			return Content{}, domain.ErrNodeInvalid
		}
	}
	return Content{children: append([]Node(nil), children...)}, nil
}

func (c Content) IsNull() bool { return len(c.children) == 0 }

func (c Content) Children() []Node { return append([]Node(nil), c.children...) }

func (c Content) HasHashtag() bool {
	for _, n := range c.children {
		if n.typ == Hashtag {
			return true
		}
	}
	return false
}

// MentionedUserIDs returns distinct mentioned user ids, in order.
func (c Content) MentionedUserIDs() []int64 { return c.refIDs(Mention) }

// HashtagIDs returns distinct hashtag ids, in order.
func (c Content) HashtagIDs() []int64 { return c.refIDs(Hashtag) }

func (c Content) refIDs(t NodeType) []int64 {
	seen := map[int64]struct{}{}
	out := []int64{}
	for _, n := range c.children {
		if n.typ != t {
			continue
		}
		if _, ok := seen[n.refID]; ok {
			continue
		}
		seen[n.refID] = struct{}{}
		out = append(out, n.refID)
	}
	return out
}

// JSON is the canonical representation of Content (post.md §2).

type nodeJSON struct {
	Type       string            `json:"type"`
	Text       string            `json:"text,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Children   []nodeJSON        `json:"children,omitempty"`
}

func (c Content) MarshalJSON() ([]byte, error) {
	if c.IsNull() {
		return []byte("null"), nil
	}
	root := nodeJSON{Type: Paragraph.String(), Children: make([]nodeJSON, len(c.children))}
	for i, n := range c.children {
		root.Children[i] = nodeJSON{Type: n.typ.String(), Text: n.text, Attributes: n.Attributes()}
	}
	return json.Marshal(root)
}

func (c *Content) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*c = Content{}
		return nil
	}
	var root nodeJSON
	if err := json.Unmarshal(data, &root); err != nil {
		return domain.ErrContentInvalid
	}
	if root.Type != Paragraph.String() || root.Text != "" || len(root.Attributes) > 0 {
		return domain.ErrContentInvalid
	}
	children := make([]Node, 0, len(root.Children))
	for _, child := range root.Children {
		n, err := nodeFromJSON(child)
		if err != nil {
			return err
		}
		children = append(children, n)
	}
	parsed, err := New(children...)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

func nodeFromJSON(j nodeJSON) (Node, error) {
	if len(j.Children) > 0 {
		return Node{}, domain.ErrNodeInvalid
	}
	switch NodeType(j.Type) {
	case Text:
		return NewTextNode(j.Text)
	case Mention:
		id, err := strconv.ParseInt(j.Attributes[AttrUserID], 10, 64)
		if err != nil {
			return Node{}, domain.ErrNodeInvalid
		}
		return NewMentionNode(j.Text, id)
	case Hashtag:
		id, err := strconv.ParseInt(j.Attributes[AttrHashtagID], 10, 64)
		if err != nil {
			return Node{}, domain.ErrNodeInvalid
		}
		return NewHashtagNode(j.Text, id)
	}
	return Node{}, domain.ErrNodeInvalid
}
