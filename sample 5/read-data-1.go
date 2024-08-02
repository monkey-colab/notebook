package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "io/ioutil"
    "log"

    "github.com/bytecodealliance/wasmtime-go"
)

// readBuffer reads the query results buffer from the WebAssembly module.
func readBuffer(instance *wasmtime.Instance, bufferPtr, bufferLen int32) []byte {
    memory := instance.GetExport("memory").Memory()
    return memory.UnsafeData()[bufferPtr:bufferPtr+bufferLen]
}

// parseResults parses the binary buffer to extract variable names and values.
func parseResults(buffer []byte) {
    reader := bytes.NewReader(buffer)

    for reader.Len() > 0 {
        var resultLen uint32
        err := binary.Read(reader, binary.LittleEndian, &resultLen)
        if err != nil {
            log.Fatal(err)
        }

        endPos := int(resultLen)
        if endPos > reader.Len() {
            log.Fatalf("Result length exceeds buffer length")
        }

        for reader.Len() > endPos {
            var varLen uint32
            err := binary.Read(reader, binary.LittleEndian, &varLen)
            if err != nil {
                log.Fatal(err)
            }

            varName := make([]byte, varLen)
            _, err = reader.Read(varName)
            if err != nil {
                log.Fatal(err)
            }

            var valueLen uint32
            err = binary.Read(reader, binary.LittleEndian, &valueLen)
            if err != nil {
                log.Fatal(err)
            }

            value := make([]byte, valueLen)
            _, err = reader.Read(value)
            if err != nil {
                log.Fatal(err)
            }

            fmt.Printf("Var: %s, Value: %s\n", string(varName), string(value))
        }
    }
}

func main() {
    // Load the WebAssembly module
    wasm, err := ioutil.ReadFile("query_results.wasm")
    if err != nil {
        log.Fatal(err)
    }

    // Create a new Wasmtime engine and store
    engine := wasmtime.NewEngine()
    store := wasmtime.NewStore(engine)

    // Compile the module
    module, err := wasmtime.NewModule(engine, wasm)
    if err != nil {
        log.Fatal(err)
    }

    // Instantiate the module
    linker := wasmtime.NewLinker(store)
    instance, err := linker.Instantiate(module)
    if err != nil {
        log.Fatal(err)
    }

    // Call the create_query_results_buffer function
    createBuffer := instance.GetExport("create_query_results_buffer").Func().Call
    res, err := createBuffer()
    if err != nil {
        log.Fatal(err)
    }
    bufferPtr := res.(int32)

    // Call the get_query_results_buffer_len function
    getBufferLen := instance.GetExport("get_query_results_buffer_len").Func().Call
    res, err = getBufferLen(bufferPtr)
    if err != nil {
        log.Fatal(err)
    }
    bufferLen := res.(int32)

    // Call the get_query_results_buffer_ptr function
    getBufferPtr := instance.GetExport("get_query_results_buffer_ptr").Func().Call
    res, err = getBufferPtr(bufferPtr)
    if err != nil {
        log.Fatal(err)
    }
    bufferPtr = res.(int32)

    // Read the buffer
    buffer := readBuffer(instance, bufferPtr, bufferLen)

    // Parse the results
    parseResults(buffer)

    // Free the buffer
    freeBuffer := instance.GetExport("free_query_results_buffer").Func().Call
    _, err = freeBuffer(bufferPtr)
    if err != nil {
        log.Fatal(err)
    }
}
