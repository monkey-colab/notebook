struct QueryResult {
    data: Vec<u8>,
}


use std::convert::TryInto;

struct QueryResult {
    data: Vec<u8>,
}

struct Row {
    bindings_offset: usize,
    bindings_count: usize,
}

struct Binding {
    name_offset: usize,
    name_length: usize,
    value_offset: usize,
    value_length: usize,
}

impl QueryResult {
    fn new(rows: Vec<Vec<(String, String)>>) -> Self {
        let mut data = Vec::new();
        let mut row_offsets = Vec::new();
        let mut binding_offsets = Vec::new();
        let mut string_offsets = Vec::new();

        // Serialize rows
        for row in &rows {
            let bindings_start = data.len();
            for (name, value) in row {
                let name_offset = data.len();
                data.extend_from_slice(name.as_bytes());
                let name_length = name.len();

                let value_offset = data.len();
                data.extend_from_slice(value.as_bytes());
                let value_length = value.len();

                binding_offsets.push(Binding {
                    name_offset,
                    name_length,
                    value_offset,
                    value_length,
                });
            }

            let bindings_count = row.len();
            row_offsets.push(Row {
                bindings_offset: bindings_start,
                bindings_count,
            });
        }

        // Serialize row_offsets
        let row_start = data.len();
        for row in &row_offsets {
            data.extend_from_slice(&row.bindings_offset.to_le_bytes());
            data.extend_from_slice(&row.bindings_count.to_le_bytes());
        }

        // Serialize binding_offsets
        let binding_start = data.len();
        for binding in &binding_offsets {
            data.extend_from_slice(&binding.name_offset.to_le_bytes());
            data.extend_from_slice(&binding.name_length.to_le_bytes());
            data.extend_from_slice(&binding.value_offset.to_le_bytes());
            data.extend_from_slice(&binding.value_length.to_le_bytes());
        }

        QueryResult { data }
    }

    fn rows(&self) -> &[Row] {
        let row_start = self.data.len() - self.row_offsets_len();
        let row_bytes = &self.data[row_start..];
        let row_count = row_bytes.len() / std::mem::size_of::<Row>();
        let rows = unsafe {
            std::slice::from_raw_parts(
                row_bytes.as_ptr() as *const Row,
                row_count,
            )
        };
        rows
    }

    fn row_offsets_len(&self) -> usize {
        self.data.len() / std::mem::size_of::<Row>()
    }

    fn bindings(&self, row: &Row) -> &[Binding] {
        let binding_start = row.bindings_offset;
        let binding_bytes = &self.data[binding_start..];
        let binding_count = row.bindings_count;
        let bindings = unsafe {
            std::slice::from_raw_parts(
                binding_bytes.as_ptr() as *const Binding,
                binding_count,
            )
        };
        bindings
    }

    fn binding_strings(&self, binding: &Binding) -> (String, String) {
        let name = std::str::from_utf8(
            &self.data[binding.name_offset..binding.name_offset + binding.name_length],
        ).unwrap().to_string();
        let value = std::str::from_utf8(
            &self.data[binding.value_offset..binding.value_offset + binding.value_length],
        ).unwrap().to_string();
        (name, value)
    }
}

fn main() {
    let rows = vec![
        vec![("A".to_string(), "ValueA".to_string()), ("B".to_string(), "ValueB".to_string())],
        vec![("C".to_string(), "ValueC".to_string())],
    ];

    let query_result = QueryResult::new(rows);

    for row in query_result.rows() {
        for binding in query_result.bindings(row) {
            let (name, value) = query_result.binding_strings(binding);
            println!("Name: {}, Value: {}", name, value);
        }
    }
}
