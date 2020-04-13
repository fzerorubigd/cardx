package cards

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/sprig"
	yaml "gopkg.in/yaml.v2"
)

// DeckWithData is the deck with data from the project file
type DeckWithData struct {
	Deck `yaml:",inline"`

	Template string
	Data     string
}

// Project is the project structure
type Project struct {
	Name      string
	Author    string
	Copyright string
	Version   string
	Decks     []*DeckWithData
}

func (dw *DeckWithData) loadCards() error {
	f, err := os.Open(dw.Data)
	if err != nil {
		return fmt.Errorf("load file failed: %w", err)
	}
	defer f.Close()

	if err := dw.LoadCSV(f); err != nil {
		return fmt.Errorf("failed to load the data: %w", err)
	}

	return nil
}

func (dw *DeckWithData) validateTemplate() error {
	tmpl, err := template.New(fmt.Sprintf("%s_user_template", dw.Name)).Funcs(sprig.FuncMap()).ParseFiles(dw.Template)
	if err != nil {
		return fmt.Errorf("load deck %q template failed: %w", dw.Name, err)
	}

	for _, tpl := range tmpl.Templates() {
		switch tpl.Name() {
		case "card":
			dw.HasCardTemplate = true
		case "css":
			dw.HasCSSTemplate = true
			fmt.Println("OO")
		}
	}

	return nil
}

// LoadYML try to load a project from a yaml file, assume the current directory is
// the project path
func (p *Project) LoadYML(r io.Reader) error {
	if err := yaml.NewDecoder(r).Decode(p); err != nil {
		return fmt.Errorf("loading project yaml file failed: %w", err)
	}

	for idx, deck := range p.Decks {
		if deck.Name == "" {
			return fmt.Errorf("deck number %d has no name", idx+1)
		}
		// TODO : more validation on size
		if deck.Width <= 0 || deck.Height <= 0 {
			return fmt.Errorf("the %dx%d is not a valid size for deck %q", deck.Width, deck.Height, deck.Name)
		}

		// set some default
		if deck.Copyright == "" {
			deck.Copyright = p.Copyright
		}

		if deck.Version == "" {
			deck.Version = p.Version
		}

		if deck.Author == "" {
			deck.Author = p.Author
		}

		if err := deck.loadCards(); err != nil {
			return fmt.Errorf("loading cards for deck %q failed: %w", deck.Name, err)
		}

		if err := deck.validateTemplate(); err != nil {
			return fmt.Errorf("validationg deck %q template failed: %w", deck.Name, err)
		}
	}

	return nil

}

// NewProject loads a project from the folder
func NewProject(root string) (*Project, error) {
	path := filepath.Join(root, "project.yml")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("loading %q failed: %w", path, err)
	}
	// TODO : backup current dir
	if err := os.Chdir(root); err != nil {
		return nil, fmt.Errorf("change directory to project root failed: %w", err)
	}

	p := &Project{}

	if err := p.LoadYML(f); err != nil {
		return nil, fmt.Errorf("load project failed: %w", err)
	}

	return p, nil
}
