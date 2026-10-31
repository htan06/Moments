package outbound

import (
	"context"
	"user-service/internal/application/domain/job"
)

type UserProducer interface {
	SendJob(ctx context.Context, job job.ProcessMediaJob) error
}
