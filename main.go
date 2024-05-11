package main

import (
	"fmt"
	"log"
	"net/http"

	"encoding/json"
	"io"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/patrickmn/go-cache"
)

type allCache struct {
    products *cache.Cache
}

const (
    defaultExpiration = 5 * time.Minute
    purgeTime         = 10 * time.Minute
)

func newCache() *allCache {
    Cache := cache.New(defaultExpiration, purgeTime)
    return &allCache{
        products: Cache,
    }
}

func (c *allCache) read(id string) (item []byte, ok bool) {
    product, ok := c.products.Get(id)
    if ok {
        log.Println("from cache")
        res, err := json.Marshal(product.(Product))
        if err != nil {
            log.Fatal("Error")
        }
        return res, true
    }
    return nil, false
}

func (c *allCache) update(id string, product Product) {
    c.products.Set(id, product, cache.DefaultExpiration)
}

var c = newCache()


func checkCache(f httprouter.Handle) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
        id := p.ByName("id")
        res, ok := c.read(id)
        if ok {
            w.Header().Set("Content-Type", "application/json")
            w.Write(res)
            return
        }
        log.Println("From Controller")
        f(w, r, p)
    }
}

type Product struct {
    Price       float64 `json:"price"`
    ID          int     `json:"id"`
    Title       string  `json:"title"`
    Category    string  `json:"category"`
    Description string  `json:"description`
    Image       string  `json:"image"`
}

func getProduct(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
    id := p.ByName("id")
    resp, err := http.Get("https://fakestoreapi.com/products/" + id)
    if err != nil {
        log.Fatal(err)
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
        return
    }
    product := Product{}
    parseErr := json.Unmarshal(body, &product)
    if parseErr != nil {
        log.Fatal(parseErr)
        return
    }
    response, ok := json.Marshal(product)
    if ok != nil {
        log.Fatal("somethng went wrong")
    }

    c.update(id, product)
    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
}


func main() {
    router := httprouter.New()
    router.GET("/product/:id", checkCache(getProduct))

    err := http.ListenAndServe(":8080", router)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Server running on :8080")
}