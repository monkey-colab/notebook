use std::collections::HashMap;
fn rdf_to_jsonld(triples: QueryTripleIter) -> Value {
    let mut graph: HashMap<String, HashMap<String, Vec<Value>>> = HashMap::new();

    // Iterate over triples and build the JSON-LD structure
    for triple in triples {
        let triple = triple.unwrap(); // Assuming the iterator contains valid triples

        // Extract the subject, predicate, and object
        let subject_iri = match triple.subject() {
            Subject::NamedNode(node) => node.as_str().to_string(),
            Subject::BlankNode(bnode) => format!("_:{}", bnode.as_str()),
        };

        let predicate_iri = triple.predicate().as_str().to_string();

        let object_value = match triple.object() {
            Term::NamedNode(node) => json!({"@id": node.as_str()}),
            Term::BlankNode(bnode) => json!({"@id": format!("_:{}", bnode.as_str())}),
            Term::Literal(literal) => match literal.language() {
                Some(lang) => json!({"@value": literal.value(), "@language": lang}),
                None => json!({"@value": literal.value(), "@type": literal.datatype().as_str()}),
            },
            _ => continue, // Skip unsupported object types for simplicity
        };

        // Update the graph HashMap
        let entry = graph.entry(subject_iri.clone()).or_insert_with(HashMap::new);
        let objects = entry.entry(predicate_iri).or_insert_with(Vec::new);
        objects.push(object_value);
    }

    // Convert the HashMap to a JSON array
    let mut result: Vec<Value> = graph.into_iter().map(|(subject, predicates)| {
        let mut subject_entry = json!({"@id": subject});
        for (predicate, objects) in predicates {
            subject_entry[predicate] = json!(objects);
        }
        subject_entry
    }).collect();

    json!(result)
}
