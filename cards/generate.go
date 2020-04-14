package cards

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig"
)

// Generate the card set in html format
func Generate(p *Project, path string, files []string, getter func(string) []byte) error {
	if path == "" {
		path = "generated"
	}

	baseTpl := getter("templates/index.html")
	for _, deck := range p.Decks {
		root := filepath.Join(path, deck.Name)
		_ = os.MkdirAll(root, 0750)
		fl := filepath.Join(root, "index.html")
		w, err := os.Create(fl)
		if err != nil {
			return fmt.Errorf("create output file failed: %w", err)
		}

		a, err := generateDeck(w, deck, string(baseTpl))
		if err != nil {
			return fmt.Errorf("generate deck %q failed: %w", deck.Name, err)
		}
		for _, fl := range files {
			if fl == "templates/index.html" {
				continue
			}

			if err := copyFile(strings.TrimPrefix(fl, "templates/"), root, getter(fl)); err != nil {
				return fmt.Errorf("copy assets %q failed: %w", fl, err)
			}
		}

		for path, translated := range a.project {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("copy project assets %q failed: %w", path, err)
			}
			if err := copyFile(translated, root, data); err != nil {
				return fmt.Errorf("copy assets %q failed: %w", fl, err)
			}
		}

	}

	return nil
}

type assets struct {
	prefix  string
	project map[string]string
	//global  map[string]string
}

func (a *assets) translateProject(in string) string {
	translated, ok := a.project[in]
	if ok {
		return translated
	}

	translated = filepath.Join(a.prefix, in)
	a.project[in] = translated
	return translated
}

func generateDeck(w io.Writer, d *DeckWithData, baseTemplate string) (*assets, error) {
	fns := sprig.FuncMap()
	ret := &assets{
		prefix:  "assets",
		project: make(map[string]string),
	}
	fns["asset"] = func(in string) string { return ret.translateProject(in) }
	tmpl, err := template.Must(template.New("single_deck").Funcs(fns).Parse(baseTemplate)).ParseFiles(d.Template)
	if err != nil {
		return nil, fmt.Errorf("parsing template for deck %q failed: %w", d.Name, err)
	}

	if !d.HasCSSTemplate {
		tmpl = template.Must(tmpl.Parse(`{{ define "css" }}{{ end  }}`))
	}

	if !d.HasCardTemplate {
		tmpl = template.Must(tmpl.Parse(`{{ define "card" }}{{ .Name  }}<br> No Card template, you need to define one in your project{{ end  }}`))
	}

	if err := tmpl.ExecuteTemplate(w, "deck", d.Deck); err != nil {
		return nil, fmt.Errorf("execute template for deck %q failed: %w", d.Name, err)
	}

	return ret, nil
}

func copyFile(src string, root string, data []byte) error {
	fl := filepath.Join(root, src)
	dir := filepath.Dir(fl)
	_ = os.MkdirAll(dir, 0750)
	return ioutil.WriteFile(fl, data, 0640)
}
