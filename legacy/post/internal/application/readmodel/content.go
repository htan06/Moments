package readmodel

// Query-side content model, separate from domain.
// Mention displays username resolved via post_mentions JOIN.
// Hashtag displays name stored in JSONB, no JOIN needed.

type Node struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	UserID int64  `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

type Content struct {
	Nodes []Node `json:"nodes"`
}

func EnrichContent(content Content) {
	for i := range content.Nodes {
		if content.Nodes[i].Type == "hashtag" && content.Nodes[i].Name != "" {
			content.Nodes[i].Text = "#" + content.Nodes[i].Name
		}
	}
}

func EnrichMentions(content Content, usernames map[int64]string) {
	for i := range content.Nodes {
		if content.Nodes[i].Type != "mention" {
			continue
		}
		if name, ok := usernames[content.Nodes[i].UserID]; ok {
			content.Nodes[i].Text = "@" + name
		}
	}
}
