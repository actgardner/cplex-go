package model

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

/* Read in a sample problem and check the row and column counts */
func TestReadProblemStats(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	defer env.Free()
	prob, err := env.ReadCopyProblem("TestProblem", "../diet.mps", ReadAutoDetect)
	if err != nil {
		t.Fatalf("Got error reading problem: %v", err)
	}
	numRows := prob.NumRows()
	if numRows != 8 {
		t.Fatal("Got %v rows, expected 8", numRows)
	}
	numCols := prob.NumColumns()
	if numCols != 4 {
		t.Fatal("Got %v columns, expected 4", numCols)
	}
	prob.Free()
}

/* Read in a sample problem, write it back out in a different format, then check the new file's row and column counts */
func TestWriteProblem(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	defer env.Free()
	prob, err := env.ReadCopyProblem("TestProblem", "../diet.mps", ReadAutoDetect)
	if err != nil {
		t.Fatalf("Got error reading problem: %v", err)
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	tempProblemFile := f.Name()
	defer os.Remove(tempProblemFile)
	err = prob.WriteProblem(tempProblemFile, WriteSAVProblemFile)
	if err != nil {
		t.Fatalf("Got error reading problem: %v", err)
	}
	prob.Free()

	prob, err = env.ReadCopyProblem("TempProblem", tempProblemFile, ReadSAVProblemFile)
	if err != nil {
		t.Fatalf("Got error reading written problem: %v", err)
	}
	numRows := prob.NumRows()
	if numRows != 8 {
		t.Fatal("Got %v rows, expected 8", numRows)
	}
	numCols := prob.NumColumns()
	if numCols != 4 {
		t.Fatal("Got %v columns, expected 4", numCols)
	}
	prob.Free()
}

func TestInfeasibility(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	defer env.Free()
	prob, err := env.ReadCopyProblem("TestProblem", "../diet.mps", ReadAutoDetect)
	if err != nil {
		t.Fatalf("Got error reading problem: %v", err)
	}
	infeas, err := prob.GetRowInfeasibility([]float64{1, 1, 1, 1}, 0, 3)
	if err != nil {
		t.Fatalf("Got error getting infeasability: %v", err)
	}
	expectedInfeas := []float64{-584, -620, -645}
	if !reflect.DeepEqaul(expectedInfeas, infeas) {
		t.Fatalf("Got infeasibility %v, expected %v", infeas, expectedInfeas)
	}
	prob.Free()
}

/* Solve a sample problem and read the solution */
func TestLPOptimize(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	defer env.Free()
	prob, err := env.ReadCopyProblem("TestProblem", "../diet.mps", ReadAutoDetect)
	if err != nil {
		t.Fatalf("Got error reading problem: %v", err)
	}

	err = prob.LPOptimize()
	if err != nil {
		t.Fatalf("Got error optimizing problem: %v", err)
	}

	solution, err := prob.GetSolution()
	if err != nil {
		t.Fatalf("Got error reading problem solution: %v", err)
	}

	if solution.Status != CPX_STAT_OPTIMAL {
		t.Fatalf("Got status %v, expected CPX_STAT_OPTIMAL", solution.Status)
	}
	if solution.ObjectiveValue != 88.2 {
		t.Fatalf("Got objective value %v, expected 88.2", solution.ObjectiveValue)
	}
	if solution.StatusString() != "optimal" {
		t.Fatalf("Got status string %q, expected %q", solution.StatusString(), "optimal")
	}

	expectedX := []float64{0, 0, 0, 0}
	expectedPi := []float64{-0, -0, 0.126, -0, 0, 0, 0, 0}
	expectedSlack := []float64{0, 0, 0, 933.3333333333333, 0, 0, 0, 0}
	expectedDj := []float64{1.2999999999999998, 0.06999999999999984, 1.03, 1.6300000000000001}

	if !reflect.DeepEqual(solution.X, expectedX) {
		t.Fatalf("Got solution X %v, expected %v", solution.X, expectedX)
	}
	if !reflect.DeepEqual(solution.Pi, expectedPi) {
		t.Fatalf("Got solution Pi %v, expected %v", solution.Pi, expectedPi)
	}
	if !reflect.DeepEqual(solution.Slack, expectedSlack) {
		t.Fatalf("Got solution Slack %v, expected %v", solution.Slack, expectedSlack)
	}
	if !reflect.DeepEqual(solution.Dj, expectedDj) {
		t.Fatalf("Got solution Dj %v, expected %v", solution.Dj, expectedDj)
	}

	iterationCount := prob.GetIterationCount()
	if iterationCount != 1 {
		t.Fatalf("Got iteration count %v, expected 1", iterationCount)
	}
	prob.Free()
}
