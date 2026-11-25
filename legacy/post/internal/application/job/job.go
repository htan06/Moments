package job

type TaskType string
type MediaType string

type TaskParam struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

const (
	HLSVideo    TaskType = "HLS_VIDEO"
	ResizeImage TaskType = "RESIZE_IMAGE"

	Image MediaType = "IMAGE"
	Video MediaType = "VIDEO"
)

type ObjectLocation struct {
	Bucket   string `json:"bucket"`
	ObjectID string `json:"object_id"`
}

type Task struct {
	ObjectDes ObjectLocation `json:"object_des"`
	TaskType  TaskType       `json:"task_type"`
	Param     TaskParam      `json:"task_param"`
}

type ProcessMediaJob struct {
	ObjectSrc ObjectLocation `json:"object_src"`
	MediaType MediaType      `json:"media_type"`
	Tasks     []Task         `json:"tasks"`
}
