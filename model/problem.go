package model

/*
#cgo CFLAGS: -IC:/Cplex/cplex/include
#cgo LDFLAGS: -lcplex1263 -LC:/Cplex/cplex/lib/x64_windows_vs2013/stat_mda

#define _LP64
#include <ilcplex/cplex.h>

*/
import "C"

import (
	"unsafe"
)

/* Types of problem files which can be read in - AutoDetect attempts to choose a type automatically based on extension and file contents */
type ProblemFileReadType string

const (
	ReadAutoDetect     ProblemFileReadType = ""
	ReadSAVProblemFile ProblemFileReadType = "SAV"
	ReadMPSProblemFile ProblemFileReadType = "MPS"
	ReadLPProblemFile  ProblemFileReadType = "LP"
)

/* Types of problem files which can be written - AutoDetect attempts to choose a type automatically based on file extension */
type ProblemFileWriteType string

const (
	WriteAutoDetect     ProblemFileWriteType = ""
	WriteSAVProblemFile ProblemFileWriteType = "SAV"
	WriteMPSProblemFile ProblemFileWriteType = "MPS"
	WriteLPProblemFile  ProblemFileWriteType = "LP"
	WriteREWProblemFile ProblemFileWriteType = "REW"
	WriteRLPProblemFile ProblemFileWriteType = "RLP"
	WriteALPProblemFile ProblemFileWriteType = "ALP"
)

type Problem struct {
	envPtr C.CPXENVptr
	ptr    C.CPXLPptr
}

/* Read a problem file on disk into this Problem */
func (p *Problem) ReadCopyProblem(path string, problemType ProblemFileReadType) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cType *C.char
	if problemType != ReadAutoDetect {
		cType = C.CString(string(problemType))
		defer C.free(unsafe.Pointer(cType))
	}
	err := C.CPXreadcopyprob(p.envPtr, p.ptr, cPath, cType)
	if err != 0 {
		return getCplexError(err)
	}
	return nil
}

/* Write this problem file to disk in the given format */
func (p *Problem) WriteProblem(path string, problemType ProblemFileWriteType) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var cType *C.char
	if problemType != WriteAutoDetect {
		cType = C.CString(string(problemType))
		defer C.free(unsafe.Pointer(cType))
	}
	err := C.CPXwriteprob(p.envPtr, p.ptr, cPath, cType)
	if err != 0 {
		return getCplexError(err)
	}
	return nil
}

/* Get the number of rows (constraints) in the Problem */
func (p *Problem) NumRows() int {
	return int(C.CPXgetnumcols(p.envPtr, p.ptr))
}

/* Get the number of columns (variables) in the Problem */
func (p *Problem) NumColumns() int {
	return int(C.CPXgetnumrows(p.envPtr, p.ptr))
}

/* Attempt to solve this Problem, returns any error that occurs. Note that exceeding a CPLEX limit or proving the problem infeasible or unbounded doesn't result in an error - see documentation for CPXlpopt for details */
func (p *Problem) LPOptimize() error {
	err := C.CPXlpopt(p.envPtr, p.ptr)
	if err != 0 {
		return getCplexError(err)
	}
	return nil
}

/* After calling LPOptimize, get the Solution for the Problem */
func (p *Problem) GetSolution() (*Solution, error) {
	numRows := p.NumRows()
	numCols := p.NumColumns()
	var lpstat C.int
	var objval C.double
	var x = make([]float64, numCols)
	var pi = make([]float64, numRows)
	var slack = make([]float64, numRows)
	var dj = make([]float64, numCols)

	err := C.CPXsolution(p.envPtr, p.ptr, &lpstat, &objval, (*C.double)(unsafe.Pointer(&x[0])), (*C.double)(unsafe.Pointer(&pi[0])), (*C.double)(unsafe.Pointer(&slack[0])), (*C.double)(unsafe.Pointer(&dj[0])))
	if err != 0 {
		return nil, getCplexError(err)
	}

	sol := &Solution{
		envPtr:         p.envPtr,
		Status:         SolutionStatus(lpstat),
		ObjectiveValue: float64(objval),
		X:              x,
		Pi:             pi,
		Slack:          slack,
		Dj:             dj,
	}

	return sol, nil
}

