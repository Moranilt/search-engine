package parser

import (
	"reflect"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	t.Run("extract single link", func(t *testing.T) {
		html := []byte("<html><head></head><body><p><a href=\"/home/profile\">Go back</a></p></body></html>")
		got := ExtractLinks(html)
		want := []string{"/home/profile"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("extract single link with additional attributes", func(t *testing.T) {
		html := []byte("<html><head></head><body><p><a alt=\"profile page\" title=\"profile page\" href=\"/home/profile\">Go back</a></p></body></html>")
		got := ExtractLinks(html)
		want := []string{"/home/profile"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("extract single link in single quotes", func(t *testing.T) {
		html := []byte("<html><head></head><body><p><a alt='profile page' title='profile page' href='/home/profile'>Go back</a></p></body></html>")
		got := ExtractLinks(html)
		want := []string{"/home/profile"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("extract multiple links", func(t *testing.T) {
		html := []byte(`<html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
				</p>
			</div>
		</div>
		</body></html>`)
		got := ExtractLinks(html)
		want := []string{"/home/profile", "/home/contacts", "/home/blog"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("extract single link but not from article", func(t *testing.T) {
		html := []byte("<html><head></head><body><p><a alt=\"profile page\" title=\"profile page\" href=\"/home/profile\">Go back</a><article href='/broken/link'>Some article</article><abbr title='Some abbr' href='/broken/link2'>CSS</abbr></p></body></html>")
		got := ExtractLinks(html)
		want := []string{"/home/profile"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// t.Run("extract single hacked link", func(t *testing.T) {
	// 	html := []byte("<html><head></head><body><p><a href='/home/profile?id=623_dgha21hgdsa\x3ca href='/hacked/'\x3e&name=;--\x3ca href=\"/hacked\"\x3e'>Go back</a></p></body></html>")
	// 	got := ExtractLinks(html)
	// 	want := []string{"/home/profile"}
	// 	if !reflect.DeepEqual(got, want) {
	// 		t.Errorf("got %v, want %v", got, want)
	// 	}
	// })

	t.Run("extract unique links", func(t *testing.T) {
		html := []byte(`<html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html>`)
		got := ExtractLinks(html)
		want := []string{"/home/profile", "/home/contacts", "/home/blog"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func BenchmarkExtractLinks(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExtractLinks([]byte(`<html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html><html><head></head><body>
		<p>
			<a href="/home/profile">Go back</a>
			<a href="/home/contacts">Contacts</a>
		</p>
		<div>
			<div>
				<h1>Hello world</h1>
				<a href="/home/blog">Blog</a>
				<p>
					<a>Broken link</a>
					<a href="/home/profile">main page</a>
				</p>
				<a href="/home/contacts">Company contacts</a>
			</div>
		</div>
		</body></html>`))
	}
}
