use std::collections::HashMap;
use sparql::query::{QueryResults, GraphResult, QuerySolution};
use oxigraph::model::{NamedNode, Subject, Predicate, Object, Literal, Triple};
use serde_json::{json, Value};

fn rdf_to_jsonld(graph: GraphResult) -> Value {
    let mut jsonld_context = HashMap::new();
    let mut jsonld_graph = Vec::new();

    for triple in graph.iter() {
        let Triple { subject, predicate, object } = triple;

        let subject_iri = match subject {
            Subject::NamedNode(node) => node.as_str().to_string(),
            Subject::BlankNode(bnode) => format!("_:{}", bnode.as_str()),
        };

        let predicate_iri = predicate.as_str().to_string();

        let jsonld_object = match object {
            Object::NamedNode(node) => json!({"@id": node.as_str()}),
            Object::BlankNode(bnode) => json!({"@id": format!("_:{}", bnode.as_str())}),
            Object::Literal(lit) => match lit.language() {
                Some(lang) => json!({"@value": lit.value(), "@language": lang}),
                None => match lit.datatype() {
                    Some(datatype) => json!({"@value": lit.value(), "@type": datatype.as_str()}),
                    None => json!(lit.value()),
                },
            },
        };

        let jsonld_subject = jsonld_graph.iter_mut().find(|entry| {
            if let Some(id) = entry.get("@id") {
                id == &json!(subject_iri)
            } else {
                false
            }
        });

        if let Some(entry) = jsonld_subject {
            if let Some(predicate_list) = entry.get_mut(predicate_iri.as_str()) {
                predicate_list.as_array_mut().unwrap().push(jsonld_object);
            } else {
                entry[predicate_iri.as_str()] = json!([jsonld_object]);
            }
        } else {
            jsonld_graph.push(json!({
                "@id": subject_iri,
                predicate_iri.as_str(): [jsonld_object]
            }));
        }

        jsonld_context.insert(predicate_iri, predicate_iri);
    }

    json!({
        "@context": jsonld_context,
        "@graph": jsonld_graph
    })
}

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
                QueryResults::Graph(graph) => {
                    if mime_type == "application/ld+json" {
                        let jsonld_output = rdf_to_jsonld(graph);
                        let jsonld_string = serde_json::to_string(&jsonld_output)
                            .map_err(|e| e.to_string())?;
                        cursor.write_all(jsonld_string.as_bytes()).map_err(|e| e.to_string())
                    } else {
                        let format = match mime_type.as_str() {
                            "application/n-triples" => ResultFormat::NTriples,
                            "application/n-quads" => ResultFormat::NQuads,
                            "text/turtle" => ResultFormat::Turtle,
                            _ => return Err("Unsupported MIME type for graph data".to_string()),
                        };
                        graph.write_graph(&mut cursor, format).map_err(|e| e.to_string())
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
