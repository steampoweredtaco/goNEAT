// Package experiment defines standard evolutionary epochs evaluators and experimental data samples collectors.
package experiment

import (
	"context"
	"errors"
	"github.com/steampoweredtaco/goNEAT/v3/neat"
	"github.com/steampoweredtaco/goNEAT/v3/neat/genetics"
	"time"
)

// EmptyDuration is to return when average duration can not tbe estimated (empty trials or generations)
const EmptyDuration = time.Duration(-1)

// GenerationEvaluator the interface describing evaluator for one epoch (generation) of the evolutionary process.
type GenerationEvaluator interface {
	// GenerationEvaluate Invoked to evaluate one generation of population of organisms within given
	// execution context.
	GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *Generation) error
}

// TrialRunObserver defines observer to be notified about experiment's trial lifecycle methods
type TrialRunObserver interface {
	// TrialRunStarted invoked to notify that new trial run just started. Invoked before any epoch evaluation in that trial run
	TrialRunStarted(trial *Trial)
	// TrialRunFinished invoked to notify that the trial run just finished. Invoked after all epochs evaluated or successful solver found.
	TrialRunFinished(trial *Trial)
	// EpochEvaluated invoked to notify that evaluation of specific epoch completed.
	EpochEvaluated(trial *Trial, epoch *Generation)
}

// Returns appropriate executor type from given context
func epochExecutorForContext(ctx context.Context) (genetics.PopulationEpochExecutor, error) {
	options, ok := neat.FromContext(ctx)
	if !ok {
		return nil, neat.ErrNEATOptionsNotFound
	}
	switch options.EpochExecutorType {
	case neat.EpochExecutorTypeSequential:
		return &genetics.SequentialPopulationEpochExecutor{}, nil
	case neat.EpochExecutorTypeParallel:
		return &genetics.ParallelPopulationEpochExecutor{}, nil
	default:
		return nil, errors.New("unsupported epoch executor type requested")
	}
}
