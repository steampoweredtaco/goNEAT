package experiment

import (
	"bytes"
	"fmt"
	"github.com/sbinet/npyio/npz"
	"github.com/steampoweredtaco/goNEAT/v3/neat/genetics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/mat"
	"math"
	"testing"
	"time"
)

func TestExperiment_Write_Read(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test Encode Decode", Trials: make(Trials, 3)}
	for i := 0; i < len(ex.Trials); i++ {
		ex.Trials[i] = *buildTestTrial(i+1, 10)
	}

	// Write experiment
	var buff bytes.Buffer
	err := ex.Write(&buff)
	require.NoError(t, err, "Failed to write experiment")

	// Read experiment
	data := buff.Bytes()
	newEx := Experiment{}
	err = newEx.Read(bytes.NewBuffer(data))
	require.NoError(t, err, "failed to read experiment")

	// Deep compare results
	assert.Equal(t, ex.Id, newEx.Id)
	assert.Equal(t, ex.Name, newEx.Name)
	require.Len(t, newEx.Trials, len(ex.Trials))

	for i := 0; i < len(ex.Trials); i++ {
		assert.EqualValues(t, ex.Trials[i], newEx.Trials[i])
	}
}

func TestExperiment_Write_writeError(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test Encode Decode", Trials: make(Trials, 3)}
	for i := 0; i < len(ex.Trials); i++ {
		ex.Trials[i] = *buildTestTrial(i+1, 10)
	}

	errWriter := ErrorWriter(1)
	err := ex.Write(&errWriter)
	assert.EqualError(t, err, alwaysErrorText)
}

func TestExperiment_Read_readError(t *testing.T) {
	errReader := ErrorReader(1)

	newEx := Experiment{}
	err := newEx.Read(&errReader)
	assert.EqualError(t, err, alwaysErrorText)
}

func TestExperiment_WriteNPZ(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test Encode Decode", Trials: make(Trials, 3)}
	for i := 0; i < len(ex.Trials); i++ {
		ex.Trials[i] = *buildTestTrial(i+1, 10)
	}

	// Write experiment
	var buff bytes.Buffer
	err := ex.WriteNPZ(&buff)
	require.NoError(t, err, "Failed to write experiment")
	assert.True(t, buff.Len() > 0)

	// Read experiment and compare values
	r, err := npz.NewReader(bytes.NewReader(buff.Bytes()), int64(buff.Len()))
	require.NoError(t, err)
	require.NotNil(t, r)

	expectedTrialsNumber := Floats{float64(len(ex.Trials))}
	trialsNumber := Floats{}
	err = r.Read("trials_number", &trialsNumber)
	assert.NoError(t, err, "failed to read trials number")
	assert.EqualValues(t, expectedTrialsNumber, trialsNumber, "wrong trials number")

	expectedFitness, expectedAges, expectedComplexity := ex.fitnessAgeComplexityMat()

	trialsFitness := &mat.Dense{}
	err = r.Read("trials_fitness", trialsFitness)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedFitness, trialsFitness, "wrong fitness")

	trialsAges := &mat.Dense{}
	err = r.Read("trials_ages", trialsAges)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedAges, trialsAges, "wrong ages")

	trialsComplexity := &mat.Dense{}
	err = r.Read("trials_complexity", trialsComplexity)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedComplexity, trialsComplexity, "wrong complexity")

	for i, tr := range ex.Trials {
		expectedFitness, expectedAges, expectedComplexities := tr.Average()

		key := fmt.Sprintf("trial_%d_epoch_mean_fitnesses", i)
		fitness := Floats{}
		err = r.Read(key, &fitness)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedFitness, fitness)

		key = fmt.Sprintf("trial_%d_epoch_mean_ages", i)
		ages := Floats{}
		err = r.Read(key, &ages)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedAges, ages)

		key = fmt.Sprintf("trial_%d_epoch_mean_complexities", i)
		complexities := Floats{}
		err = r.Read(key, &complexities)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedComplexities, complexities)

		key = fmt.Sprintf("trial_%d_epoch_best_fitnesses", i)
		expectedChapFitness := tr.ChampionsFitness()
		champFitness := Floats{}
		err = r.Read(key, &champFitness)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedChapFitness, champFitness)

		key = fmt.Sprintf("trial_%d_epoch_best_ages", i)
		expectedBestAges := tr.ChampionSpeciesAges()
		bestAges := Floats{}
		err = r.Read(key, &bestAges)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedBestAges, bestAges)

		key = fmt.Sprintf("trial_%d_epoch_best_complexities", i)
		expectedBestComplexities := tr.ChampionsComplexities()
		bestComplexities := Floats{}
		err = r.Read(key, &bestComplexities)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedBestComplexities, bestComplexities)

		key = fmt.Sprintf("trial_%d_epoch_diversity", i)
		expectedDiversity := tr.Diversity()
		diversity := Floats{}
		err = r.Read(key, &diversity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedDiversity, diversity)
	}
	err = r.Close()
	assert.NoError(t, err, "failed to close reader")
}

