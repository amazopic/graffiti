package com.example;

import java.util.List;

public class Greeter {
    public String greet(String name) {
        return sanitize(name);
    }

    private String sanitize(String n) {
        return n.trim();
    }
}
