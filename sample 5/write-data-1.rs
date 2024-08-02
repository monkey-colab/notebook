use std::slice;

pub struct QueryResultsBuffer {
    buffer: Vec<u8>,
}

impl QueryResultsBuffer {
    pub fn new() -> Self {
        QueryResultsBuffer {
            buffer: Vec::new(),
        }
    }

    pub fn add_binding(&mut self, var: &str, value: &str) {
        // Write variable name length and variable name
        let var_len = var.len() as u32;
        self.buffer.extend_from_slice(&var_len.to_le_bytes());
        self.buffer.extend_from_slice(var.as_bytes());

        // Write value length and value
        let value_len = value.len() as u32;
        self.buffer.extend_from_slice(&value_len.to_le_bytes());
        self.buffer.extend_from_slice(value.as_bytes());
    }

    pub fn finalize_result(&mut self) {
        // Calculate the length of the current result
        let result_len = self.buffer.len() as u32;

        // Insert the length of the current result at the beginning
        let mut result_len_bytes = result_len.to_le_bytes().to_vec();
        result_len_bytes.extend_from_slice(&self.buffer);
        self.buffer = result_len_bytes;
    }

    pub fn as_ptr(&self) -> *const u8 {
        self.buffer.as_ptr()
    }

    pub fn len(&self) -> usize {
        self.buffer.len()
    }
}

// Helper function to convert terms to string representation
fn term_to_string(term: &Term) -> String {
    // Convert term to string (implement as needed)
    term.to_string()
}

pub fn print_query_results(query_results: QueryResults, buffer: &mut QueryResultsBuffer) {
    match query_results {
        QueryResults::Solutions(solutions) => {
            for solution in solutions {
                match solution {
                    Ok(binding) => {
                        for (var, value) in binding.iter() {
                            let value_string = term_to_string(value);
                            buffer.add_binding(var.as_str(), &value_string);
                        }
                        buffer.finalize_result(); // Finalize each result
                    }
                    Err(e) => {
                        eprintln!("Error in solution: {}", e);
                    }
                }
            }
        }
        QueryResults::Graph(_) => {
            eprintln!("Expected bindings but got a graph");
        }
        QueryResults::Boolean(result) => {
            buffer.add_binding("ASK Query Result", if result { "true" } else { "false" });
            buffer.finalize_result(); // Finalize each result
        }
    }
}

// Expose functions for Go to access the buffer
#[no_mangle]
pub extern "C" fn create_query_results_buffer() -> *mut QueryResultsBuffer {
    Box::into_raw(Box::new(QueryResultsBuffer::new()))
}

#[no_mangle]
pub extern "C" fn free_query_results_buffer(buffer: *mut QueryResultsBuffer) {
    if !buffer.is_null() {
        unsafe {
            Box::from_raw(buffer);
        }
    }
}

#[no_mangle]
pub extern "C" fn get_query_results_buffer_ptr(buffer: *const QueryResultsBuffer) -> *const u8 {
    unsafe {
        if buffer.is_null() {
            std::ptr::null()
        } else {
            (*buffer).as_ptr()
        }
    }
}

#[no_mangle]
pub extern "C" fn get_query_results_buffer_len(buffer: *const QueryResultsBuffer) -> usize {
    unsafe {
        if buffer.is_null() {
            0
        } else {
            (*buffer).len()
        }
    }
}
