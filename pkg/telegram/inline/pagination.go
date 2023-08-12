package inline

import (
	"strings"

	"github.com/pkg/errors"
)

type PaginationInline struct {
	Current *Page
}

const (
	splitPICount = 1
	delimiterPI  = "*"
)

func NewPaginationInline() *PaginationInline {
	return &PaginationInline{
		Current: NewPage(),
	}
}

func NewPaginationInlineExt(current *Page) *PaginationInline {
	return &PaginationInline{
		Current: current,
	}
}

func (pi *PaginationInline) Type() string {
	return paginationInlineType
}

func (pi *PaginationInline) Serialize() string {
	current := ""
	if pi.Current != nil {
		current = pi.Current.Serialize()
	}

	return strings.Join([]string{
		current,
	}, delimiterPI)
}

func (pi *PaginationInline) UnSerialize(s string) error {
	args := strings.SplitN(s, delimiterPI, splitPICount)

	if len(args) > 0 {
		page, err := NewPage().UnSerialize(args[0])
		if err != nil {
			return errors.Wrap(err, "pagination inline: unserialize")
		}
		pi.Current = page
	}

	return nil
}
