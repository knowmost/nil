package log

import (
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/rs/zerolog"
)

func NewTaskEvent(
	logger logging.Logger,
	level zerolog.Level,
	task *types.Task,
) *logging.Event {
	return logger.WithLevel(level).
		Stringer(logging.FieldTaskId, task.Id).
		Stringer(logging.FieldBatchId, task.BatchId).
		Stringer(logging.FieldTaskType, task.TaskType).
		Interface(logging.FieldTaskParentId, task.ParentTaskId)
}

func NewTaskResultEvent(
	logger logging.Logger,
	level zerolog.Level,
	result *types.TaskResult,
) *logging.Event {
	event := logger.WithLevel(level).
		Stringer(logging.FieldTaskId, result.TaskId).
		Stringer(logging.FieldTaskExecutorId, result.Sender).
		Bool("isSuccess", result.IsSuccess())

	if result.IsSuccess() {
		return event
	}

	return event.
		Str("errorText", result.Error.ErrText).
		Stringer("errorType", result.Error.ErrType)
}

func NewStateUpdateEvent(
	logger logging.Logger,
	level zerolog.Level,
	stateUpdate *types.UpdateStateData,
) *logging.Event {
	return logger.WithLevel(level).
		Stringer(logging.FieldBatchId, stateUpdate.BatchId).
		Stringer("OldStateRoot", stateUpdate.OldProvedStateRoot).
		Stringer("NewStateRoot", stateUpdate.NewProvedStateRoot).
		Int("BlobCount", len(stateUpdate.DataProofs)).
		Stringer("L2Tol1Root", stateUpdate.L2Tol1Root).
		Stringer("DepositNonce", stateUpdate.DepositNonce).
		Stringer("L1MessageHash", stateUpdate.L1MessageHash)
}
