package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"ecommerce/cart-service/internal/domain"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
)

type redisCartRepo struct {
	client *redis.Client
}

// NewRedisCartRepository creates a new Redis repo
func NewRedisCartRepository(client *redis.Client) domain.CartRepository {
	return &redisCartRepo{client: client}
}

// We use a prefix to easily identify cart keys in Redis
const cartPrefix = "cart:"
// Carts expire automatically after 7 days of inactivity
const cartTTL = 7 * 24 * time.Hour 

// helper function to generate the key
func getCartKey(userID int64) string {
	return fmt.Sprintf("%s%d", cartPrefix, userID)
}

// 1. ADD OR UPDATE ITEM IN CART
func (r *redisCartRepo) AddToCart(ctx context.Context, userID int64, productID int64, quantity int32) error {
	tracer := otel.Tracer("cart-repository")
	ctx, span := tracer.Start(ctx, "Redis.AddToCart")
	defer span.End()

	key := getCartKey(userID)
	field := strconv.FormatInt(productID, 10) // Redis Hash fields must be strings

	// HINCRBY is magic: 
	// If the product isn't in the cart, it adds it with 'quantity'.
	// If it IS in the cart, it adds 'quantity' to the existing number!
	err := r.client.HIncrBy(ctx, key, field, int64(quantity)).Err()
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Reset the 7-day expiration timer every time they add an item
	return r.client.Expire(ctx, key, cartTTL).Err()
}

// 2. REMOVE ITEM FROM CART
func (r *redisCartRepo) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	tracer := otel.Tracer("cart-repository")
	ctx, span := tracer.Start(ctx, "Redis.RemoveItem")
	defer span.End()

	key := getCartKey(userID)
	field := strconv.FormatInt(productID, 10)

	// HDEL deletes a specific field (product) from the hash
	err := r.client.HDel(ctx, key, field).Err()
	if err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// 3. GET ENTIRE CART
func (r *redisCartRepo) GetCart(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	tracer := otel.Tracer("cart-repository")
	ctx, span := tracer.Start(ctx, "Redis.GetCart")
	defer span.End()

	key := getCartKey(userID)

	// HGETALL returns a map[string]string -> map["product_id"]"quantity"
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	var cart []domain.CartItem

	// Convert the map of strings back into our Go struct
	for prodIDStr, qtyStr := range result {
		productID, _ := strconv.ParseInt(prodIDStr, 10, 64)
		quantity, _ := strconv.ParseInt(qtyStr, 10, 32)

		// Filter out items where quantity dropped to 0 or below
		if quantity > 0 {
			cart = append(cart, domain.CartItem{
				ProductID: productID,
				Quantity:  int32(quantity),
			})
		} else {
            // Cleanup: If quantity is 0 or negative, remove it from Redis completely
            _ = r.client.HDel(ctx, key, prodIDStr).Err()
        }
	}

	return cart, nil
}

// 4. CLEAR ENTIRE CART (Call this when payment succeeds!)
func (r *redisCartRepo) ClearCart(ctx context.Context, userID int64) error {
	tracer := otel.Tracer("cart-repository")
	ctx, span := tracer.Start(ctx, "Redis.ClearCart")
	defer span.End()

	key := getCartKey(userID)

	// DEL completely deletes the key and all its contents
	return r.client.Del(ctx, key).Err()
}

// cart-service/internal/repository/cart_redis.go

func (r *redisCartRepo) GetCartCount(ctx context.Context, userID int64) (int32, error) {
	tracer := otel.Tracer("cart-repository")
	ctx, span := tracer.Start(ctx, "Redis.GetCartCount")
	defer span.End()

	key := getCartKey(userID)

	count, err := r.client.HLen(ctx, key).Result()
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	return int32(count), nil
}
