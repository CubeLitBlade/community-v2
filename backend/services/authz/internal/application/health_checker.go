package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

type ServingStatus int

const (
	StatusUnknown ServingStatus = iota
	StatusServing
	StatusNotServing
)

type ServingStatusSetter interface {
	SetStatus(status ServingStatus)
}

type HealthChecker struct {
	checkers []port.HealthChecker
	setter   ServingStatusSetter
	logger   *slog.Logger

	ready  bool
	cancel context.CancelFunc
}

func NewHealthChecker(checkers []port.HealthChecker, setter ServingStatusSetter, logger *slog.Logger) *HealthChecker {
	return &HealthChecker{
		checkers: checkers,
		setter:   setter,
		logger:   logger.With(slog.String("component", "health_checker")),

		ready:  true,
		cancel: nil,
	}
}

func (s *HealthChecker) Start(ctx context.Context) error {
	s.logger.InfoContext(ctx, "health checker begins its deathwatch")

	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.cancel = cancel

	go s.run(runCtx)

	return nil
}

func (s *HealthChecker) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *HealthChecker) run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := StatusNotServing
			if s.checkAll(ctx) {
				status = StatusServing
			}

			s.setter.SetStatus(status)
		case <-ctx.Done():
			s.setter.SetStatus(StatusNotServing)
			return
		}
	}
}

func (s *HealthChecker) checkAll(ctx context.Context) bool {
	for _, checker := range s.checkers {
		checkCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		err := checker.Check(checkCtx)

		cancel()

		if err != nil {
			if s.ready {
				s.ready = false
				s.logger.ErrorContext(ctx, "dependency dropped dead",
					slog.String("dependency", checker.Name()),
					slog.Any("error", err),
				)
			} else {
				s.logger.WarnContext(ctx, "still no pulse on dependency",
					slog.String("dependency", checker.Name()),
					slog.Any("error", err),
				)
			}

			return false
		}
	}

	if !s.ready {
		s.ready = true
		s.logger.InfoContext(ctx, "necromancy! dependency heartbeat restored")
	}

	return true
}
