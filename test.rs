use std::collections::HashMap;
use oxigraph::sparql::{QueryResults, QueryTripleIter};
use oxigraph::model::{Subject, Object, NamedNode, Triple};
use serde_json::{json, Value};
use std::io::{Cursor, Write};

/// Convert a `QueryTripleIter` into JSON-LD format.
fn rdf_to_jsonld(triples: QueryTripleIter) -> Value {
    let mut context = HashMap::new();
    let mut graph = Vec::new();

    // Iterate over triples and build the JSON-LD structure
    for triple in triples {
        let Triple { subject, predicate, object } = triple.unwrap();

        let subject_iri = match subject {
            Subject::NamedNode(node) => node.as_str().to_string(),
            Subject::BlankNode(bnode) => format!("_:{}", bnode.as_str()),
        };

        let predicate_iri = predicate.as_str().to_string();

        let object_value = match object {
            Object::NamedNode(node) => json!({"@id": node.as_str()}),
            Object::BlankNode(bnode) => json!({"@id": format!("_:{}", bnode.as_str())}),
            Object::Literal(literal) => match literal.language() {
                Some(lang) => json!({"@value": literal.value(), "@language": lang}),
                None => match literal.datatype() {
                    Some(datatype) => json!({"@value": literal.value(), "@type": datatype.as_str()}),
                    None => json!(literal.value()),
                },
            },
        };

        // Try to find an existing subject entry in the JSON-LD graph
        if let Some(entry) = graph.iter_mut().find(|e| e["@id"] == subject_iri) {
            // Append to the existing predicate array, or create it if it doesn't exist
            entry
                .entry(predicate_iri.clone())
                .or_insert_with(|| json!([]))
                .as_array_mut()
                .unwrap()
                .push(object_value);
        } else {
            // Create a new entry for the subject
            graph.push(json!({
                "@id": subject_iri,
                predicate_iri: [object_value]
            }));
        }

        context.insert(predicate_iri, predicate_iri);  // Add the predicate to the context
    }

    // Return the complete JSON-LD object
    json!({
        "@context": context,
        "@graph": graph
    })
}

/// The main query function handling different MIME types including JSON-LD.
#[no_mangle]
pub extern "C" fn query_with_mime(q_ptr: *mut u8, size: usize, mime_ptr: *mut u8, mime_size: usize) -> u64 {
    let query = ptr_to_string(q_ptr, size);
    let mime_type = ptr_to_string(mime_ptr, mime_size);
    let store = GRAPH_STORE.lock().unwrap();

    let mut buffer = Vec::new();

    buffer.push(0); // Placeholder for success/failure
    buffer.extend_from_slice(&[0; 4]); // Placeholder for the length

    let result = store.query(query.as_str())
        .map_err(|e| e.to_string())
        .and_then(|results| {
            let mut cursor = Cursor::new(&mut buffer);
            cursor.set_position(5);

            match results {
                QueryResults::Boolean(value) => {
                    let format = match mime_type.as_str() {
                        "application/sparql-results+json" => ResultFormat::Json,
                        "application/sparql-results+xml" => ResultFormat::Xml,
                        _ => return Err("Unsupported MIME type for boolean results".to_string()),
                    };
                    value.write(&mut cursor, format).map_err(|e| e.to_string())
                }
                QueryResults::Solutions(solutions) => {
                    let format = match mime_type.as_str() {
                        "application/sparql-results+json" => ResultFormat::Json,
                        "application/sparql-results+xml" => ResultFormat::Xml,
                        "text/csv" => ResultFormat::Csv,
                        "text/tab-separated-values" => ResultFormat::Tsv,
                        _ => return Err("Unsupported MIME type".to_string()),
                    };
                    solutions.write(&mut cursor, format).map_err(|e| e.to_string())
                }
                QueryResults::Graph(triples) => {
                    if mime_type == "application/ld+json" {
                        let jsonld_output = rdf_to_jsonld(triples);
                        let jsonld_string = serde_json::to_string(&jsonld_output)
                            .map_err(|e| e.to_string())?;
                        cursor.write_all(jsonld_string.as_bytes()).map_err(|e| e.to_string())
                    } else {
                        // Handle other formats like Turtle or N-Triples
                        let format = match mime_type.as_str() {
                            "application/n-triples" => ResultFormat::NTriples,
                            "application/n-quads" => ResultFormat::NQuads,
                            "text/turtle" => ResultFormat::Turtle,
                            _ => return Err("Unsupported MIME type for graph data".to_string()),
                        };
                        // Implement serialization logic for other formats here.
                        unimplemented!("Other formats not implemented yet")
                    }
                }
            }
        });

    match result {
        Ok(_) => {
            buffer[0] = 1;
            let content_len = (buffer.len() - 5) as u32;
            buffer[1..5].copy_from_slice(&content_len.to_le_bytes());
        }
        Err(err_msg) => {
            let err_bytes = err_msg.as_bytes();
            buffer.extend_from_slice(err_bytes);
            let err_len = (buffer.len() - 5) as u32;
            buffer[1..5].copy_from_slice(&err_len.to_le_bytes());
        }
    }

    let len = buffer.len();
    make_wasm_result(Box::into_raw(buffer.into_boxed_slice()) as *mut u8, len)
}
