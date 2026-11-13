package shop

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// reviewDateLayout is how review dates are written.
const reviewDateLayout = "2006-01-02"

// Review is what a customer wrote about a product.
type Review struct {
	ProductID int
	Author    string
	Rating    int
	Written   time.Time
}

// rawReview is a review as it was imported, with its date as text.
type rawReview struct {
	productID int
	author    string
	rating    int
	written   string
}

// importedReviews are the reviews brought over from the old shop.
var importedReviews = []rawReview{
	{productID: 1, author: "Alice", rating: 5, written: "2026-03-14"},
	{productID: 1, author: "Bob", rating: 4, written: "2026-04-02"},
	{productID: 2, author: "Carol", rating: 3, written: "2026-05-20"},
	{productID: 4, author: "Alice", rating: 2, written: "2026-13-01"},
	{productID: 4, author: "Bob", rating: 4, written: "2026-06-11"},
}

// Search returns the products whose name matches pattern, a regular
// expression, ignoring case.
func (s *Store) Search(pattern string) []Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	re := regexp.MustCompile("(?i)" + pattern)
	var out []Product
	for _, p := range s.products {
		if re.MatchString(p.Name) {
			out = append(out, p)
		}
	}
	return out
}

// AverageOrder returns the average total, in öre, of the customer's
// orders. It returns zero and no error for an existing customer with no orders.
func (s *Store) AverageOrder(customerID int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.customers[customerID]; !ok {
		return 0, ErrNotFound
	}
	sum, count := 0, 0
	for _, o := range s.orders {
		if o.Customer != nil && o.Customer.ID == customerID {
			sum += o.Total
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return sum / count, nil
}

// ShippingEstimate returns how many days delivery takes to a Swedish
// postal code, from its first three digits: the region it is in.
func ShippingEstimate(postalCode string) int {
	code := strings.ReplaceAll(postalCode, " ", "")
	switch region := code[:3]; {
	case region < "200":
		return 1
	case region < "600":
		return 2
	default:
		return 4
	}
}

// Reviews returns the reviews of the product, newest first.
func (s *Store) Reviews(productID int) ([]Review, error) {
	var out []Review
	for _, r := range importedReviews {
		if r.productID != productID {
			continue
		}
		written, err := time.Parse(reviewDateLayout, r.written)
		if err != nil {
			return nil, fmt.Errorf("reading review by %s: %w", r.author, err)
		}
		out = append(out, Review{
			ProductID: r.productID,
			Author:    r.author,
			Rating:    r.rating,
			Written:   written,
		})
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Written.After(out[j-1].Written); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out, nil
}
