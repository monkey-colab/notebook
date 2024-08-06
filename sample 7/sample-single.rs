use oxigraph::store::Store;
use lazy_static::lazy_static;
use std::cell::RefCell;

lazy_static! {
    static ref SINGLETON_STORE: RefCell<Store> = RefCell::new(Store::new().unwrap());
}

fn get_store() -> &'static RefCell<Store> {
    &SINGLETON_STORE
}

fn main() {
    let store = get_store();
    
    {
        let mut store_ref = store.borrow_mut();
        // Use the store as needed
        // e.g., store_ref.load_graph(..);
    }
    
    {
        let store_ref = store.borrow();
        // Read from the store as needed
        // e.g., store_ref.query(..);
    }
}


use oxigraph::store::Store;
use lazy_static::lazy_static;
use std::sync::{Arc, Mutex};

lazy_static! {
    static ref SINGLETON_STORE: Arc<Mutex<Store>> = Arc::new(Mutex::new(Store::new().unwrap()));
}

fn get_store() -> Arc<Mutex<Store>> {
    Arc::clone(&SINGLETON_STORE)
}

fn main() {
    let store = get_store();
    
    {
        let mut store_lock = store.lock().unwrap();
        // Use the store as needed
        // e.g., store_lock.load_graph(..);
    }
    
    {
        let store_lock = store.lock().unwrap();
        // Read from the store as needed
        // e.g., store_lock.query(..);
    }
}
