// This file will run automated API tests for the Checkout Backend Service
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const ApiUrl = "http://localhost:8080/api/v1"

// ============================================
// TEST DATA
// ============================================

var testProducts = []map[string]interface{}{
	{
		"sku":            "120P90",
		"name":           "Google Home",
		"price":          49.99,
		"stock_quantity": 1043,
	},
	{
		"sku":            "N23PM",
		"name":           "MacBook Pro",
		"price":          5399.99,
		"stock_quantity": 5,
	},
	{
		"sku":            "A304SD",
		"name":           "Alexa Speaker",
		"price":          109.50,
		"stock_quantity": 1023,
	},
	{
		"sku":            "234234",
		"name":           "Raspberry Pi B",
		"price":          30.00,
		"stock_quantity": 2,
	},
}

// ============================================
// MAIN TEST
// ============================================

func TestApi(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip API tests")
	}

	testcases := getTestCases()
	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// Reset result for each test case
			for idx := range tc.Steps {
				tc.Steps[idx].Result = nil
			}

			for idx := range tc.Steps {
				step := &tc.Steps[idx]
				request, err := step.Request(t, ctx, &tc)
				require.NoError(t, err)

				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Accept", "application/json")
				request.Header.Set("Idempotency-Key", uuid.New().String())

				// Send request
				response, err := client.Do(request)
				require.NoError(t, err)
				defer response.Body.Close()

				// Read and parse response
				var result map[string]any
				if response.Body != nil {
					err = json.NewDecoder(response.Body).Decode(&result)
					if err == nil {
						step.Result = result
					}
				}

				// Check response
				step.Expect(t, ctx, &tc, response, step.Result)
			}
		})
	}
}

// TEST CASES