func TestExperiment_WriteNPZ_writeError(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test Encode Decode", Trials: make(Trials, 3)}
	for i := 0; i < len(ex.Trials); i++ {
		ex.Trials[i] = *buildTestTrial(i+1, 10)
	}

	errWriter := ErrorWriter(1)
	err := ex.Write(&errWriter)
	assert.EqualError(t, err, alwaysErrorText)
}

func TestExperiment_AvgTrialDuration(t *testing.T) {
	trials := Trials{
		Trial{Duration: time.Duration(3)},
		Trial{Duration: time.Duration(10)},
		Trial{Duration: time.Duration(2)},
	}
	ex := Experiment{Id: 1, Name: "Test AvgTrialDuration", Trials: trials}
	duration := ex.AvgTrialDuration()
	assert.Equal(t, time.Duration(5), duration)
}

func TestExperiment_AvgTrialDuration_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test AvgTrialDuration_emptyTrials", Trials: Trials{}}
	duration := ex.AvgTrialDuration()
	assert.Equal(t, EmptyDuration, duration)
}

func TestExperiment_AvgEpochDuration(t *testing.T) {
	durations := [][]time.Duration{
		{time.Duration(3), time.Duration(10), time.Duration(2)},
		{time.Duration(1), time.Duration(1), time.Duration(1)},
	}
	trials := Trials{
		*buildTestTrialWithGenerationsDuration(durations[0]),
		*buildTestTrialWithGenerationsDuration(durations[1]),
	}
	ex := Experiment{Id: 1, Name: "Test AvgEpochDuration", Trials: trials}
	duration := ex.AvgEpochDuration()
	assert.Equal(t, time.Duration(3), duration)
}

func TestExperiment_AvgEpochDuration_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test AvgEpochDuration_emptyTrials", Trials: Trials{}}
	duration := ex.AvgEpochDuration()
	assert.Equal(t, EmptyDuration, duration)
}

func TestExperiment_AvgGenerationsPerTrial(t *testing.T) {
	numGenerations := []int{5, 8, 6, 1}
	trials := Trials{
		*buildTestTrial(0, numGenerations[0]),
		*buildTestTrial(1, numGenerations[1]),
		*buildTestTrial(2, numGenerations[2]),
		*buildTestTrial(3, numGenerations[3]),
	}
	ex := Experiment{Id: 1, Name: "Test AvgGenerationsPerTrial", Trials: trials}
	gens := ex.AvgGenerationsPerTrial()
	assert.Equal(t, 5.0, gens)
}

func TestExperiment_AvgGenerationsPerTrial_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test AvgGenerationsPerTrial_emptyTrials", Trials: Trials{}}
	gens := ex.AvgGenerationsPerTrial()
	assert.Equal(t, 0.0, gens)
}

