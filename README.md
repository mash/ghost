Ghost - Build REST APIs from structs using Generics
===================================================

## Why Ghost?

We all love Go for it's high performance IO, Goroutines, channels etc but we also have to write boring REST APIs that required a lot of hand code writing. We had to repeat ourselves for each resource and when I read the "When To Use Generics" I thought, REST APIs are not a collection, but we are actually writing the exact same code multiple times. Is building REST APIs a generics thing? What's the minimal we could write to get a REST API?

## What is Ghost?

Ghost is a collection of generic structs and interfaces that when composed, become a http.Handler that provides a REST API.

```
type User struct {
	Name string
}

type SearchQuery struct {
	Name string
}

func main() {
	store := ghost.NewMapStore(&User{}, SearchQuery{}, uint64(0))
	g := ghost.New(store)
	// g is a http.Handler and it provides GET /?name=:name, GET /:userid, POST /, PUT /:userid, DELETE /:userid
	http.ListenAndServe("127.0.0.1:8080", g)
}
```

Minimal work is to write your struct, choose a store from the [stores](./store) and call `ghost.New`.  
In reality you would write your own store. That's where the business logic is, and what translates into value.

## What is generic in Ghost?

Ghost has 3 type parameters, R Resource, Q Query and P PKey.

Resource is the REST resource. Ghost provides CRUD operations for resources.

Query represents the query when searching for resources to return a list.

PKey is a key that identifies the resource. A primary key.

## Types in Ghost



## TODO

- [x] Validator
- [x] Hooks
- [x] Example using common ORMs
- [ ] OpenAPIv3 integration