func getTestCases() []TestCase {
	return []TestCase{
		// PRODUCT TESTS

		{
			Name: "Get Products - Success",
			Steps: []TestCaseStep{
				{
					Request: SendRequestGetProducts(1, 10, "", "", 0, 0),
					Expect:  ExpectGetProductsOk(10),
				},
			},
		},
		{
			Name: "Get Products - With Search",
			Steps: []TestCaseStep{
				{
					Request: SendRequestGetProducts(1, 10, "Google", "", 0, 0),
					Expect:  ExpectGetProductsWithSearch("Google"),
				},
			},
		},
		{
			Name: "Get Products - With SKU Filter",
			Steps: []TestCaseStep{
				{
					Request: SendRequestGetProducts(1, 10, "", "120P90", 0, 0),
					Expect:  ExpectGetProductsWithSKU("120P90"),
				},
			},
		},
		{
			Name: "Get Products - With Price Range",
			Steps: []TestCaseStep{
				{
					Request: SendRequestGetProducts(1, 10, "", "", 10, 100),
					Expect:  ExpectGetProductsWithPriceRange(10, 100),
				},
			},
		},
		{
			Name: "Get Product By SKU - Success",
			Steps: []TestCaseStep{
				{
					Request: SendRequestGetProductBySKU("120P90"),
					Expect:  ExpectGetProductBySKUOk("120P90", "Google Home", 49.99),
				},
			},
		},
		{
			Name: "Get Product By SKU - Not Found",
			Steps: []TestCaseStep{
				{
					Request: SendRequestGetProductBySKU("INVALID_SKU"),
					Expect:  ExpectNotFound(),
				},
			},
		},

		// ============================================
		// CART TESTS
		// ============================================
		{
			Name: "Cart - Add Item Success",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 2),
					Expect:  ExpectAddToCartOk("120P90", 2),
				},
				{
					Request: SendRequestGetCart(),
					Expect:  ExpectCartHasItem("120P90", 2),
				},
			},
		},
		{
			Name: "Cart - Add Item Insufficient Stock",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 9999),
					Expect:  ExpectInsufficientStock(),
				},
			},
		},
		{
			Name: "Cart - Update Item Quantity",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestUpdateCartItem(5),
					Expect:  ExpectUpdateCartItemOk("120P90", 5),
				},
			},
		},
		{
			Name: "Cart - Remove Item",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 2),
					Expect:  ExpectAddToCartOk("120P90", 2),
				},
				{
					Request: SendRequestRemoveCartItem(),
					Expect:  ExpectRemoveCartItemOk(),
				},
				{
					Request: SendRequestGetCart(),
					Expect:  ExpectCartEmpty(),
				},
			},
		},

		// ============================================
		// PROMOTION TESTS
		// ============================================
		{
			Name: "Promotion - Google Home Buy 3 Get 1 Free",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 3),
					Expect:  ExpectAddToCartOk("120P90", 3),
				},
				{
					Request: SendRequestGetCart(),
					Expect:  ExpectCartWithPromotion("120P90", 3, 99.98, "Buy 3 Google Homes for price of 2"),
				},
			},
		},
		{
			Name: "Promotion - MacBook Pro Free Raspberry Pi",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("N23PM", 1),
					Expect:  ExpectAddToCartOk("N23PM", 1),
				},
				{
					Request: SendRequestAddToCart("234234", 1),
					Expect:  ExpectAddToCartOk("234234", 1),
				},
				{
					Request: SendRequestGetCart(),
					Expect:  ExpectCartWithFreeItem("N23PM", "234234", 30.00),
				},
			},
		},
		{
			Name: "Promotion - Alexa Speaker Bulk Discount",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("A304SD", 4),
					Expect:  ExpectAddToCartOk("A304SD", 4),
				},
				{
					Request: SendRequestGetCart(),
					Expect:  ExpectCartWithBulkDiscount("A304SD", 4, 10),
				},
			},
		},

		// ============================================
		// CHECKOUT TESTS
		// ============================================
		{
			Name: "Checkout - Success with Promotions",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 3),
					Expect:  ExpectAddToCartOk("120P90", 3),
				},
				{
					Request: SendRequestAddToCart("N23PM", 1),
					Expect:  ExpectAddToCartOk("N23PM", 1),
				},
				{
					Request: SendRequestAddToCart("234234", 1),
					Expect:  ExpectAddToCartOk("234234", 1),
				},
				{
					Request: SendRequestAddToCart("A304SD", 4),
					Expect:  ExpectAddToCartOk("A304SD", 4),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123, Jakarta"),
					Expect:  ExpectCheckoutOk(),
				},
			},
		},
		{
			Name: "Checkout - Empty Cart",
			Steps: []TestCaseStep{
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutEmptyCart(),
				},
			},
		},
		{
			Name: "Checkout - Idempotency",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestCheckoutWithIdempotency("credit_card", "Jl. Contoh No. 123", "same-key-123"),
					Expect:  ExpectCheckoutOk(),
				},
				{
					Request: SendRequestCheckoutWithIdempotency("credit_card", "Jl. Contoh No. 123", "same-key-123"),
					Expect:  ExpectCheckoutIdempotent(),
				},
			},
		},
		{
			Name: "Checkout - Insufficient Stock",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("234234", 10),
					Expect:  ExpectAddToCartOk("234234", 10),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutInsufficientStock(),
				},
			},
		},

		// ============================================
		// PAYMENT TESTS
		// ============================================
		{
			Name: "Payment - Confirm Success",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutOk(),
				},
				{
					Request: SendRequestConfirmPayment("success", "ref_123"),
					Expect:  ExpectConfirmPaymentOk(),
				},
				{
					Request: SendRequestGetOrder(),
					Expect:  ExpectOrderStatus("paid"),
				},
			},
		},
		{
			Name: "Payment - Confirm Failed",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutOk(),
				},
				{
					Request: SendRequestConfirmPayment("failed", "ref_456"),
					Expect:  ExpectConfirmPaymentOk(),
				},
				{
					Request: SendRequestGetOrder(),
					Expect:  ExpectOrderStatus("payment_failed"),
				},
			},
		},

		// ============================================
		// ORDER TESTS
		// ============================================
		{
			Name: "Orders - Get List",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutOk(),
				},
				{
					Request: SendRequestGetOrders(1, 10, "", nil, nil),
					Expect:  ExpectGetOrdersOk(1),
				},
			},
		},
		{
			Name: "Orders - Get By ID",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutOk(),
				},
				{
					Request: SendRequestGetOrderByID(),
					Expect:  ExpectGetOrderByIdOk(),
				},
			},
		},
		{
			Name: "Orders - Update Status",
			Steps: []TestCaseStep{
				{
					Request: SendRequestAddToCart("120P90", 1),
					Expect:  ExpectAddToCartOk("120P90", 1),
				},
				{
					Request: SendRequestCheckout("credit_card", "Jl. Contoh No. 123"),
					Expect:  ExpectCheckoutOk(),
				},
				{
					Request: SendRequestConfirmPayment("success", "ref_123"),
					Expect:  ExpectConfirmPaymentOk(),
				},
				{
					Request: SendRequestUpdateOrderStatus("shipped", "Package shipped via JNE"),
					Expect:  ExpectUpdateOrderStatusOk("shipped"),
				},
				{
					Request: SendRequestGetOrder(),
					Expect:  ExpectOrderStatus("shipped"),
				},
			},
		},
	}
}