/* Get the infeasibility of a solution for a set of constraints - x is the solution, if x is nil the last computed solution will be used. See the documentation for CPXgetrowinfeas for details of how to interpret the output */
func (p *Problem) GetRowInfeasibility(x []float64, begin, end int) ([]float64, error) {
	var solution *C.double
	if x != nil {
		solution = (*C.double)(unsafe.Pointer(&x[0]))
	}
	infeasOut := make([]float64, end-begin)
	err := C.CPXgetrowinfeas(p.envPtr, p.ptr, solution, (*C.double)(unsafe.Pointer(&infeasOut[0])), C.int(begin), C.int(end))
	if err != 0 {
		return nil, getCplexError(err)
	}
	return infeasOut, nil
}

/* Get the number of simplex iterations required to find a solution */
func (p *Problem) GetIterationCount() int {
	return int(C.CPXgetitcnt(p.envPtr, p.ptr))
}

/* Set a single coefficient in the Problem, at row index i and column index j, to newValue */
func (p *Problem) ChangeCoefficient(i, j int, newValue float64) error {
	err := C.CPXchgcoef(p.envPtr, p.ptr, C.int(i), C.int(j), C.double(newValue))
	if err != 0 {
		return getCplexError(err)
	}
	return nil
}

type ObjectiveSense C.int

const (
	ObjectiveSenseMin ObjectiveSense = -1
	ObjectiveSenseMax ObjectiveSense = 1
)

type ConstraintSense C.char

const (
	ConstraintSenseLessThan    ConstraintSense = 'L'
	ConstraintSenseEqualTo     ConstraintSense = 'E'
	ConstraintSenseGreaterThan ConstraintSense = 'G'
	ConstraintSenseRanged      ConstraintSense = 'R'
)

/* Set all of the fields in this Problem - see CPXcopylpwnames in the CPLEX documentation for details */
func (p *Problem) CopyLPWNames(numCols, numRows int, objSense ObjectiveSense, objective, rhs []float64, sense []ConstraintSense, matbeg, matind, matcnt []int, matval, lb, ub, rngval []float64, colName, rowName []string) error {
	cColName := make([]*C.char, len(colName))
	for i, str := range colName {
		cColName[i] = C.CString(str)
		defer C.free(str)
	}
	cRowName := make([]*C.char, len(rowName))
	for i, str := range rowName {
		cRowName[i] = C.CString(str)
		defer C.free(str)
	}
	err := C.CPXcopylpwnames(p.envPtr, p.ptr, C.int(numCols), C.int(numRows), C.int(objSense), (*C.double)(unsafe.Pointer(&objective[0])), (*C.double)(unsafe.Pointer(&rhs[0])), (*C.char)(unsafe.Pointer(&sense[0])), (*C.int)(unsafe.Pointer(&matbeg[0])), (*C.int)(unsafe.Pointer(&matcnt[0])), (*C.int)(unsafe.Pointer(&matind[0])), (*C.double)(unsafe.Pointer(&matval[0])), (*C.double)(unsafe.Pointer(&lb[0])), (*C.double)(unsafe.Pointer(&ub[0])), (*C.double)(unsafe.Pointer(&rngval[0])), (**C.char)(unsafe.Pointer(&cColName[0])), (**C.char)(unsafe.Pointer(&cRowName[0])))
	if err != 0 {
		return getCplexError(err)
	}
	return nil
}

/* Free the memory associated with the Problem */
func (p *Problem) Free() {
	C.CPXfreeprob(p.envPtr, &p.ptr)
}
