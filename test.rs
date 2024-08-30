use oxigraph::sparql::{QueryResults, QueryTripleIter};
use oxigraph::model::{NamedNode, Term, Triple};
use serde_json::{json, Value};
use std::io::{Cursor, Write};
use std::collections::HashMap;

/// Convert a `QueryTripleIter` into JSON-LD format.
fn rdf_to_jsonld(triples: QueryTripleIter) -> Value {
    let mut context = HashMap::new();
    let mut graph = Vec::new();

    // Iterate over triples and build the JSON-LD structure
    for triple in triples {
        let Triple { subject, predicate, object } = triple.unwrap();

        let subject_iri = match subject {
            Term::NamedNode(node) => node.as_str().to_string(),
            Term::BlankNode(bnode) => format!("_:{}", bnode.as_str()),
            _ => continue, // Skip unsupported types for simplicity
        };

        let predicate_iri = predicate.as_str().to_string();

        let object_value = match object {
            Term::NamedNode(node) => json!({"@id": node.as_str()}),
            Term::BlankNode(bnode) => json!({"@id": format!("_:{}", bnode.as_str())}),
            Term::Literal(literal) => match literal.language() {
                Some(lang) => json!({"@value": literal.value(), "@language": lang}),
                None => json!({"@value": literal.value(), "@type": literal.datatype().as_str()}),
            },
            _ => continue, // Skip unsupported types for simplicity
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
