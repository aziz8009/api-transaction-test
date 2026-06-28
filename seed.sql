-- ============================================
-- SEED PRODUCTS
-- ============================================
INSERT INTO products (sku, name, price, stock_quantity) VALUES
('120P90', 'Google Home', 49.99, 1043),
('N23PM', 'MacBook Pro', 5399.99, 5),
('A304SD', 'Alexa Speaker', 109.50, 1023),
('234234', 'Raspberry Pi B', 30.00, 2)
ON CONFLICT (sku) DO NOTHING;

-- ============================================
-- SEED PROMOTIONS
-- ============================================

-- 1. Google Home: Buy 3 Get 2 (Buy 3, get 1 free)
INSERT INTO promotions (id, name, description, type, is_active, priority) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'Buy 3 Google Homes for price of 2',
    'Buy 3 Google Homes and get 1 free',
    'multi_buy',
    true,
    10
)
ON CONFLICT (id) DO NOTHING;

-- 2. MacBook Pro: Free Raspberry Pi
INSERT INTO promotions (id, name, description, type, is_active, priority) VALUES
(
    '22222222-2222-2222-2222-222222222222',
    'MacBook Pro with free Raspberry Pi B',
    'Each MacBook Pro comes with free Raspberry Pi B',
    'bundle',
    true,
    20
)
ON CONFLICT (id) DO NOTHING;

-- 3. Alexa Speaker: Bulk Discount
INSERT INTO promotions (id, name, description, type, is_active, priority) VALUES
(
    '33333333-3333-3333-3333-333333333333',
    'Alexa Speaker Bulk Discount',
    'Buy more than 3 Alexa Speakers get 10% discount',
    'bulk_discount',
    true,
    5
)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- SEED PROMOTION RULES
-- ============================================

-- Google Home: Buy 3 Get 2 (buy 3, get 1 free)
INSERT INTO promotion_rules (promotion_id, condition_type, condition_value, action_type, action_value, target_product_sku) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'product_sku',
    '120P90',
    'buy_x_get_y_free',
    '3:1',
    NULL
)
ON CONFLICT DO NOTHING;

-- MacBook Pro: Free Raspberry Pi
INSERT INTO promotion_rules (promotion_id, condition_type, condition_value, action_type, action_value, target_product_sku) VALUES
(
    '22222222-2222-2222-2222-222222222222',
    'product_sku',
    'N23PM',
    'free_product',
    '1',
    '234234'
)
ON CONFLICT DO NOTHING;

-- Alexa Speaker: Bulk discount (more than 3 items = 10% off)
INSERT INTO promotion_rules (promotion_id, condition_type, condition_value, action_type, action_value, target_product_sku) VALUES
(
    '33333333-3333-3333-3333-333333333333',
    'min_quantity',
    '3',
    'bulk_discount',
    '3:10%',
    NULL
)
ON CONFLICT DO NOTHING;

-- ============================================
-- ADDITIONAL PROMOTIONS FOR DEMO (Optional)
-- ============================================

-- 4. Cart Total Discount: 5% off for orders over $500
INSERT INTO promotions (id, name, description, type, is_active, priority) VALUES
(
    '44444444-4444-4444-4444-444444444444',
    '5% Off Orders Over $500',
    'Get 5% discount on total order when spending over $500',
    'percentage_discount',
    true,
    1
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO promotion_rules (promotion_id, condition_type, condition_value, action_type, action_value, target_product_sku) VALUES
(
    '44444444-4444-4444-4444-444444444444',
    'cart_total',
    '500',
    'discount_percentage',
    '5',
    NULL
)
ON CONFLICT DO NOTHING;

-- ============================================
-- VERIFY DATA
-- ============================================

-- Check products
SELECT 'Products count: ' || COUNT(*)::text FROM products;

-- Check promotions
SELECT 'Promotions count: ' || COUNT(*)::text FROM promotions;

-- Check promotion rules
SELECT 'Promotion rules count: ' || COUNT(*)::text FROM promotion_rules;

-- Display all promotions with their rules
SELECT 
    p.name as promotion_name,
    p.type as promotion_type,
    pr.condition_type,
    pr.condition_value,
    pr.action_type,
    pr.action_value,
    pr.target_product_sku
FROM promotions p
JOIN promotion_rules pr ON p.id = pr.promotion_id
WHERE p.is_active = true
ORDER BY p.priority DESC;