func TestExperiment_MostRecentTrialEvalTime(t *testing.T) {
	now := time.Now()
	trials := Trials{
		Trial{
			Generations: Generations{Generation{Executed: now}},
		},
		Trial{
			Generations: Generations{Generation{Executed: now.Add(time.Duration(-1))}},
		},
		Trial{
			Generations: Generations{Generation{Executed: now.Add(time.Duration(-2))}},
		},
	}
	ex := Experiment{Id: 1, Name: "Test MostRecentTrialEvalTime", Trials: trials}
	mostRecent := ex.MostRecentTrialEvalTime()
	assert.Equal(t, now, mostRecent)
}

func TestExperiment_MostRecentTrialEvalTime_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test MostRecentTrialEvalTime_emptyTrials", Trials: Trials{}}
	mostRecent := ex.MostRecentTrialEvalTime()
	assert.Equal(t, time.Time{}, mostRecent)
}

func TestExperiment_BestOrganism(t *testing.T) {
	fitnessMultipliers := Floats{1.0, 2.0, 3.0}
	trials := make(Trials, len(fitnessMultipliers))
	for i, fm := range fitnessMultipliers {
		trials[i] = *buildTestTrialWithFitnessMultiplier(i, i+2, fm)
	}
	ex := Experiment{Id: 1, Name: "Test BestOrganism", Trials: trials}
	bestOrg, trialId, ok := ex.BestOrganism(true)
	assert.True(t, ok)
	// the last trial
	assert.Equal(t, 2, trialId)
	// the best organism of last generation of last trial
	assert.Equal(t, fitnessScore(2+2)*fitnessMultipliers[2], bestOrg.Fitness)
}

func TestExperiment_BestOrganism_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test BestOrganism_emptyTrials", Trials: Trials{}}
	bestOrg, trialId, ok := ex.BestOrganism(true)
	assert.False(t, ok)
	assert.Equal(t, -1, trialId)
	assert.Nil(t, bestOrg)
}

func TestExperiment_Solved(t *testing.T) {
	trials := Trials{
		*buildTestTrial(1, 2),
		*buildTestTrial(2, 3),
		*buildTestTrial(3, 5),
	}
	ex := Experiment{Id: 1, Name: "Test Solved", Trials: trials}
	solved := ex.Solved()
	assert.True(t, solved)
}

func TestExperiment_Solved_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test Solved_emptyTrials", Trials: Trials{}}
	solved := ex.Solved()
	assert.False(t, solved)
}

func TestExperiment_BestFitness(t *testing.T) {
	fitnessMultipliers := Floats{1.0, 2.0, 3.0}
	trials := make(Trials, len(fitnessMultipliers))
	expectedFitness := make(Floats, len(fitnessMultipliers))
	for i, fm := range fitnessMultipliers {
		trials[i] = *buildTestTrialWithFitnessMultiplier(i, i+2, fm)
		expectedFitness[i] = fitnessScore(i+2) * fm
	}
	ex := Experiment{Id: 1, Name: "Test ChampionFitness", Trials: trials}
	bestFitness := ex.BestFitness()
	assert.Equal(t, len(expectedFitness), len(bestFitness))
	assert.EqualValues(t, expectedFitness, bestFitness)
}

func TestExperiment_BestFitness_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test BestFitness_emptyTrials", Trials: Trials{}}
	bestFitness := ex.BestFitness()
	assert.Equal(t, 0, len(bestFitness))
}

func TestExperiment_BestSpeciesAge(t *testing.T) {
	trials := Trials{
		*buildTestTrial(10, 1),
		*buildTestTrial(20, 2),
		*buildTestTrial(30, 3),
	}
	// assign species to the best organisms
	expected := Floats{10, 15, 1}
	for i, t := range trials {
		if org, ok := t.BestOrganism(false); ok {
			org.Species = genetics.NewSpecies(i)
			org.Species.Age = int(expected[i])
		}
	}

	ex := Experiment{Id: 1, Name: "Test BestSpeciesAge", Trials: trials}
	bestAge := ex.BestSpeciesAge()
	assert.EqualValues(t, expected, bestAge)
}

