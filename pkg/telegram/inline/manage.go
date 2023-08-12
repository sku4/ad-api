package inline

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ManageSubInline struct {
	SubscriptionID uint64
	Back           *PaginationInline
}

const (
	splitMSICount = 2
)

func NewManageSubInline() *ManageSubInline {
	return &ManageSubInline{
		Back: NewPaginationInline(),
	}
}

func NewManageSubInlineExt(subscriptionID uint64, back *PaginationInline) *ManageSubInline {
	return &ManageSubInline{
		SubscriptionID: subscriptionID,
		Back:           back,
	}
}

func (mi *ManageSubInline) Type() string {
	return manageSubInlineType
}

func (mi *ManageSubInline) Serialize() string {
	return strings.Join([]string{
		strconv.FormatUint(mi.SubscriptionID, 10),
		mi.Back.Serialize(),
	}, delimiter)
}

func (mi *ManageSubInline) UnSerialize(s string) error {
	args := strings.SplitN(s, delimiter, splitMSICount)

	if len(args) > 0 {
		subscriptionID, err := strconv.ParseUint(args[0], 10, 0)
		if err != nil {
			return errors.Wrap(err, "manage sub inline: unserialize")
		}
		mi.SubscriptionID = subscriptionID
	}

	if len(args) > 1 {
		page := NewPaginationInline()
		err := page.UnSerialize(args[1])
		if err != nil {
			return errors.Wrap(err, "manage sub inline: unserialize")
		}
		mi.Back = page
	}

	return nil
}
