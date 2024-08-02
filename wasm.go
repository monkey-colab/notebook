package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"unsafe"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMManager manages the lifecycle and interaction with a WASM module.
type WASMManager struct {
	context      context.Context
	runtime      wazero.Runtime
	module       api.Module
	allocateFn   api.Function
	deallocateFn api.Function
	memory       api.Memory
	bufferField  reflect.Value
}

var (
	wasmManagerInstance *WASMManager
	once                sync.Once
	wasmInitError       error
)

// NewWASMManager initializes a new WASMManager singleton with the provided WASM binary.
func NewWASMManager(wasmBinary []byte) (*WASMManager, error) {
	once.Do(func() {
		ctx := context.Background()
		runtime := wazero.NewRuntime(ctx)

		// Creating and instantiating the host module for logging.
		_, err := runtime.NewHostModuleBuilder("env").
			NewFunctionBuilder().WithFunc(logString).Export("log").
			Instantiate(ctx)
		if err != nil {
			wasmInitError = err
			return
		}

		// Instantiating the WASI module.
		wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

		// Instantiating the provided WASM binary.
		module, err := runtime.Instantiate(ctx, wasmBinary)
		if err != nil {
			wasmInitError = err
			return
		}

		// Accessing the module memory.
		memory := module.Memory()
		if memory == nil {
			wasmInitError = errors.New("failed to access module memory")
			return
		}

		// HACK: Using reflection to access the private Buffer field.
		rv := reflect.ValueOf(memory).Elem()
		fv := rv.FieldByName("Buffer")
		if !fv.IsValid() {
			wasmInitError = errors.New("buffer field not found")
			return
		}

		// Exporting allocate and deallocate functions from the WASM module.
		allocateFn := module.ExportedFunction("allocate_memory")
		if allocateFn == nil {
			wasmInitError = errors.New("unable to find allocate_memory function")
			return
		}

		deallocateFn := module.ExportedFunction("deallocate_memory")
		if deallocateFn == nil {
			wasmInitError = errors.New("unable to find deallocate_memory function")
			return
		}

		// Initializing the singleton instance of WASMManager.
		wasmManagerInstance = &WASMManager{
			context:      ctx,
			runtime:      runtime,
			module:       module,
			allocateFn:   allocateFn,
			deallocateFn: deallocateFn,
			memory:       memory,
			bufferField:  fv,
		}
	})

	if wasmInitError != nil {
		return nil, wasmInitError
	}
	return wasmManagerInstance, nil
}

// getWASMManagerInstance returns the initialized WASMManager instance.
func getWASMManagerInstance() *WASMManager {
	if wasmManagerInstance == nil {
		panic("WASMManager is not initialized")
	}
	return wasmManagerInstance
}

// Memory represents a block of allocated memory in WASM.
type Memory struct {
	data uintptr
	size int
}

// NewMemory allocates a new block of memory in WASM and returns a pointer to it.
func NewMemory(size int) (*Memory, error) {
	manager := getWASMManagerInstance()

	results, err := manager.allocateFn.Call(manager.context, uint64(size))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory: %w", err)
	}
	if len(results) == 0 {
		return nil, errors.New("allocate function did not return a result")
	}

	data := uintptr(results[0])
	return &Memory{
		data: data,
		size: size,
	}, nil
}

// Free deallocates a previously allocated block of memory in WASM.
func (m *Memory) Free() error {
	manager := getWASMManagerInstance()

	_, err := manager.deallocateFn.Call(manager.context, uint64(m.data), uint64(m.size))
	if err != nil {
		return fmt.Errorf("failed to deallocate memory: %w", err)
	}
	return nil
}

// Write writes data into the allocated memory block in WASM.
func (m *Memory) Write(data []byte) error {
	manager := getWASMManagerInstance()

	size := len(data)
	if size > m.size {
		return errors.New("data size is larger than the allocated memory")
	}

	if !manager.memory.Write(uint32(m.data), data) {
		return errors.New("failed to write to memory")
	}
	return nil
}

// Read reads data from the allocated memory block in WASM.
func (m *Memory) Read() ([]byte, error) {
	manager := getWASMManagerInstance()

	data, read := manager.memory.Read(uint32(m.data), uint32(m.size))
	if !read {
		return nil, errors.New("failed to read memory")
	}
	return data, nil
}

