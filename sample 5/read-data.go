package main

import (
    "encoding/binary"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/tetratelabs/wazero"
)

// Result holds column names and rows of data
type Result struct {
    ColumnNames []string
    Rows        []*Row
}

// Row represents a single row in the Result
type Row struct {
    Values []string
    parent *Result
}

// NewResult initializes a Result from raw memory data
func NewResult(memory []byte) (*Result, error) {
    offset := 0

    // Read number of columns
    numColumns := int(memory[offset])
    offset++

    // Read column names
    columnNames := make([]string, numColumns)
    for i := 0; i < numColumns; i++ {
        nameLen := int(memory[offset])
        offset++
        columnNames[i] = string(memory[offset : offset+nameLen])
        offset += nameLen
    }

    // Read number of rows
    numRows := int(binary.LittleEndian.Uint32(memory[offset : offset+4]))
    offset += 4

    // Read rows
    rows := make([]*Row, numRows)
    for i := 0; i < numRows; i++ {
        values := make([]string, numColumns)
        for j := 0; j < numColumns; j++ {
            valueLen := int(memory[offset])
            offset++
            values[j] = string(memory[offset : offset+valueLen])
            offset += valueLen
        }
        rows[i] = &Row{
            Values: values,
            parent: &Result{ColumnNames: columnNames},
        }
    }

    return &Result{
        ColumnNames: columnNames,
        Rows:        rows,
    }, nil
}

// GetColumnIndex returns the index of a column by its name
func (r *Result) GetColumnIndex(name string) (int, bool) {
    for i, colName := range r.ColumnNames {
        if colName == name {
            return i, true
        }
    }
    return -1, false
}

// GetRow retrieves a row by its index
func (r *Result) GetRow(index int) (*Row, bool) {
    if index < 0 || index >= len(r.Rows) {
        return nil, false
    }
    return r.Rows[index], true
}

// GetValueByIndex retrieves the value of a cell in a row by column index
func (row *Row) GetValueByIndex(index int) (string, bool) {
    if index < 0 || index >= len(row.Values) {
        return "", false
    }
    return row.Values[index], true
}

// GetValueByName retrieves the value of a cell in a row by column name
func (row *Row) GetValueByName(name string) (string, bool) {
    colIdx, found := row.parent.GetColumnIndex(name)
    if !found {
        return "", false
    }
    return row.GetValueByIndex(colIdx)
}

func main() {
    // Initialize the WASM runtime
    runtime := wazero.NewRuntime()
    defer runtime.Close()

    // Load the WASM module
    wasmBytes, err := os.ReadFile("path_to_your_compiled_wasm_file.wasm")
    if err != nil {
        log.Fatalf("Failed to read WASM file: %v", err)
    }

    module, err := runtime.InstantiateModule(wasmBytes, wazero.NewModuleConfig().WithName("example"))
    if err != nil {
        log.Fatalf("Failed to instantiate module: %v", err)
    }

    // Create a shared memory buffer
    memSize := 1024 // Adjust as needed
    mem, err := runtime.NewMemory(memSize)
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }

    // Call the WASM function to get data from memory
    fillMemoryFunc := module.ExportedFunction("fill_memory_with_data")
    _, err = fillMemoryFunc.Call(mem, []byte{})
    if err != nil {
        log.Fatalf("Failed to call WASM function: %v", err)
    }

    // Read the memory data
    buffer := make([]byte, memSize)
    _, err = mem.ReadAt(buffer, 0)
    if err != nil {
        log.Fatalf("Failed to read memory: %v", err)
    }

    // Initialize the Result
    result, err := NewResult(buffer)
    if err != nil {
        log.Fatalf("Failed to initialize result: %v", err)
    }

    // Example usage of Result and Row methods
    row, found := result.GetRow(0)
    if !found {
        log.Println("Row not found")
        return
    }

    value, found := row.GetValueByName("Age")
    if found {
        fmt.Printf("Value by column name 'Age': %s\n", value)
    } else {
        fmt.Println("Column 'Age' not found")
    }

    value, found = row.GetValueByIndex(2)
    if found {
        fmt.Printf("Value by column index 2: %s\n", value)
    } else {
        fmt.Println("Index 2 not found")
    }
}
