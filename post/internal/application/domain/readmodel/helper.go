package readmodel

import (
	"github.com/htan06/Moments/post/internal/application/domain/vo"
)

func EnrichContent(content []vo.Content) {
	for i := range content {
		switch content[i].Type {
		case vo.Mention:
			content[i].Text = "@" + content[i].Text
		case vo.Hashtag:
			content[i].Text = "#" + content[i].Text
		}
	}
}