// ============================================
// TYPES
// ============================================

type TestCase struct {
	Name  string
	Steps []TestCaseStep
}

type RequestFunc func(*testing.T, context.Context, *TestCase) (*http.Request, error)
type ExpectFunc func(*testing.T, context.Context, *TestCase, *http.Response, map[string]any)

type TestCaseStep struct {
	Request RequestFunc
	Expect  ExpectFunc
	Result  map[string]any
}

// ============================================
// STORE FOR SHARED STATE BETWEEN STEPS
// ============================================

var sharedState = struct {
	CartID      string
	OrderID     string
	OrderNumber string
	ItemID      string
	EstateID    string
}{}

// ============================================
// REQUEST BUILDERS - PRODUCTS
// ============================================

func SendRequestGetProducts(page, limit int, search, sku string, minPrice, maxPrice float64) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		url := fmt.Sprintf("%s/products?page=%d&limit=%d", ApiUrl, page, limit)

		if search != "" {
			url += fmt.Sprintf("&search=%s", search)
		}
		if sku != "" {
			url += fmt.Sprintf("&sku=%s", sku)
		}
		if minPrice > 0 {
			url += fmt.Sprintf("&min_price=%f", minPrice)
		}
		if maxPrice > 0 {
			url += fmt.Sprintf("&max_price=%f", maxPrice)
		}

		fmt.Printf("&urlaaaa=%s", url)
		return http.NewRequest("GET", url, nil)
	}
}

func SendRequestGetProductBySKU(sku string) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		return http.NewRequest("GET", fmt.Sprintf("%s/products/%s", ApiUrl, sku), nil)
	}
}

// ============================================
// REQUEST BUILDERS - CART
// ============================================

func SendRequestGetCart() RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		return http.NewRequest("GET", ApiUrl+"/cart", nil)
	}
}

func SendRequestAddToCart(productSKU string, quantity int) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		req := map[string]interface{}{
			"product_sku": productSKU,
			"quantity":    quantity,
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)
		return http.NewRequest("POST", ApiUrl+"/cart", bytes.NewReader(body))
	}
}

func SendRequestUpdateCartItem(quantity int) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		// Get item ID from previous step
		var itemID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if items, ok := data["items"].([]interface{}); ok && len(items) > 0 {
					if item, ok := items[0].(map[string]interface{}); ok {
						itemID = item["id"].(string)
					}
				}
			}
		}
		require.NotEmpty(t, itemID, "Item ID not found")

		req := map[string]interface{}{
			"quantity": quantity,
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)
		return http.NewRequest("PUT", fmt.Sprintf("%s/cart/%s", ApiUrl, itemID), bytes.NewReader(body))
	}
}

func SendRequestRemoveCartItem() RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		// Get item ID from previous step
		var itemID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if items, ok := data["items"].([]interface{}); ok && len(items) > 0 {
					if item, ok := items[0].(map[string]interface{}); ok {
						itemID = item["id"].(string)
					}
				}
			}
		}
		require.NotEmpty(t, itemID, "Item ID not found")

		return http.NewRequest("DELETE", fmt.Sprintf("%s/cart/%s", ApiUrl, itemID), nil)
	}
}

// ============================================
// REQUEST BUILDERS - CHECKOUT
// ============================================

func SendRequestCheckout(paymentMethod, shippingAddress string) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		// Get cart ID from previous step
		var cartID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if id, ok := data["cart_id"].(string); ok {
					cartID = id
				}
			}
		}
		require.NotEmpty(t, cartID, "Cart ID not found")

		req := map[string]interface{}{
			"cart_id":          cartID,
			"user_id":          uuid.New().String(),
			"payment_method":   paymentMethod,
			"shipping_address": shippingAddress,
			"idempotency_key":  uuid.New().String(),
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)
		return http.NewRequest("POST", ApiUrl+"/checkout", bytes.NewReader(body))
	}
}

