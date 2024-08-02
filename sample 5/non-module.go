package main

import (
    "fmt"
    "log"
    "os"
    "unsafe"

    "github.com/tetratelabs/wazero/v2"
)

// Helper function to allocate memory and return the pointer
func allocateMemory(module wazero.Module, size int) uint32 {
    memAlloc, err := module.ExportedFunction("allocate_memory").Call(uint64(size))
    if err != nil {
        log.Fatal(err)
    }
    return uint32(memAlloc[0])
}

// Helper function to write a string to allocated memory
func writeStringToMemory(module wazero.Module, ptr uint32, str string) {
    bytes := []byte(str)
    if len(bytes) > 100 { // Adjust size if needed
        log.Fatal("String is too long")
    }
    mem, err := module.Memory().Read(ptr, uint32(len(bytes)))
    if err != nil {
        log.Fatal(err)
    }
    copy(mem, bytes)
}

// Helper function to deallocate memory
func deallocateMemory(module wazero.Module, ptr uint32, size int) {
    _, err := module.ExportedFunction("deallocate_memory").Call(uint64(ptr), uint64(size))
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    // Initialize the wazero runtime
    runtime := wazero.NewRuntime()
    defer runtime.Close()

    // Load the WebAssembly module
    wasmBytes, err := os.ReadFile("path/to/your/rust_wasm.wasm")
    if err != nil {
        log.Fatal(err)
    }
    
    module, err := runtime.InstantiateModule(wasmBytes)
    if err != nil {
        log.Fatal(err)
    }

    // Allocate memory for the first string
    ptr1 := allocateMemory(module, 100)

    // Prepare and write the first string
    writeStringToMemory(module, ptr1, "Hello, ")

    // Allocate memory for the second string
    ptr2 := allocateMemory(module, 100)

    // Prepare and write the second string
    writeStringToMemory(module, ptr2, "World!")

    // Call the concat_strings function
    concatFunc, err := module.ExportedFunction("concat_strings").Call(uint64(ptr1), 7, uint64(ptr2), 6)
    if err != nil {
        log.Fatal(err)
    }

    resultPtr := concatFunc[0]

    // Read the concatenated result
    result := make([]byte, 1024) // Adjust size as necessary
    _, err = module.Memory().Read(resultPtr, 1024)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Concatenated string from Rust:", string(result))

    // Deallocate memory for both strings
    deallocateMemory(module, ptr1, 100)
    deallocateMemory(module, ptr2, 100)
}
