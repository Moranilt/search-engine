package parser

import "testing"

func TestExtractTitle(t *testing.T) {
	t.Run("extract simple title", func(t *testing.T) {
		html := []byte(`
		<html>
			<head>
				<title>Title of the page</title>
			</head>
			<body>
				
			</body>
		</html>
		`)

		got := ExtractTitle(html)
		want := "Title of the page"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
