package vo

type NodeType string

const (
	Text    NodeType = "TEXT"
	Mention NodeType = "MENTION"
	Hashtag NodeType = "HASHTAG"
)

type Content struct {
	Type NodeType `json:"type"`
	Text string   `json:"text"`
}
