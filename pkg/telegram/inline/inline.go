package inline

import (
	"strings"

	"github.com/pkg/errors"
)

type Inlineable interface {
	Type() string
	Serialize() string
	UnSerialize(string) error
}

const (
	splitInlineCount = 4
	countPartsThree  = 3
)

type Inline struct {
	Entity  string
	Command string
	Data    any
}

func NewInline() *Inline {
	return &Inline{}
}

func NewInlineExt(entity string, command string, data any) *Inline {
	return &Inline{
		Entity:  entity,
		Command: command,
		Data:    data,
	}
}

func (i *Inline) Serialize() string {
	var inline Inlineable
	switch v := i.Data.(type) {
	case *ManageSubInline:
		inline = v
	case *PaginationInline:
		inline = v
	default:
		return ""
	}

	return strings.Join([]string{
		i.Entity,
		i.Command,
		inline.Type(),
		inline.Serialize(),
	}, delimiter)
}

func (i *Inline) UnSerialize(s string) (*Inline, error) {
	args := strings.SplitN(s, delimiter, splitInlineCount)

	if len(args) > 0 {
		i.Entity = args[0]
	}

	if len(args) > 1 {
		i.Command = args[1]
	}

	if len(args) > countPartsThree {
		t := args[2]
		var inline Inlineable
		switch t {
		case manageSubInlineType:
			inline = NewManageSubInline()
		case paginationInlineType:
			inline = NewPaginationInline()
		default:
			return nil, errors.Wrap(ErrUnknownType, "inline")
		}

		err := inline.UnSerialize(args[3])
		if err != nil {
			return nil, errors.Wrap(err, "inline: unserialize")
		}
		i.Data = inline
	}

	return i, nil
}
