package wasm_module

import (
    "log"
    "github.com/tetratelabs/wazero/v2"
)

// WasmClient holds the runtime and module for WebAssembly interactions
type WasmClient struct {
    runtime wazero.Runtime
    module  wazero.Module
}

// NewWasmClient initializes a new WasmClient with the given WebAssembly module bytes
func NewWasmClient(wasmBytes []byte) (*WasmClient, error) {
    runtime := wazero.NewRuntime()
    module, err := runtime.InstantiateModule(wasmBytes)
    if err != nil {
        return nil, err
    }
    return &WasmClient{
        runtime: runtime,
        module:  module,
    }, nil
}

// AllocateMemory allocates memory and returns the pointer
func (c *WasmClient) AllocateMemory(size int) uint32 {
    memAlloc, err := c.module.ExportedFunction("allocate_memory").Call(uint64(size))
    if err != nil {
        log.Fatal(err)
    }
    return uint32(memAlloc[0])
}

// WriteStringToMemory writes a string to the allocated memory
func (c *WasmClient) WriteStringToMemory(ptr uint32, str string) {
    bytes := []byte(str)
    if len(bytes) > 100 { // Adjust size if needed
        log.Fatal("String is too long")
    }
    mem, err := c.module.Memory().Read(ptr, uint32(len(bytes)))
    if err != nil {
        log.Fatal(err)
    }
    copy(mem, bytes)
}

// DeallocateMemory deallocates memory
func (c *WasmClient) DeallocateMemory(ptr uint32, size int) {
    _, err := c.module.ExportedFunction("deallocate_memory").Call(uint64(ptr), uint64(size))
    if err != nil {
        log.Fatal(err)
    }
}

// CallConcatStrings calls the concat_strings function from the WebAssembly module
func (c *WasmClient) CallConcatStrings(ptr1 uint32, len1 int, ptr2 uint32, len2 int) string {
    concatFunc, err := c.module.ExportedFunction("concat_strings").Call(uint64(ptr1), uint64(len1), uint64(ptr2), uint64(len2))
    if err != nil {
        log.Fatal(err)
    }
    resultPtr := concatFunc[0]
    result := make([]byte, 1024) // Adjust size as necessary
    _, err = c.module.Memory().Read(resultPtr, 1024)
    if err != nil {
        log.Fatal(err)
    }
    return string(result)
}

// Close releases the resources used by the WasmClient
func (c *WasmClient) Close() {
    c.runtime.Close()
}