// ToByte converts the memory block to a byte slice.
func (m *Memory) ToByte() []byte {
	manager := getWASMManagerInstance()

	// Access the whole memory buffer as a byte slice.
	bufLen := manager.memory.Size()
	buf := unsafe.Slice((*byte)(unsafe.Pointer(manager.bufferField.UnsafeAddr())), bufLen)

	// Slice the buffer to get the portion we are interested in.
	memSlice := buf[m.data : m.data+uintptr(m.size)]

	return memSlice

	// Safely access the underlying memory buffer, and create a slice from the buffer.
	//bufferPointer := *(**uint8)(unsafe.Pointer(manager.bufferField.UnsafeAddr()))
	//return unsafe.Slice((*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(bufferPointer))+m.data)), m.size)
}

// Param is an interface for parameters to be passed to a WASM function.
type Param interface {
	getValues() []uint64
}

// SimpleParam represents a simple param to be passed to a WASM function.
type SimpleParam struct {
	val uint64
}

// NewSimpleParam creates a new SimpleParam.
func NewSimpleParam(val uint64) *SimpleParam {
	return &SimpleParam{val}
}

// getValues returns the underlying value of SimpleParam.
func (p *SimpleParam) getValues() []uint64 {
	return []uint64{p.val}
}

// MemoryParam represents a memory param to be passed to a WASM function.
type MemoryParam struct {
	Memory
}

// getValues returns the memory address and size.
func (p *MemoryParam) getValues() []uint64 {
	return []uint64{uint64(p.data), uint64(p.size)}
}

// NewMemoryParam creates a new MemoryParam and writes the given data to it.
func NewMemoryParam(data []byte, size int) (*MemoryParam, error) {
	mem, err := NewMemory(size)
	if err != nil {
		return nil, fmt.Errorf("failed to make param: %w", err)
	}

	err = mem.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to set param data: %w", err)
	}

	return &MemoryParam{Memory: *mem}, nil
}

// StringParam represents a string param to be passed to WASM function.
type StringParam struct {
	MemoryParam
}

// NewStringParam creates a new StringParam from a string.
func NewStringParam(value string) (*StringParam, error) {
	p, err := NewMemoryParam([]byte(value), len(value))
	if err != nil {
		return nil, err
	}
	return &StringParam{MemoryParam: *p}, nil
}

// Result represents the result from the WASM module.
type Result struct {
	Memory
}

// NewResult creates a new Result object from a pointer.
func NewResult(ptr uint64) *Result {
	uptr := uint32(ptr >> 32)
	size := uint32(ptr)
	return &Result{
		Memory{
			data: uintptr(uptr),
			size: int(size),
		},
	}
}

// Function represents a callable function in the WASM module.
type Function struct {
	fn api.Function
}

// NewFunction creates a new Function object by looking up a function by name in the WASM module.
func NewFunction(name string) (*Function, error) {
	manager := getWASMManagerInstance()

	fn := manager.module.ExportedFunction(name)
	if fn == nil {
		return nil, errors.New("unable to find function")
	}

	return &Function{fn}, nil
}

// CallWithResult calls the function with the provided parameters and returns the result.
func (f *Function) CallWithResult(params []Param) (*Result, error) {
	rval, err := f.Call(params)
	if err != nil {
		return nil, fmt.Errorf("failed to call function: %w", err)
	}

	if rval == 0 {
		return nil, errors.New("memory function returned null")
	}

	return NewResult(rval), nil
}

// Call calls the function with the provided parameters and returns the raw result.
func (f *Function) Call(params []Param) (uint64, error) {
	manager := getWASMManagerInstance()

	var stack []uint64
	for _, p := range params {
		stack = append(stack, p.getValues()...)
	}

	err := f.fn.CallWithStack(manager.context, stack)
	if err != nil {
		return 0, fmt.Errorf("failed to call function: %w", err)
	}

	return stack[0], nil
}

// logString logs a string from the WASM module memory to the console.
func logString(ctx context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", offset, byteCount)
	}
	fmt.Println(string(buf))
}
