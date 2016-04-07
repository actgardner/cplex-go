package model

import (
	"testing"
)

func TestEnvironment(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	err = env.Free()
	if err != nil {
		t.Fatalf("Got error freeing env: %v", err)
	}
}

func TestCreateProblem(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	defer env.Free()
	prob, err := env.CreateProblem("TestProblem")
	err = env.Free()
	if err != nil {
		t.Fatalf("Got error creating new problem: %v", err)
	}
	prob.Free()
}

func TestReadCopyProblem(t *testing.T) {
	env, err := NewEnvironment()
	if err != nil {
		t.Fatalf("Got error creating env: %v", err)
	}
	defer env.Free()
	prob, err := env.ReadCopyProblem("TestProblem", "../diet.mps", ReadAutoDetect)
	if err != nil {
		t.Fatalf("Got error reading problem: %v", err)
	}
	prob.Free()
}
