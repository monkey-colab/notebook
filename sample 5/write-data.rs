use wasm_bindgen::prelude::*;
use std::collections::HashMap;

#[wasm_bindgen]
pub fn fill_memory_with_data(memory: &mut [u8], column_names: Vec<&str>, rows: Vec<Vec<&str>>) -> usize {
    let mut offset = 0;

    // Write the number of columns
    let num_columns = column_names.len();
    memory[offset] = num_columns as u8;
    offset += 1;

    // Write column names
    for name in &column_names {
        let name_bytes = name.as_bytes();
        let name_len = name_bytes.len() as u8;

        memory[offset] = name_len;
        offset += 1;
        memory[offset..offset + name_len as usize].copy_from_slice(name_bytes);
        offset += name_len as usize;
    }

    // Write the number of rows
    let num_rows = rows.len();
    memory[offset..offset + 4].copy_from_slice(&(num_rows as u32).to_le_bytes());
    offset += 4;

    // Write rows
    for row in rows {
        for (i, cell) in row.iter().enumerate() {
            let cell_bytes = cell.as_bytes();
            let cell_len = cell_bytes.len() as u8;

            memory[offset] = cell_len;
            offset += 1;
            memory[offset..offset + cell_len as usize].copy_from_slice(cell_bytes);
            offset += cell_len as usize;
        }
    }

    offset
}
