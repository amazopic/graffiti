package system

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

func prov(kind graph.EndpointKind, key, node string) graph.Endpoint {
	return graph.Endpoint{Kind: kind, Key: key, Display: key, Node: node, Confidence: graph.ConfExtracted, Source: "openapi"}
}
func con(kind graph.EndpointKind, key, node string) graph.Endpoint {
	return graph.Endpoint{Kind: kind, Key: key, Display: key, Node: node, Confidence: graph.ConfInferred, Source: "literal"}
}

func findLink(ls []SystemLink, from, to, key string) *SystemLink {
	for i := range ls {
		if ls[i].FromService == from && ls[i].ToService == to && ls[i].Key == key {
			return &ls[i]
		}
	}
	return nil
}

func TestMatch_DiscoversCrossServiceEdges(t *testing.T) {
	docs := map[string]*graph.Document{
		"gateway": {
			Consumes: []graph.Endpoint{
				con(graph.EPHTTP, "GET /carts/{}", "gateway.go:handler"),
				con(graph.EPQueue, "orders.created", "gateway.go:emit"),
				con(graph.EPHTTP, "GET /missing", "gateway.go:x"),
			},
		},
		"carts": {
			Provides: []graph.Endpoint{
				prov(graph.EPHTTP, "GET /carts/{}", "carts.go:Get"),
				prov(graph.EPHTTP, "GET /unused", "carts.go:Unused"),
			},
		},
		"orders": {
			Provides: []graph.Endpoint{prov(graph.EPQueue, "orders.created", "orders.go:Sub")},
		},
	}

	res := Match(docs)

	if l := findLink(res.Links, "gateway", "carts", "GET /carts/{}"); l == nil {
		t.Fatalf("expected gateway→carts http link; links=%+v", res.Links)
	} else if l.FromNode != "gateway.go:handler" || l.ToNode != "carts.go:Get" {
		t.Errorf("link nodes wrong: %+v", l)
	}
	if findLink(res.Links, "gateway", "orders", "orders.created") == nil {
		t.Errorf("expected gateway→orders queue link; links=%+v", res.Links)
	}
	// dangling: GET /missing
	foundDangling := false
	for _, d := range res.Dangling {
		if d.Service == "gateway" && d.Key == "GET /missing" {
			foundDangling = true
		}
	}
	if !foundDangling {
		t.Errorf("expected dangling gateway GET /missing; got %+v", res.Dangling)
	}
	// orphan: carts GET /unused
	foundOrphan := false
	for _, o := range res.Orphans {
		if o.Service == "carts" && o.Key == "GET /unused" {
			foundOrphan = true
		}
	}
	if !foundOrphan {
		t.Errorf("expected orphan carts GET /unused; got %+v", res.Orphans)
	}

	// Impact: changing carts should affect gateway.
	imp := Impact(res, "carts")
	if len(imp.Affected) != 1 || imp.Affected[0] != "gateway" {
		t.Errorf("impact(carts).Affected = %v, want [gateway]", imp.Affected)
	}
	if len(imp.Direct) == 0 {
		t.Errorf("impact(carts).Direct should list the gateway→carts link")
	}
}

func TestMatch_Ambiguous(t *testing.T) {
	docs := map[string]*graph.Document{
		"client": {Consumes: []graph.Endpoint{con(graph.EPRPC, "Cart.Get", "c.go:x")}},
		"a":      {Provides: []graph.Endpoint{prov(graph.EPRPC, "Cart.Get", "a.go:y")}},
		"b":      {Provides: []graph.Endpoint{prov(graph.EPRPC, "Cart.Get", "b.go:z")}},
	}
	res := Match(docs)
	if len(res.Links) != 0 {
		t.Errorf("ambiguous match must not produce confident links; got %+v", res.Links)
	}
	if len(res.Ambiguous) != 2 {
		t.Errorf("expected 2 ambiguous candidates; got %+v", res.Ambiguous)
	}
}

func TestMatch_HTTPPathOnlyFallback(t *testing.T) {
	// consumer knows only the path (method GET default), provider declares POST.
	docs := map[string]*graph.Document{
		"client": {Consumes: []graph.Endpoint{con(graph.EPHTTP, "GET /orders/{}", "c.go:x")}},
		"orders": {Provides: []graph.Endpoint{prov(graph.EPHTTP, "POST /orders/{}", "o.go:y")}},
	}
	res := Match(docs)
	l := findLink(res.Links, "client", "orders", "POST /orders/{}")
	if l == nil {
		t.Fatalf("expected path-only fallback link; got %+v / dangling %+v", res.Links, res.Dangling)
	}
	if l.Confidence != graph.ConfInferred {
		t.Errorf("path-only link should be INFERRED; got %s", l.Confidence)
	}
}
