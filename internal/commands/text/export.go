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
			return "const banner = `" + rendered + "`", nil
		}
		return "const banner = " + strconv.Quote(rendered), nil
	case "js", "ts":
		return "console.log(`" + strings.ReplaceAll(strings.ReplaceAll(rendered, "\\", "\\\\"), "`", "\\`") + "`);", nil
	case "py":
		if !strings.Contains(rendered, `"""`) {
			return `print("""` + rendered + `""")`, nil
		}
		return "print(" + strconv.Quote(rendered) + ")", nil
	case "json":
		data, _ := json.Marshal(rendered)
		return string(data), nil
	case "html":
		return "<pre>" + html.EscapeString(rendered) + "</pre>", nil
	case "rust":
		return `println!("{}", r#"` + strings.ReplaceAll(rendered, `"#`, `"\#`) + `"#);`, nil
	case "csharp":
		return `Console.WriteLine(@"` + strings.ReplaceAll(rendered, `"`, `""`) + `");`, nil
	case "sh":
		return "cat <<'EOF'\n" + rendered + "EOF", nil
	default:
		return "", fmt.Errorf("unknown export format %q", format)
	}
}
