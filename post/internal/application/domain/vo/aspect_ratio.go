package vo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/htan06/Moments/post/internal/application/domain"
	"github.com/htan06/Moments/post/internal/errs"
)

type AspectRatio struct {
	Width  int
	Height int
}

func NewAspectRatio(value string) (AspectRatio, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			nil,
			domain.AspectRatioInvalid,
		)
	}

	Width, err := strconv.Atoi(parts[0])
	if err != nil {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			err,
			domain.AspectRatioInvalid,
		)
	}

	Height, err := strconv.Atoi(parts[1])
	if err != nil {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			err,
			domain.AspectRatioInvalid,
		)
	}

	ratio := AspectRatio{
		Width:  Width,
		Height: Height,
	}

	if !ratio.IsSupported() {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			nil,
			domain.AspectRatioInvalid,
		)
	}

	return ratio, nil
}

func (a AspectRatio) String() string {
	return fmt.Sprintf("%d:%d", a.Width, a.Height)
}

var supportedAspectRatios = map[AspectRatio]struct{}{
	{Width: 1, Height: 1}:  {},
	{Width: 3, Height: 4}:  {},
	{Width: 4, Height: 3}:  {},
	{Width: 5, Height: 3}:  {},
	{Width: 16, Height: 9}: {},
}

func (a AspectRatio) Demensions() (widht int, Height int) {
	return a.Width, a.Height
}

func (a AspectRatio) IsSupported() bool {
	_, ok := supportedAspectRatios[a]
	return ok
}