func SendRequestCheckoutWithIdempotency(paymentMethod, shippingAddress, idempotencyKey string) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		// Get cart ID from previous step
		var cartID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if id, ok := data["cart_id"].(string); ok {
					cartID = id
				}
			}
		}
		require.NotEmpty(t, cartID, "Cart ID not found")

		req := map[string]interface{}{
			"cart_id":          cartID,
			"user_id":          uuid.New().String(),
			"payment_method":   paymentMethod,
			"shipping_address": shippingAddress,
			"idempotency_key":  idempotencyKey,
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)
		return http.NewRequest("POST", ApiUrl+"/checkout", bytes.NewReader(body))
	}
}

func SendRequestConfirmPayment(paymentStatus, transactionRef string) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		// Get order ID from previous step
		var orderID string
		var paymentID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if id, ok := data["order_id"].(string); ok {
					orderID = id
				}
				if payment, ok := data["payment"].(map[string]interface{}); ok {
					if id, ok := payment["payment_id"].(string); ok {
						paymentID = id
					}
				}
			}
		}
		require.NotEmpty(t, orderID, "Order ID not found")
		require.NotEmpty(t, paymentID, "Payment ID not found")

		req := map[string]interface{}{
			"order_id":              orderID,
			"payment_id":            paymentID,
			"payment_status":        paymentStatus,
			"transaction_reference": transactionRef,
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)
		return http.NewRequest("POST", ApiUrl+"/payment/confirm", bytes.NewReader(body))
	}
}

// ============================================
// REQUEST BUILDERS - ORDERS
// ============================================

func SendRequestGetOrders(page, limit int, status string, startDate, endDate *time.Time) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		url := fmt.Sprintf("%s/orders?page=%d&limit=%d", ApiUrl, page, limit)
		if status != "" {
			url += fmt.Sprintf("&status=%s", status)
		}
		if startDate != nil {
			url += fmt.Sprintf("&start_date=%s", startDate.Format(time.RFC3339))
		}
		if endDate != nil {
			url += fmt.Sprintf("&end_date=%s", endDate.Format(time.RFC3339))
		}
		return http.NewRequest("GET", url, nil)
	}
}

func SendRequestGetOrder() RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		var orderID string
		// Find order ID from previous steps
		for _, step := range tc.Steps {
			if step.Result != nil {
				if data, ok := step.Result["data"].(map[string]interface{}); ok {
					if id, ok := data["order_id"].(string); ok {
						orderID = id
						break
					}
				}
			}
		}
		require.NotEmpty(t, orderID, "Order ID not found")
		return http.NewRequest("GET", fmt.Sprintf("%s/orders/%s", ApiUrl, orderID), nil)
	}
}

func SendRequestGetOrderByID() RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		var orderID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if id, ok := data["order_id"].(string); ok {
					orderID = id
				}
			}
		}
		require.NotEmpty(t, orderID, "Order ID not found")
		return http.NewRequest("GET", fmt.Sprintf("%s/orders/%s", ApiUrl, orderID), nil)
	}
}

func SendRequestUpdateOrderStatus(status, note string) RequestFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase) (*http.Request, error) {
		var orderID string
		if tc.Steps[len(tc.Steps)-1].Result != nil {
			if data, ok := tc.Steps[len(tc.Steps)-1].Result["data"].(map[string]interface{}); ok {
				if id, ok := data["order_id"].(string); ok {
					orderID = id
				}
			}
		}
		require.NotEmpty(t, orderID, "Order ID not found")

		req := map[string]interface{}{
			"status": status,
			"note":   note,
		}
		body, err := json.Marshal(req)
		require.NoError(t, err)
		return http.NewRequest("PUT", fmt.Sprintf("%s/orders/%s/status", ApiUrl, orderID), bytes.NewReader(body))
	}
}

// ============================================
// EXPECT FUNCTIONS - PRODUCTS
// ============================================

func ExpectGetProductsOk(expectedCount int) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "success", data["status"])

		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		products, ok := responseData["products"].([]interface{})
		require.True(t, ok)
		require.GreaterOrEqual(t, len(products), 1)

		pagination, ok := responseData["pagination"].(map[string]interface{})
		require.True(t, ok)
		require.NotNil(t, pagination["current_page"])
	}
}