func TestExperiment_BestSpeciesAge_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test BestSpeciesAge_emptyTrials", Trials: Trials{}}
	bestAge := ex.BestSpeciesAge()
	assert.Equal(t, 0, len(bestAge))
}

func TestExperiment_BestComplexity(t *testing.T) {
	trials := Trials{
		*buildTestTrialWithBestOrganismGenesis(1, 3),
		*buildTestTrialWithBestOrganismGenesis(2, 4),
		*buildTestTrialWithBestOrganismGenesis(3, 2),
	}
	ex := Experiment{Id: 1, Name: "Test BestComplexity", Trials: trials}
	bestComplexity := ex.BestComplexity()
	expected := Floats{7.0, 7.0, 7.0}
	assert.EqualValues(t, expected, bestComplexity)
}

func TestExperiment_BestComplexity_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test BestComplexity_emptyTrials", Trials: Trials{}}
	bestComplexity := ex.BestComplexity()
	assert.Equal(t, 0, len(bestComplexity))
}

func TestExperiment_AvgDiversity(t *testing.T) {
	trials := Trials{
		*buildTestTrial(1, 2),
		*buildTestTrial(1, 3),
		*buildTestTrial(1, 5),
	}
	ex := Experiment{Id: 1, Name: "Test AvgDiversity", Trials: trials}

	avgDiversity := ex.AvgDiversity()
	expected := Floats{testDiversity, testDiversity, testDiversity}
	assert.EqualValues(t, expected, avgDiversity)
}

func TestExperiment_AvgDiversity_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test AvgDiversity_emptyTrials", Trials: Trials{}}
	avgDiversity := ex.AvgDiversity()
	assert.Equal(t, 0, len(avgDiversity))
}

func TestExperiment_EpochsPerTrial(t *testing.T) {
	expected := Floats{2, 3, 5}
	trials := Trials{
		*buildTestTrial(1, int(expected[0])),
		*buildTestTrial(1, int(expected[1])),
		*buildTestTrial(1, int(expected[2])),
	}
	ex := Experiment{Id: 1, Name: "Test EpochsPerTrial", Trials: trials}

	epochs := ex.EpochsPerTrial()
	assert.EqualValues(t, expected, epochs)
}

func TestExperiment_EpochsPerTrial_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test EpochsPerTrial_emptyTrials", Trials: Trials{}}
	epochs := ex.EpochsPerTrial()
	assert.Equal(t, 0, len(epochs))
}

func TestExperiment_TrialsSolved(t *testing.T) {
	solvedExpected := 2
	trials := createTrialsWithNSolved([]int{2, 3, 5}, solvedExpected, false)

	ex := Experiment{Id: 1, Name: "Test TrialsSolved", Trials: trials}
	solved := ex.TrialsSolved()
	assert.Equal(t, solvedExpected, solved)
}

func TestExperiment_TrialsSolved_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test TrialsSolved_emptyTrials", Trials: Trials{}}
	solved := ex.TrialsSolved()
	assert.Equal(t, 0, solved)
}

func TestExperiment_SuccessRate(t *testing.T) {
	solvedExpected := 2
	trials := createTrialsWithNSolved([]int{2, 3, 5}, solvedExpected, false)

	ex := Experiment{Id: 1, Name: "Test SuccessRate", Trials: trials}
	successRate := ex.SuccessRate()
	expectedRate := float64(solvedExpected) / 3.0
	assert.Equal(t, expectedRate, successRate)
}

func TestExperiment_SuccessRate_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test TrialsSolved_emptyTrials", Trials: Trials{}}
	successRate := ex.SuccessRate()
	assert.Equal(t, 0.0, successRate)
}

