use oxigraph::model::*;
use oxigraph::store::*;
use std::sync::Mutex;
use std::ffi::{CString, CStr};
use std::os::raw::c_char;
use std::ptr;

// A global mutable reference to an in-memory Oxigraph store
lazy_static::lazy_static! {
    static ref GRAPH_STORE: Mutex<Store> = Mutex::new(Store::new());
}

// Load RDF data into the graph
#[no_mangle]
pub extern "C" fn load_rdf_data(data: *const c_char) -> bool {
    let c_str = unsafe { CStr::from_ptr(data) };
    let rdf_data = c_str.to_str().unwrap_or("");

    let mut store = GRAPH_STORE.lock().unwrap();
    let parser = oxigraph::sparql::Parser::default();
    match parser.parse(rdf_data) {
        Ok(triples) => store.insert(triples).is_ok(),
        Err(_) => false,
    }
}

// Clear the graph
#[no_mangle]
pub extern "C" fn clear_graph() -> bool {
    let mut store = GRAPH_STORE.lock().unwrap();
    store.clear().is_ok()
}

// Query the graph
#[no_mangle]
pub extern "C" fn query_graph(query: *const c_char) -> *mut c_char {
    let c_str = unsafe { CStr::from_ptr(query) };
    let query_str = c_str.to_str().unwrap_or("");

    let store = GRAPH_STORE.lock().unwrap();
    match store.query(query_str) {
        Ok(results) => {
            let result_str = results.iter().map(|triple| triple.to_string()).collect::<Vec<_>>().join("\n");
            CString::new(result_str).unwrap().into_raw()
        }
        Err(_) => ptr::null_mut(),
    }
}

// Add quad from a pattern
#[no_mangle]
pub extern "C" fn add_quad_from_pattern(subject: *const c_char, predicate: *const c_char, object: *const c_char, graph: *const c_char) -> bool {
    let c_subject = unsafe { CStr::from_ptr(subject) };
    let c_predicate = unsafe { CStr::from_ptr(predicate) };
    let c_object = unsafe { CStr::from_ptr(object) };
    let c_graph = unsafe { CStr::from_ptr(graph) };

    let subject = c_subject.to_str().unwrap_or("");
    let predicate = c_predicate.to_str().unwrap_or("");
    let object = c_object.to_str().unwrap_or("");
    let graph = c_graph.to_str().unwrap_or("");

    let mut store = GRAPH_STORE.lock().unwrap();

    let quad = Quad::new(
        Term::from_str(subject),
        Term::from_str(predicate),
        Term::from_str(object),
        Term::from_str(graph)
    );

    if let Ok(quad) = quad {
        store.insert(vec![quad]).is_ok()
    } else {
        false
    }
}

// Free memory allocated for query result
#[no_mangle]
pub extern "C" fn free_query_result(result: *mut c_char) {
    if !result.is_null() {
        unsafe { CString::from_raw(result) };
    }
}
