package post

type Visibility string

const (
	Public  Visibility = "PUBLIC"
	Private Visibility = "PRIVATE"
)

func (v Visibility) Valid() bool    { return v == Public || v == Private }
func (v Visibility) String() string { return string(v) }

type AspectRatio string

const (
	Square     AspectRatio = "SQUARE"
	Classic    AspectRatio = "CLASSIC"
	Standard   AspectRatio = "STANDARD"
	Portrait   AspectRatio = "PORTRAIT"
	Widescreen AspectRatio = "WIDESCREEN"
)

var aspectRatios = map[AspectRatio][2]int{
	Square:     {1, 1},
	Classic:    {3, 2},
	Standard:   {4, 3},
	Portrait:   {3, 4},
	Widescreen: {16, 9},
}

func (a AspectRatio) Valid() bool    { _, ok := aspectRatios[a]; return ok }
func (a AspectRatio) String() string { return string(a) }

// Ratio returns the width:height of the aspect ratio, (0, 0) if invalid.
func (a AspectRatio) Ratio() (width, height int) {
	r := aspectRatios[a]
	return r[0], r[1]
}

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPublished Status = "PUBLISHED"
)

func (s Status) Valid() bool    { return s == StatusPending || s == StatusPublished }
func (s Status) String() string { return string(s) }

type MediaType string

const (
	Image MediaType = "IMAGE"
	Video MediaType = "VIDEO"
)

func (t MediaType) Valid() bool    { return t == Image || t == Video }
func (t MediaType) String() string { return string(t) }

type MediaStatus string

const (
	MediaProcessing MediaStatus = "PROCESSING"
	MediaPublished  MediaStatus = "PUBLISHED"
)

func (s MediaStatus) Valid() bool    { return s == MediaProcessing || s == MediaPublished }
func (s MediaStatus) String() string { return string(s) }