func TestExperiment_AvgWinnerStatistics(t *testing.T) {
	trials := Trials{
		*buildTestTrial(1, 2),
		*buildTestTrial(1, 3),
		*buildTestTrial(1, 5),
	}
	ex := Experiment{Id: 1, Name: "Test AvgWinnerStatistics", Trials: trials}

	avgNodes, avgGenes, avgEvals, avgDiversity := ex.AvgWinnerStatistics()
	assert.EqualValues(t, testWinnerNodes, avgNodes)
	assert.EqualValues(t, testWinnerGenes, avgGenes)
	assert.EqualValues(t, testWinnerEvals, avgEvals)
	assert.EqualValues(t, testDiversity, avgDiversity)
}

func TestExperiment_AvgWinnerStatistics_not_solved(t *testing.T) {
	solved := 0
	trials := createTrialsWithNSolved([]int{2, 3, 5}, solved, false)
	ex := Experiment{Id: 1, Name: "Test AvgWinnerStatistics_not_solved", Trials: trials}
	avgNodes, avgGenes, avgEvals, avgDiversity := ex.AvgWinnerStatistics()
	assert.EqualValues(t, -1, avgNodes)
	assert.EqualValues(t, -1, avgGenes)
	assert.EqualValues(t, -1, avgEvals)
	assert.EqualValues(t, -1, avgDiversity)
}

func TestExperiment_AvgWinnerStatistics_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test AvgWinnerStatistics_emptyTrials", Trials: Trials{}}
	avgNodes, avgGenes, avgEvals, avgDiversity := ex.AvgWinnerStatistics()
	assert.EqualValues(t, -1, avgNodes)
	assert.EqualValues(t, -1, avgGenes)
	assert.EqualValues(t, -1, avgEvals)
	assert.EqualValues(t, -1, avgDiversity)
}

func TestExperiment_EfficiencyScore(t *testing.T) {
	solved := 2
	trials := createTrialsWithNSolved([]int{2, 3, 5}, solved, true)
	ex := Experiment{Id: 1, Name: "Test EfficiencyScore", Trials: trials}

	meanComplexity := 7.0
	meanFitness := math.E
	successRate := ex.SuccessRate()
	logPenaltyScore := math.Log(ex.penaltyScore(meanComplexity))
	score := successRate * meanFitness / logPenaltyScore

	actualScore := ex.EfficiencyScore()
	assert.EqualValues(t, score, actualScore)
}

func TestExperiment_EfficiencyScore_maxFitness(t *testing.T) {
	solved := 2
	trials := createTrialsWithNSolved([]int{2, 3, 5}, solved, true)
	ex := Experiment{Id: 1, Name: "Test EfficiencyScore_maxFitness", Trials: trials, MaxFitnessScore: math.E * 2.0}

	meanComplexity := 7.0
	meanFitness := math.E
	successRate := ex.SuccessRate()
	logPenaltyScore := math.Log(ex.penaltyScore(meanComplexity))
	meanFitness = (meanFitness / ex.MaxFitnessScore) * 100
	score := successRate * meanFitness / logPenaltyScore

	actualScore := ex.EfficiencyScore()
	assert.EqualValues(t, score, actualScore)
}

func TestExperiment_EfficiencyScore_emptyTrials(t *testing.T) {
	ex := Experiment{Id: 1, Name: "Test EfficiencyScore_emptyTrials", Trials: Trials{}}
	score := ex.EfficiencyScore()
	assert.EqualValues(t, 0, score)
}

func createTrialsWithNSolved(generations []int, solvedNumber int, enableGenesis bool) Trials {
	trials := make(Trials, len(generations))
	for i := range generations {
		if enableGenesis {
			trials[i] = *buildTestTrialWithBestOrganismGenesis(i, generations[i])
		} else {
			trials[i] = *buildTestTrial(i, generations[i])
		}
	}

	for _, trial := range trials {
		solved := solvedNumber > 0
		solvedNumber -= 1
		for j := range trial.Generations {
			trial.Generations[j].Solved = solved
		}
	}
	return trials
}
