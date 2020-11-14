// Package atom provides and implements the atom plugin.
package atom

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"github.com/spf13/afero"
	"github.com/verless/verless/model"
)

const (
	// filename is the filename for the RSS feed.
	filename string = "atom.xml"
)

// New creates a new atom plugin that generated a RSS feed with the
// provided metadata and stores the XML file in outputDir.
func New(meta *model.Meta, fs afero.Fs, outputDir string) *atom {
	a := atom{
		meta: meta,
		feed: &feeds.Feed{
			Title:       meta.Title,
			Link:        &feeds.Link{Href: meta.Base},
			Description: meta.Description,
			Author:      &feeds.Author{Name: meta.Author},
			Updated:     time.Time{},
			Created:     time.Now(),
			Subtitle:    meta.Subtitle,
		},
		fs:        fs,
		outputDir: outputDir,
	}

	return &a
}

// atom is the actual atom plugin. It stores all RSS feed items
// as a feeds.Feed and renders those items in a XML file.s
type atom struct {
	meta      *model.Meta
	feed      *feeds.Feed
	feedMutex sync.RWMutex
	fs        afero.Fs
	outputDir string
}

// ProcessPage takes a page to be processed by the plugin, reads
// metadata for that page and creates a new feed item from it.
func (a *atom) ProcessPage(page *model.Page) error {
	if page.Hidden || page.IsCustomListPage() {
		return nil
	}

	canonical := fmt.Sprintf("%s%s/%s", a.meta.Base, page.Route, page.ID)

	item := &feeds.Item{
		Title:       page.Title,
		Link:        &feeds.Link{Href: canonical},
		Description: page.Description,
		Id:          canonical,
		Created:     page.Date,
	}

	a.feedMutex.Lock()
	a.feed.Add(item)
	a.feedMutex.Unlock()
	return nil
}

// PreWrite isn't needed by the atom plugin.
func (a *atom) PreWrite(_ *model.Site) error {
	return nil
}

// PostWrite writes the internal feed.Feed instance into a file
// directly in the output directory.
func (a *atom) PostWrite() error {
	path := filepath.Join(a.outputDir, filename)
	atomFile, err := a.fs.Create(path)
	if err != nil {
		return err
	}

	return a.feed.WriteAtom(atomFile)
}
