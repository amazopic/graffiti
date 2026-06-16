use std::collections::HashMap;

pub struct Store { data: HashMap<String, String> }

impl Store {
    pub fn get(&self, k: &str) -> String {
        normalize(k)
    }
}

fn normalize(k: &str) -> String {
    k.to_string()
}
