package model

/*
#cgo CFLAGS: -IC:/Cplex/cplex/include
#cgo LDFLAGS: -lcplex1263 -LC:/Cplex/cplex/lib/x64_windows_vs2013/stat_mda

#define _LP64
#include <ilcplex/cplex.h>

*/
import "C"

import (
	"fmt"
	"unsafe"
)

type Environment struct {
	ptr C.CPXENVptr
}

/* Create a new CPLEX environment - the Environment must be Free'd when no longer needed */
func NewEnvironment() (*Environment, error) {
	var err C.int
	envPtr := C.CPXopenCPLEX(&err)
	if err != 0 {
		return nil, getCplexError(err)
	}
	return &Environment{envPtr}, nil
}

func getCplexError(err C.int) error {
	errString := make([]C.char, C.CPXMESSAGEBUFSIZE)
	C.CPXgeterrorstring(unsafe.Pointer(uintptr(0)), err, &errString[0])
	return fmt.Errorf(C.GoString(&errString[0]))
}

/* Create a new Problem with the given name */
func (e *Environment) CreateProblem(name string) (*Problem, error) {
	var err C.int
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	problemPtr := C.CPXcreateprob(e.ptr, &err, cName)
	if err != 0 {
		return nil, getCplexError(err)
	}
	return &Problem{e.ptr, problemPtr}, nil
}

/* Create a new Problem and read a problem file on disk into it */
func (e *Environment) ReadCopyProblem(name, path string, problemType ProblemFileReadType) (*Problem, error) {
	problem, err := e.CreateProblem(name)
	if err != nil {
		return nil, err
	}
	err = problem.ReadCopyProblem(path, problemType)
	if err != nil {
		problem.Free()
		return nil, err
	}
	return problem, nil
}

/* Free the memory for this Environment */
func (e *Environment) Free() error {
	err := C.CPXcloseCPLEX(&e.ptr)
	if err != 0 {
		return getCplexError(err)
	}
	return nil
}
