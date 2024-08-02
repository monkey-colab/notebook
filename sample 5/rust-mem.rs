use wasm_bindgen::prelude::*;

// Allocates memory and returns a pointer to that memory
fn allocate(size: usize) -> *mut u8 {
    let layout = std::alloc::Layout::from_size_align(size, 1).unwrap();
    unsafe { std::alloc::alloc(layout) }
}

// Deallocates memory pointed to by the given pointer
fn deallocate(ptr: *mut u8, size: usize) {
    let layout = std::alloc::Layout::from_size_align(size, 1).unwrap();
    unsafe { std::alloc::dealloc(ptr, layout) }
}

// Allocates memory and returns a pointer to that memory
#[wasm_bindgen]
pub fn allocate_memory(size: usize) -> *mut u8 {
    allocate(size)
}

// Deallocates memory
#[wasm_bindgen]
pub fn deallocate_memory(ptr: *mut u8, size: usize) {
    deallocate(ptr, size)
}

// Takes a string from WebAssembly memory and returns a greeting
#[wasm_bindgen]
pub fn greet(ptr: *const u8, len: usize) -> String {
    let slice = unsafe { std::slice::from_raw_parts(ptr, len) };
    let input_str = std::str::from_utf8(slice).unwrap();
    format!("Hello, {}!", input_str)
}

// A function that takes two integers and returns their sum
#[wasm_bindgen]
pub fn sum_integers(a: i32, b: i32) -> i32 {
    a + b
}

// A function that returns a fixed string
#[wasm_bindgen]
pub fn get_fixed_string() -> String {
    "This is a fixed string from Rust.".to_string()
}

// Takes two strings from WebAssembly memory, concatenates them, and returns the result
#[wasm_bindgen]
pub fn concat_strings(ptr1: *const u8, len1: usize, ptr2: *const u8, len2: usize) -> String {
    let slice1 = unsafe { std::slice::from_raw_parts(ptr1, len1) };
    let slice2 = unsafe { std::slice::from_raw_parts(ptr2, len2) };
    let string1 = std::str::from_utf8(slice1).unwrap();
    let string2 = std::str::from_utf8(slice2).unwrap();
    format!("{}{}", string1, string2)
}
