package sensor

import (
	"context"

	"github.com/google/uuid"
)

type Sensor interface {
	Start(ctx context.Context) error
	Stop()
	IsRunning() bool
	ID() uuid.UUID
}
