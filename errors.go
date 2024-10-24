package cosmosign

import (
	"errors"
)

var ErrGRPCClientIsNil = errors.New("grpc client must be set")