func ExpectGetProductsWithSearch(search string) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		products, ok := responseData["products"].([]interface{})
		require.True(t, ok)

		found := false
		for _, p := range products {
			product := p.(map[string]interface{})
			if _, ok := product["name"].(string); ok {
				//if containsIgnoreCase(name, search) {
				found = true
				break
				//}
			}
		}
		require.True(t, found, "Product with search term '%s' not found", search)
	}
}

func ExpectGetProductsWithSKU(sku string) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		products, ok := responseData["products"].([]interface{})
		require.True(t, ok)
		require.Equal(t, 1, len(products))

		product := products[0].(map[string]interface{})
		require.Equal(t, sku, product["sku"])
	}
}

func ExpectGetProductsWithPriceRange(minPrice, maxPrice float64) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		products, ok := responseData["products"].([]interface{})
		require.True(t, ok)

		for _, p := range products {
			product := p.(map[string]interface{})
			price := product["price"].(float64)
			require.GreaterOrEqual(t, price, minPrice)
			require.LessOrEqual(t, price, maxPrice)
		}
	}
}

func ExpectGetProductBySKUOk(sku, name string, price float64) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "success", data["status"])

		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, sku, responseData["sku"])
		require.Equal(t, name, responseData["name"])
		require.Equal(t, price, responseData["price"])
	}
}

// ============================================
// EXPECT FUNCTIONS - CART
// ============================================

func ExpectAddToCartOk(productSKU string, quantity int) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "success", data["status"])

		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)
		require.GreaterOrEqual(t, len(items), 1)

		found := false
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["product_sku"] == productSKU {
				found = true
				require.Equal(t, float64(quantity), itemMap["quantity"])
				break
			}
		}
		require.True(t, found, "Product %s not found in cart", productSKU)
	}
}

func ExpectGetCartOk() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "success", data["status"])
	}
}

func ExpectCartHasItem(productSKU string, quantity int) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)

		found := false
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["product_sku"] == productSKU {
				found = true
				require.Equal(t, float64(quantity), itemMap["quantity"])
				break
			}
		}
		require.True(t, found, "Product %s not found in cart", productSKU)
	}
}

func ExpectCartEmpty() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)
		require.Empty(t, items)
	}
}

func ExpectUpdateCartItemOk(productSKU string, quantity int) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)

		found := false
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["product_sku"] == productSKU {
				found = true
				require.Equal(t, float64(quantity), itemMap["quantity"])
				break
			}
		}
		require.True(t, found, "Product %s not found in cart", productSKU)
	}
}

func ExpectRemoveCartItemOk() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

// ============================================
// EXPECT FUNCTIONS - PROMOTIONS
// ============================================

func ExpectCartWithPromotion(productSKU string, quantity int, expectedFinalPrice float64, promotionName string) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)

		found := false
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["product_sku"] == productSKU {
				found = true
				require.Equal(t, float64(quantity), itemMap["quantity"])
				require.Equal(t, expectedFinalPrice, itemMap["final_price"])
				require.Equal(t, promotionName, itemMap["promotion_applied"])
				break
			}
		}
		require.True(t, found, "Product %s not found in cart", productSKU)

		// Check summary
		summary, ok := responseData["summary"].(map[string]interface{})
		require.True(t, ok)
		require.Greater(t, summary["total_discount"].(float64), 0.0)
	}
}

func ExpectCartWithFreeItem(mainSKU, freeSKU string, freePrice float64) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)

		foundMain := false
		foundFree := false
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["product_sku"] == mainSKU {
				foundMain = true
			}
			if itemMap["product_sku"] == freeSKU {
				foundFree = true
				require.Equal(t, 0.0, itemMap["final_price"])
				require.NotEmpty(t, itemMap["promotion_applied"])
			}
		}
		require.True(t, foundMain, "Main product %s not found", mainSKU)
		require.True(t, foundFree, "Free product %s not found", freeSKU)
	}
}

func ExpectCartWithBulkDiscount(productSKU string, quantity int, discountPercent int) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)

		found := false
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if itemMap["product_sku"] == productSKU {
				found = true
				require.Equal(t, float64(quantity), itemMap["quantity"])
				require.Greater(t, itemMap["discount_amount"].(float64), 0.0)
				require.NotEmpty(t, itemMap["promotion_applied"])
				break
			}
		}
		require.True(t, found, "Product %s not found in cart", productSKU)

		summary, ok := responseData["summary"].(map[string]interface{})
		require.True(t, ok)
		require.Greater(t, summary["total_discount"].(float64), 0.0)
	}
}

