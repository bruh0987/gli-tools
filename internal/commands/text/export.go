package text

import (
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"
)

func Export(rendered string, format string) (string, error) {
	switch strings.ToLower(format) {
	case "", "text":
		return rendered, nil
	case "go":
		if !strings.Contains(rendered, "`") {
			return "`" + rendered + "`", nil
		}
		return strconv.Quote(rendered), nil
	case "js", "ts":
		return "`" + strings.ReplaceAll(strings.ReplaceAll(rendered, "\\", "\\\\"), "`", "\\`") + "`", nil
	case "py":
		if !strings.Contains(rendered, `"""`) {
			return `"""` + rendered + `"""`, nil
		}
		return strconv.Quote(rendered), nil
	case "json":
		data, _ := json.Marshal(rendered)
		return string(data), nil
	case "html":
		return "<pre>" + html.EscapeString(rendered) + "</pre>", nil
	case "rust":
		return `r#"` + strings.ReplaceAll(rendered, `"#`, `"\#`) + `"#`, nil
	case "csharp":
		return `@"` + strings.ReplaceAll(rendered, `"`, `""`) + `"`, nil
	case "sh":
		return "cat <<'EOF'\n" + rendered + "EOF", nil
	default:
		return "", fmt.Errorf("unknown export format %q", format)
	}
}
