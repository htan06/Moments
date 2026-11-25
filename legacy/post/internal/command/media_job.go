package command

import "github.com/htan06/Moments/post/internal/domain"

// ProcessMediaJob is the message sent to the media worker. It is an
// integration contract, not part of the Post domain (moved from
// internal/domain/post/job.go).

type TaskType string

const (
	HLSVideo    TaskType = "HLS_VIDEO"
	ResizeImage TaskType = "RESIZE_IMAGE"
)

type TaskParam struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (p TaskParam) Valid() bool { return p.Width > 0 && p.Height > 0 }

type ObjectLocation struct {
	Bucket   string `json:"bucket"`
	ObjectID string `json:"object_id"`
}

func (o ObjectLocation) Valid() bool { return o.Bucket != "" && o.ObjectID != "" }

type Task struct {
	ObjectDes ObjectLocation `json:"object_des"`
	TaskType  TaskType       `json:"task_type"`
	Param     TaskParam      `json:"task_param"`
}

type ProcessMediaJob struct {
	ObjectSrc ObjectLocation `json:"object_src"`
	MediaType string         `json:"media_type"`
	Tasks     []Task         `json:"tasks"`
}

func (j ProcessMediaJob) Validate() error {
	if !j.ObjectSrc.Valid() {
		return domain.ErrMediaKeyEmpty
	}
	if len(j.Tasks) == 0 {
		return domain.ErrMediaCountInvalid
	}
	for _, t := range j.Tasks {
		if !t.ObjectDes.Valid() {
			return domain.ErrMediaKeyEmpty
		}
		if !t.Param.Valid() {
			return domain.ErrWidthInvalid
		}
	}
	return nil
}