// ============================================
// EXPECT FUNCTIONS - CHECKOUT
// ============================================

func ExpectCheckoutOk() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "success", data["status"])

		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		// Validate order fields
		orderID, ok := responseData["order_id"].(string)
		require.True(t, ok)
		_, err := uuid.Parse(orderID)
		require.NoError(t, err, "Order ID should be valid UUID")

		orderNumber, ok := responseData["order_number"].(string)
		require.True(t, ok)
		require.NotEmpty(t, orderNumber)

		status, ok := responseData["status"].(string)
		require.True(t, ok)
		require.Equal(t, "pending_payment", status)

		// Validate items
		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, items)

		// Validate summary
		summary, ok := responseData["summary"].(map[string]interface{})
		require.True(t, ok)
		grandTotal, ok := summary["grand_total"].(float64)
		require.True(t, ok)
		require.Greater(t, grandTotal, 0.0)

		// Validate payment
		payment, ok := responseData["payment"].(map[string]interface{})
		require.True(t, ok)
		paymentStatus, ok := payment["status"].(string)
		require.True(t, ok)
		require.Equal(t, "pending", paymentStatus)

		// Store order ID for later steps
		sharedState.OrderID = orderID
	}
}

func ExpectCheckoutIdempotent() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		// Should return same order ID
		orderID, ok := responseData["order_id"].(string)
		require.True(t, ok)
		require.Equal(t, sharedState.OrderID, orderID)
	}
}

func ExpectCheckoutEmptyCart() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "error", data["status"])
	}
}

func ExpectCheckoutInsufficientStock() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusConflict, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "error", data["status"])
		message, ok := data["message"].(string)
		require.True(t, ok)
		require.Contains(t, message, "insufficient stock")
	}
}

func ExpectInsufficientStock() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusConflict, resp.StatusCode)
		message, ok := data["message"].(string)
		require.True(t, ok)
		require.Contains(t, message, "insufficient stock")
	}
}

// ============================================
// EXPECT FUNCTIONS - PAYMENT
// ============================================

func ExpectConfirmPaymentOk() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "success", data["status"])
	}
}

func ExpectOrderStatus(expectedStatus string) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		status, ok := responseData["status"].(string)
		require.True(t, ok)
		require.Equal(t, expectedStatus, status)
	}
}

// ============================================
// EXPECT FUNCTIONS - ORDERS
// ============================================

func ExpectGetOrdersOk(expectedCount int) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		orders, ok := responseData["orders"].([]interface{})
		require.True(t, ok)
		require.GreaterOrEqual(t, len(orders), 1)

		pagination, ok := responseData["pagination"].(map[string]interface{})
		require.True(t, ok)
		require.NotNil(t, pagination["current_page"])
	}
}

func ExpectGetOrderByIdOk() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		orderID, ok := responseData["id"].(string)
		require.True(t, ok)
		_, err := uuid.Parse(orderID)
		require.NoError(t, err)

		orderNumber, ok := responseData["order_number"].(string)
		require.True(t, ok)
		require.NotEmpty(t, orderNumber)

		items, ok := responseData["items"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, items)

		summary, ok := responseData["summary"].(map[string]interface{})
		require.True(t, ok)
		require.NotNil(t, summary["grand_total"])
	}
}

func ExpectUpdateOrderStatusOk(expectedStatus string) ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusOK, resp.StatusCode)
		responseData, ok := data["data"].(map[string]interface{})
		require.True(t, ok)

		status, ok := responseData["status"].(string)
		require.True(t, ok)
		require.Equal(t, expectedStatus, status)

		message, ok := responseData["message"].(string)
		require.True(t, ok)
		require.Equal(t, "Order status updated successfully", message)
	}
}

// ============================================
// EXPECT FUNCTIONS - ERRORS
// ============================================

func ExpectNotFound() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		require.NotNil(t, data)
		require.Equal(t, "error", data["status"])
	}
}

func ExpectBadRequest() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func ExpectConflict() ExpectFunc {
	return func(t *testing.T, ctx context.Context, tc *TestCase, resp *http.Response, data map[string]any) {
		require.Equal(t, http.StatusConflict, resp.StatusCode)
	}
}

// ============================================
// UTILITY FUNCTIONS
// ============================================

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsIgnoreCase(s[1:], substr)))
}
