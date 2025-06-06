package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/api"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/log"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/srv"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/rs/zerolog"
)

type ResultSenderConfig struct {
	SendInterval time.Duration
}

func MakeDefaultSenderConfig() ResultSenderConfig {
	return ResultSenderConfig{
		SendInterval: 5 * time.Second,
	}
}

type TaskResultSource interface {
	TryGetPending(ctx context.Context) (*types.TaskResult, error)

	SetAsSubmitted(ctx context.Context, taskId types.TaskId) error
}

type TaskResultSender struct {
	srv.WorkerLoop

	requestHandler api.TaskRequestHandler
	resultSource   TaskResultSource
	config         ResultSenderConfig
}

func NewTaskResultSender(
	requestHandler api.TaskRequestHandler,
	resultSource TaskResultSource,
	metrics srv.WorkerMetrics,
	logger logging.Logger,
) *TaskResultSender {
	sender := &TaskResultSender{
		requestHandler: requestHandler,
		resultSource:   resultSource,
		config:         MakeDefaultSenderConfig(),
	}

	loopConfig := srv.NewWorkerLoopConfig("task_result_sender", sender.config.SendInterval, sender.processPendingResult)
	sender.WorkerLoop = srv.NewWorkerLoop(loopConfig, metrics, logger)
	return sender
}

func (s *TaskResultSender) processPendingResult(ctx context.Context) error {
	pendingResult, err := s.getPending(ctx)
	if err != nil || pendingResult == nil {
		return err
	}

	if err := s.sendPending(ctx, pendingResult); err != nil {
		return fmt.Errorf("taskId=%s: %w", pendingResult.TaskId, err)
	}

	return nil
}

func (s *TaskResultSender) getPending(ctx context.Context) (*types.TaskResult, error) {
	pendingResult, err := s.resultSource.TryGetPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get next task result: %w", err)
	}
	if pendingResult == nil {
		s.Logger.Debug().Msg("no task result available, waiting for new one")
	}
	return pendingResult, nil
}

func (s *TaskResultSender) sendPending(ctx context.Context, result *types.TaskResult) error {
	log.NewTaskResultEvent(s.Logger, zerolog.DebugLevel, result).Msg("sending task result")

	if err := s.requestHandler.SetTaskResult(ctx, result); err != nil {
		return fmt.Errorf("failed to send task result: %w", err)
	}

	log.NewTaskResultEvent(s.Logger, zerolog.DebugLevel, result).Msg("task result successfully sent")

	if err := s.resultSource.SetAsSubmitted(ctx, result.TaskId); err != nil {
		return fmt.Errorf("failed to set task result as submitted: %w", err)
	}

	log.NewTaskResultEvent(s.Logger, zerolog.DebugLevel, result).Msg("task result is set as submitted")
	return nil
}
