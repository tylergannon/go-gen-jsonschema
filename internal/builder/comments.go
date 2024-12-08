package builder

import (
	"github.com/dave/dst"
	"strings"
)

func buildComments(decs *dst.NodeDecs) string {
	return formatComments(appendDecorations(clipCommentsString(decs.Start), clipCommentsString(decs.End)))
}

// formatComments removes either "//" or "// " from the front of each
// string. It organizes the text (somewhat crudely) into paragraphs,
// and honors newlines and leading whitespace within code blocks delimited
// by fences.
func formatComments(comments dst.Decorations) string {
	b := strings.Builder{}
	var _comments = make([]string, len(comments))
	for i, dec := range comments {
		trimmed := strings.TrimRight(
			strings.TrimPrefix(
				strings.TrimPrefix(dec, "// "), "//"),
			" \t",
		)
		_comments[i] = trimmed
	}
	if len(comments) == 0 {
		return ""
	}
	inCodeBlock := strings.HasPrefix(_comments[0], "```")
	b.WriteString(_comments[0])
	for i := 1; i < len(_comments); i++ {
		if inCodeBlock {
			b.WriteString("\n")
		} else if len(_comments[i]) == 0 {
			b.WriteString("\n")
			continue
		} else if len(_comments[i-1]) > 0 {
			b.WriteString(" ")
		}
		b.WriteString(_comments[i])
		if strings.HasPrefix(_comments[i], "```") {
			inCodeBlock = !inCodeBlock
			_comments[i] = ""
		}
	}
	return b.String()
}

// clipCommentsString realizes the "comment" decorations for a type as being
// solely the contiguous non-empty strings at the end of the input.
// Returns the who input if there are no empty strings.  Otherwise, returns
// the portion of the input following the last occurrence of an empty string.
func clipCommentsString(decs dst.Decorations) dst.Decorations {
	for i := len(decs) - 1; i >= 0; i-- {
		if strings.TrimSpace(decs[i]) == "" {
			return decs[i+1:]
		}
	}
	return decs
}

func appendDecorations(start dst.Decorations, end dst.Decorations) dst.Decorations {
	result := append(dst.Decorations{}, start...)
	return append(result, end...)
}
