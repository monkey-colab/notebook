use oxigraph::io::GraphFormat;
use oxigraph::model::GraphName;
use oxigraph::sparql::QueryResults;
use oxigraph::store::Store;
use oxigraph::model::Term;

fn main() {
    // Initialize the store
    let store = Store::new().unwrap();

    // Expanded N-Triples data without marital status
    let ntriples_data = r#"
        <http://example.org/alice> <http://example.org/name> "Alice" .
        <http://example.org/alice> <http://example.org/birthDate> "1985-06-15"^^<http://www.w3.org/2001/XMLSchema#date> .
        <http://example.org/alice> <http://example.org/income> "50000.00"^^<http://www.w3.org/2001/XMLSchema#decimal> .
        <http://example.org/alice> <http://example.org/employed> "true"^^<http://www.w3.org/2001/XMLSchema#boolean> .
        <http://example.org/bob> <http://example.org/name> "Bob" .
        <http://example.org/bob> <http://example.org/birthDate> "1990-07-22"^^<http://www.w3.org/2001/XMLSchema#date> .
        <http://example.org/bob> <http://example.org/income> "60000.00"^^<http://www.w3.org/2001/XMLSchema#decimal> .
        <http://example.org/bob> <http://example.org/employed> "false"^^<http://www.w3.org/2001/XMLSchema#boolean> .
        <http://example.org/charlie> <http://example.org/name> "Charlie" .
        <http://example.org/charlie> <http://example.org/birthDate> "1982-11-30"^^<http://www.w3.org/2001/XMLSchema#date> .
        <http://example.org/charlie> <http://example.org/income> "70000.00"^^<http://www.w3.org/2001/XMLSchema#decimal> .
        <http://example.org/charlie> <http://example.org/employed> "true"^^<http://www.w3.org/2001/XMLSchema#boolean> .
        <http://example.org/david> <http://example.org/name> "David" .
        <http://example.org/david> <http://example.org/birthDate> "1988-03-10"^^<http://www.w3.org/2001/XMLSchema#date> .
        <http://example.org/david> <http://example.org/income> "80000.00"^^<http://www.w3.org/2001/XMLSchema#decimal> .
        <http://example.org/david> <http://example.org/employed> "false"^^<http://www.w3.org/2001/XMLSchema#boolean> .
        <http://example.org/alice> <http://example.org/knows> <http://example.org/bob> .
        <http://example.org/alice> <http://example.org/knows> <http://example.org/charlie> .
        <http://example.org/bob> <http://example.org/knows> <http://example.org/david> .
        <http://example.org/charlie> <http://example.org/knows> <http://example.org/david> .
    "#;

    // Load the N-Triples data into the store
    store
        .load_graph(
            ntriples_data.as_bytes(),
            GraphFormat::NTriples,
            &GraphName::DefaultGraph,
            None,
        )
        .unwrap();

    // Query for NamedNodes (subjects)
    let named_nodes_query = "SELECT DISTINCT ?s WHERE { ?s ?p ?o }";
    let named_nodes_results = store.query(named_nodes_query).unwrap();
    println!("Named Nodes:");
    print_query_results(named_nodes_results);

    // Query for Literals (names)
    let literals_query = "SELECT ?name WHERE { ?s <http://example.org/name> ?name }";
    let literals_results = store.query(literals_query).unwrap();
    println!("\nLiterals:");
    print_query_results(literals_results);

    // Query for Birth Dates
    let birth_dates_query = "SELECT ?s ?birthDate WHERE { ?s <http://example.org/birthDate> ?birthDate }";
    let birth_dates_results = store.query(birth_dates_query).unwrap();
    println!("\nBirth Dates:");
    print_query_results(birth_dates_results);

    // Query for Income
    let income_query = "SELECT ?s ?income WHERE { ?s <http://example.org/income> ?income }";
    let income_results = store.query(income_query).unwrap();
    println!("\nIncome:");
    print_query_results(income_results);

    // Query for Employment Status (Boolean)
    let employment_query = "SELECT ?s ?employed WHERE { ?s <http://example.org/employed> ?employed }";
    let employment_results = store.query(employment_query).unwrap();
    println!("\nEmployment Status:");
    print_query_results(employment_results);

    // Check if there is anyone employed (true)
    let ask_employed_query = "ASK { ?s <http://example.org/employed> \"true\"^^<http://www.w3.org/2001/XMLSchema#boolean> }";
    let ask_employed_results = store.query(ask_employed_query).unwrap();

    match ask_employed_results {
        QueryResults::Boolean(true) => println!("\nThere is someone employed."),
        QueryResults::Boolean(false) => println!("\nThere is no one employed."),
        _ => eprintln!("Expected boolean result for ASK query.")
    }

    // Query for names and birth dates
    let query = r#"
        SELECT ?name ?birthDate
        WHERE {
            ?s <http://example.org/name> ?name .
            ?s <http://example.org/birthDate> ?birthDate .
        }
    "#;

    let results = store.query(query).unwrap();
    println!("Results for names and birth dates:");
    print_query_results(results);

        // Construct query to create a new graph with names and knows relationships
    let construct_query = r#"
        CONSTRUCT {
            ?s <http://example.org/name> ?name .
            ?s <http://example.org/knows> ?o .
        }
        WHERE {
            ?s <http://example.org/name> ?name .
            ?s <http://example.org/knows> ?o .
        }
    "#;

    let construct_results = store.query(construct_query).unwrap();
    println!("\nConstruct Query Results (New Graph):");
    print_query_results(construct_results);    
}

fn print_query_results(query_results: QueryResults) {
    match query_results {
        QueryResults::Solutions(solutions) => {
            for solution in solutions {
                match solution {
                    Ok(binding) => {
                        for (var, value) in binding.iter() {
                            let value_string = term_to_string(value);
                            println!("{}: {}", var.as_str(), value_string);
                        }
                    }
                    Err(e) => {
                        eprintln!("Error in solution: {}", e);
                    }
                }
            }
        }
        QueryResults::Graph(triples) => {
            for triple in triples {
                match triple {
                    Ok(triple) => {
                        println!("Triple: {} {} {}", triple.subject, triple.predicate, triple.object);
                    }
                    Err(e) => {
                        eprintln!("Error in triple: {}", e);
                    }
                }
            }
        }
        QueryResults::Boolean(result) => {
            println!("ASK Query Result: {}", result);
        }
    }
}

fn term_to_string(term: &Term) -> String {
    match term {
        Term::NamedNode(node) => format!("NamedNode({})", node.as_str()),
        Term::BlankNode(node) => format!("BlankNode({})", node.as_str()),
        Term::Literal(lit) => {
            let value = lit.value();
            format!("Literal({}^^{})", value, lit.datatype().as_str())
        }
        Term::Triple(triple) => format!(
            "Triple({})", triple.to_string()
        )
    }
}
