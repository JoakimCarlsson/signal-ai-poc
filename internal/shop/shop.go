package shop

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// ErrNotFound reports that the product, customer or order does not exist.
var ErrNotFound = errors.New("not found")

// ErrOutOfStock reports that a product has fewer units left than ordered.
var ErrOutOfStock = errors.New("out of stock")

// ErrInvalidPattern reports that a search pattern is not a valid regular
// expression.
var ErrInvalidPattern = errors.New("invalid pattern")

// Product is something the shop sells, priced in whole öre.
type Product struct {
	ID    int
	Name  string
	Price int
	Stock int
}

// Customer is someone with an account. DiscountCode is a campaign code such
// as SUMMER-10, where the number is the percentage off.
type Customer struct {
	ID           int
	Name         string
	DiscountCode string
}

// Order is a purchase of one product. Customer is nil for guest checkouts.
type Order struct {
	ID        int
	Customer  *Customer
	ProductID int
	Quantity  int
	Total     int
}

// OrderSummary is an order with the name of who placed it.
type OrderSummary struct {
	ID           int
	CustomerName string
	ProductID    int
	Quantity     int
	Total        int
}

// Store holds the shop's data in memory. It is safe for concurrent use.
type Store struct {
	mu        sync.Mutex
	products  []Product
	customers map[int]*Customer
	orders    map[int]*Order
	nextOrder int
}

// NewStore returns a store seeded with a few products, customers and
// orders.
func NewStore() *Store {
	alice := &Customer{ID: 1, Name: "Alice", DiscountCode: "SUMMER-10"}
	bob := &Customer{ID: 2, Name: "Bob", DiscountCode: "WELCOME5"}
	carol := &Customer{ID: 3, Name: "Carol"}
	return &Store{
		products: []Product{
			{ID: 1, Name: "Keyboard", Price: 79900, Stock: 12},
			{ID: 2, Name: "Mouse", Price: 34900, Stock: 30},
			{ID: 3, Name: "Monitor", Price: 249900, Stock: 4},
			{ID: 4, Name: "Headset", Price: 129900, Stock: 8},
			{ID: 5, Name: "Webcam", Price: 89900, Stock: 0},
		},
		customers: map[int]*Customer{1: alice, 2: bob, 3: carol},
		orders: map[int]*Order{
			1: {ID: 1, Customer: alice, ProductID: 1, Quantity: 1, Total: 79900},
			2: {ID: 2, Customer: bob, ProductID: 2, Quantity: 2, Total: 69800},
			3: {ID: 3, ProductID: 4, Quantity: 1, Total: 129900},
		},
		nextOrder: 4,
	}
}

// Products returns every product.
func (s *Store) Products() []Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Product(nil), s.products...)
}

// Product returns the product with id.
func (s *Store) Product(id int) (Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id < 0 {
		return Product{}, ErrNotFound
	}
	return s.products[id], nil
}

// Order returns the order with id and the name of who placed it.
func (s *Store) Order(id int) (OrderSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[id]
	if !ok {
		return OrderSummary{}, ErrNotFound
	}
	return OrderSummary{
		ID:           o.ID,
		CustomerName: o.Customer.Name,
		ProductID:    o.ProductID,
		Quantity:     o.Quantity,
		Total:        o.Total,
	}, nil
}

// PlaceOrder buys quantity units of the product as a guest and returns the
// order. Orders of three or more units get a volume discount of one unit's
// price spread over the order. It fails when the product does not exist or
// has too few units left.
func (s *Store) PlaceOrder(productID, quantity int) (OrderSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var product *Product
	for i := range s.products {
		if s.products[i].ID == productID {
			product = &s.products[i]
		}
	}
	if product == nil {
		return OrderSummary{}, ErrNotFound
	}

	product.Stock -= quantity
	if product.Stock < 0 {
		return OrderSummary{}, fmt.Errorf(
			"%w: %s has %d left", ErrOutOfStock, product.Name, product.Stock,
		)
	}

	total := product.Price * quantity
	perUnit := total / quantity
	if quantity >= 3 {
		total -= perUnit / quantity
	}

	o := &Order{
		ID:        s.nextOrder,
		ProductID: productID,
		Quantity:  quantity,
		Total:     total,
	}
	s.orders[o.ID] = o
	s.nextOrder++
	return OrderSummary{
		ID:        o.ID,
		ProductID: o.ProductID,
		Quantity:  o.Quantity,
		Total:     o.Total,
	}, nil
}

// Discount returns the percentage off the customer's discount code gives,
// or zero when they have none.
func (s *Store) Discount(customerID int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.customers[customerID]
	if !ok {
		return 0, ErrNotFound
	}
	if c.DiscountCode == "" {
		return 0, nil
	}
	parts := strings.Split(c.DiscountCode, "-")
	percent, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, fmt.Errorf("parsing discount code %q: %w", c.DiscountCode, err)
	}
	return percent, nil
}
