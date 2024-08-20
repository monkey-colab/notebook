#[no_mangle]
pub extern "C" fn query_with_mime(q_ptr: *mut u8, size: usize, mime_ptr: *mut u8, mime_size: usize) -> u64 {
    let query = ptr_to_string(q_ptr, size);
    let mime_type = ptr_to_string(mime_ptr, mime_size);
    let store = GRAPH_STORE.lock().unwrap();

    let mut buffer = Vec::new();

    // Initially push a placeholder for success/failure (1 byte) and length (4 bytes)
    buffer.push(0); // Placeholder for success/failure
    buffer.extend_from_slice(&[0; 4]); // Placeholder for the length

    let result = store.query(query.as_str())
        .map_err(|e| e.to_string())
        .and_then(|results| {
            let format = match mime_type.as_str() {
                "application/sparql-results+json" => ResultFormat::Json,
                "application/sparql-results+xml" => ResultFormat::Xml,
                "text/csv" => ResultFormat::Csv,
                "text/tab-separated-values" => ResultFormat::Tsv,
                _ => return Err("Unsupported MIME type".to_string()),
            };

            // Use a cursor to write directly into the buffer after the length placeholder
            let mut cursor = Cursor::new(&mut buffer);
            cursor.set_position(5); // Skip the 1 byte for success/failure and 4 bytes for length

            results.write(&mut cursor, format).map_err(|e| e.to_string())
        });

    match result {
        Ok(_) => {
            // Update the success byte to 1 (indicating success)
            buffer[0] = 1;

            // Calculate the actual content length (excluding the first 5 bytes)
            let content_len = (buffer.len() - 5) as u32;

            // Write the actual length into the reserved space
            buffer[1..5].copy_from_slice(&content_len.to_le_bytes());
        }
        Err(err_msg) => {
            // Failure byte is already 0, just write the error message

            // Write the error message to the buffer
            let err_bytes = err_msg.as_bytes();
            buffer.extend_from_slice(err_bytes);

            // Calculate the length of the error message
            let err_len = (buffer.len() - 5) as u32;

            // Write the actual length into the reserved space
            buffer[1..5].copy_from_slice(&err_len.to_le_bytes());
        }
    }

    let len = buffer.len();
    make_wasm_result(Box::into_raw(buffer.into_boxed_slice()) as *mut u8, len)
}
