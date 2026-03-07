# Mock Gen

This tool produces mocks for all interfaces in a file.

If compiled it can be used as a command line tool:

```shell
mock_gen <path_to_go_source_file>
```

It can be also executed using go run:

```bash
go run <path_mock_gen>/main.go -- <path_to_go_source_file> 
```

It will produce a file with at `./test/<source_file_name>_mocks.go`.

## Options

- `--output, -o` - Specify a custom output file path (optional)
- `--generate-return-handler` - Generates a single return handler field (e.g., Method_return func(...)) that allows dynamic return value handling
- `--thread-safe` - Adds a mutex to the mock struct and makes all methods thread-safe

## Example

With the following source file:
```go
package product

type Service interface {
	GetProductByID(ctx context.Context, productID model.ProductID) model.Product
}
```
It will produce the following mock:
```go
package test

import "your/project/internal/service/product/model"

type ServiceMock struct {
	GetProductByID_calls           int
	GetProductByID_param_context   context.Context
	GetProductByID_param_productID model.ProductID
	GetProductByID_return_product  model.Product
}

func (this *ServiceMock) GetProductByID(param1 context.Context, param2 model.ProductID) model.Product {
	this.GetProductByID_calls++
	this.GetProductByID_param_context = param1
	this.GetProductByID_param_productID = param2
	return this.GetProductByID_return_product
}
```

## Thread-Safe Example

When using the `--thread-safe` option:

```shell
mock_gen --thread-safe <path_to_go_source_file>
```

The generated mock will include a mutex and all methods will be thread-safe:
```go
package test

import (
	"sync"
	
	"your/project/internal/service/product/model"
)

type ServiceMock struct {
	mu sync.Mutex

	GetProductByID_calls           int
	GetProductByID_param_context   context.Context
	GetProductByID_param_productID model.ProductID
	GetProductByID_return_product  model.Product
}

func (this *ServiceMock) GetProductByID(param1 context.Context, param2 model.ProductID) model.Product {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.GetProductByID_calls++
	this.GetProductByID_param_context = param1
	this.GetProductByID_param_productID = param2
	return this.GetProductByID_return_product
}
```
