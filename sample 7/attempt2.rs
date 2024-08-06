struct QueryResult {
    data: Vec<u8>,
    rows: Vec<Row>,
}

struct Row {
    bindings: Vec<Binding>,
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

        for row in rows {
            let mut bindings = Vec::new();
            for (name, value) in row {
                let name_offset = data.len();
                let name_length = name.len();
                data.extend_from_slice(name.as_bytes());

                let value_offset = data.len();
                let value_length = value.len();
                data.extend_from_slice(value.as_bytes());

                bindings.push(Binding {
                    name_offset,
                    name_length,
                    value_offset,
                    value_length,
                });
            }
            row_offsets.push(Row { bindings });
        }

        QueryResult {
            data,
            rows: row_offsets,
        }
    }

    fn get_binding_strings(&self, binding: &Binding) -> (String, String) {
        let name = String::from_utf8(
            self.data[binding.name_offset..binding.name_offset + binding.name_length].to_vec(),
        ).unwrap();
        let value = String::from_utf8(
            self.data[binding.value_offset..binding.value_offset + binding.value_length].to_vec(),
        ).unwrap();
        (name, value)
    }
}

fn main() {
    let rows = vec![
        vec![("A".to_string(), "ValueA".to_string()), ("B".to_string(), "ValueB".to_string())],
        vec![("C".to_string(), "ValueC".to_string())],
    ];

    let query_result = QueryResult::new(rows);

    for row in &query_result.rows {
        for binding in &row.bindings {
            let (name, value) = query_result.get_binding_strings(binding);
            println!("Name: {}, Value: {}", name, value);
        }
    }
}
