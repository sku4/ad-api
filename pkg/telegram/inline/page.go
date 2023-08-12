package inline

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Page struct {
	PageID int
}

const (
	splitPageCount = 1
)

func NewPage() *Page {
	return &Page{
		PageID: 1,
	}
}

func NewPageExt(pageID int) *Page {
	return &Page{
		PageID: pageID,
	}
}

func (p *Page) Serialize() string {
	return strings.Join([]string{
		strconv.Itoa(p.PageID),
	}, delimiter)
}

func (p *Page) UnSerialize(s string) (*Page, error) {
	args := strings.SplitN(s, delimiter, splitPageCount)

	if len(args) > 0 {
		pageID, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, errors.Wrap(err, "pagination inline: atoi")
		}
		p.PageID = pageID
	}

	return p, nil
}
