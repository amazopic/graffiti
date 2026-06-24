package contract

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

func provKeys(t *testing.T, rel, src string) ([]graph.Endpoint, []graph.Endpoint) {
	t.Helper()
	return scanSource(rel, []byte(src))
}

func TestScan_NestJS(t *testing.T) {
	src := `import { Controller, Get, Post, MessagePattern } from '@nestjs/common';
@Controller('carts')
export class CartsController {
  @Get(':id')
  getCart() {}
  @Post()
  create() {}
  @MessagePattern('cart.checkout')
  onCheckout() {}
}`
	p, _ := provKeys(t, "carts.controller.ts", src)
	wantHTTP := map[string]bool{"GET /carts/{}": true, "POST /carts": true}
	for k := range wantHTTP {
		if has(p, graph.EPHTTP, k) == nil {
			t.Errorf("NestJS: missing provide %q; got %+v", k, p)
		}
	}
	if has(p, graph.EPQueue, "cart.checkout") == nil {
		t.Errorf("NestJS: missing @MessagePattern provide; got %+v", p)
	}
}

func TestScan_Spring(t *testing.T) {
	src := `@RestController
@RequestMapping("/api/users")
public class UserController {
  @GetMapping("/{id}")
  public User get() {}
  @RequestMapping(value = "/search", method = RequestMethod.POST)
  public List<User> search() {}
  @KafkaListener(topics = "user.events")
  public void onEvent() {}
}`
	p, _ := provKeys(t, "UserController.java", src)
	if has(p, graph.EPHTTP, "GET /api/users/{}") == nil {
		t.Errorf("Spring: missing GET /api/users/{} (class prefix + @GetMapping); got %+v", p)
	}
	if has(p, graph.EPHTTP, "POST /api/users/search") == nil {
		t.Errorf("Spring: missing POST /api/users/search (@RequestMapping); got %+v", p)
	}
	if has(p, graph.EPQueue, "user.events") == nil {
		t.Errorf("Spring: missing @KafkaListener provide; got %+v", p)
	}
}

func TestScan_FastAPI(t *testing.T) {
	src := `from fastapi import APIRouter
router = APIRouter()
@router.get("/items/{item_id}")
def read_item(): ...
@app.post("/items")
def make(): ...`
	p, _ := provKeys(t, "main.py", src)
	if has(p, graph.EPHTTP, "GET /items/{}") == nil {
		t.Errorf("FastAPI: missing GET /items/{}; got %+v", p)
	}
	if has(p, graph.EPHTTP, "POST /items") == nil {
		t.Errorf("FastAPI: missing POST /items; got %+v", p)
	}
}

func TestScan_KafkaNats(t *testing.T) {
	src := `func run() {
	kafkaTemplate.send("orders.created", payload)
	nc.Publish("billing.charge", data)
	nc.Subscribe("orders.created", handler)
}`
	p, c := provKeys(t, "worker.go", src)
	if has(c, graph.EPQueue, "orders.created") == nil {
		t.Errorf("Kafka: missing producer consume orders.created; got %+v", c)
	}
	if has(c, graph.EPQueue, "billing.charge") == nil {
		t.Errorf("NATS: missing publish consume billing.charge; got %+v", c)
	}
	if has(p, graph.EPQueue, "orders.created") == nil {
		t.Errorf("NATS: missing subscribe provide orders.created; got %+v", p)
	}
}

func TestScan_FrontendIsConsumer(t *testing.T) {
	// A React/Vue client: .get("/x") and axios/fetch are CONSUMES, never provides.
	react := `import React from 'react';
export function useCart() {
  const r = api.get('/carts/42');
  fetch('/orders');
  axios.post('/checkout', body);
}`
	p, c := provKeys(t, "useCart.tsx", react)
	if len(p) != 0 {
		t.Errorf("frontend file must yield no provides; got %+v", p)
	}
	for _, want := range []string{"GET /carts/{}", "GET /orders", "POST /checkout"} {
		if has(c, graph.EPHTTP, want) == nil {
			t.Errorf("frontend: missing consume %q; got %+v", want, c)
		}
	}

	// A .ts file importing vue is also frontend.
	vue := `import { defineComponent } from 'vue';
const data = await $fetch('/api/products');`
	p2, c2 := provKeys(t, "Products.ts", vue)
	if len(p2) != 0 {
		t.Errorf("vue .ts must yield no provides; got %+v", p2)
	}
	if has(c2, graph.EPHTTP, "GET /api/products") == nil {
		t.Errorf("vue: missing $fetch consume; got %+v", c2)
	}
}

func TestScan_BackendAxiosIsConsumeNotRoute(t *testing.T) {
	// A Go/Nest backend gateway using axios must read as consume, not a route.
	src := `package main
func proxy() { axios.get("http://carts:8080/carts/7") }`
	p, c := provKeys(t, "gateway.go", src)
	if has(p, graph.EPHTTP, "GET /carts/{}") != nil {
		t.Errorf("axios.get must not be a provide; provides=%+v", p)
	}
	if has(c, graph.EPHTTP, "GET /carts/{}") == nil {
		t.Errorf("axios.get should be a consume; consumes=%+v", c)
	}
}
