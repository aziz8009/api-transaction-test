# Checkout Backend Service - IFP 2026

## Overview

Backend service untuk sistem checkout e-commerce dengan dukungan promosi dinamis. Dibangun menggunakan **Golang** dengan arsitektur yang bersih dan terstruktur mengikuti prinsip SOLID.

### Fitur Utama

- Manajemen Produk (List, Pagination, Search, Filter)
- Shopping Cart (Add, Update, Delete, View)
- Sistem Promosi Dinamis (Strategy Pattern)
- Proses Checkout dengan Idempotensi
- Manajemen Order & Status Tracking
- Validasi Stok (Prevent Overselling)
- Logging & Monitoring
- Unit & Integration Tests

---

## Arsitektur

```bash
ApiBackendTest/
├── cmd/
│ └── main # Main API entry point
├── internal/
│ ├── config/ # Configuration management
│ ├── database/ # Database connection & utilities=
│ ├── models/ # Domain models
│ ├── repositories/ # Repository interfaces
│ ├── dto/
│ │ ├── request/ # Request DTOs
│ │ └── response/ # Response DTOs
│ ├── service/ # Business logic layer
│ ├── handler/ # HTTP handlers (Echo)
│ ├── middlewar/ # HTTP middleware
│ └── utils/ # Utility functions
├── scripts/ # Utility scripts
├── Dockerfile # API Docker image
├── docker-compose.yml # Docker Compose configuration
├── Makefile # Make commands
└── README.md
```

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- PostgreSQL 15+

### Running with Docker (Recommended)

```bash
# Clone repository
git clone <repository-url>
cd api-transaction-test

# Start all services
make docker-up

# Check logs
make docker-logs

```

### Local Development

```bash
# Setup environment
cp .env.example .env

# import db
database.sql

# seed data
seed.sql

# Run init , clean all cache and install dependencies
make init

# Run application
make run

# Run tests
make test
```

## API Reference

#### Get all products

```http
  GET /api/v1/products
```

| Parameter   | Type     | Description                     |
| :---------- | :------- | :------------------------------ |
| `page`      | `number` | **Required**. number pages      |
| `limit`     | `number` | **Required**. number limit data |
| `search`    | `number` | **Required**. number limit data |
| `min_price` | `number` | **Required**. number limit data |
| `max_price` | `number` | **Required**. number limit data |

#### Get product by sku

```http
  GET /api/v1/products/:sku
```

| Parameter | Type     | Description               |
| :-------- | :------- | :------------------------ |
| `sku`     | `string` | **Required**. Product SKU |

#### Get list carts

```http
  GET /api/v1/carts/
```

| Parameter | Type | Description |
| :-------- | :--- | :---------- |

#### Add to cart

```http
  POST /api/v1/cart/
```

| body param    | Type     | Description               |
| :------------ | :------- | :------------------------ |
| `product_sku` | `string` | **Required**. Product SKU |
| `quantity`    | `string` | **Required**. quantity    |

#### Update items cart

```http
  PUT /api/v1/cart/:itemId
```

| Parameter | Type   | Description                |
| :-------- | :----- | :------------------------- |
| `itemId`  | `UUID` | **Required**. Cart item ID |

| body param | Type     | Description            |
| :--------- | :------- | :--------------------- |
| `quantity` | `string` | **Required**. quantity |

#### Delete items cart

```http
  DELETE /api/v1/cart/:itemId
```

| Parameter | Type   | Description                |
| :-------- | :----- | :------------------------- |
| `itemId`  | `UUID` | **Required**. Cart item ID |

#### Checkout

```http
  POST /api/v1/checkout/
```

| body param         | Type     | Description                                                |
| :----------------- | :------- | :--------------------------------------------------------- |
| `idempotency_key`  | `string` | **Required**. Unique key to prevent duplicate transactions |
| `cart_id`          | `UUID`   | **Required**. cart_id                                      |
| `user_id`          | `UUID`   | **Required**. user_id                                      |
| `payment_method`   | `string` | **Required**. Payment method (e.g., credit_card)           |
| `shipping_address` | `string` | **Required**. Shipping address                             |

#### Confirm Payment

```http
  POST /api/v1/payment/confirm
```

| body param              | Type     | Description                                              |
| :---------------------- | :------- | :------------------------------------------------------- |
| `order_id`              | `UUID`   | **Required**. order_id                                   |
| `payment_id`            | `UUID`   | **Required**. payment_id                                 |
| `payment_status`        | `string` | **Required**. Payment status (success/failed)            |
| `transaction_reference` | `string` | **Required**. Transaction reference from payment gateway |

#### GET All Orders

```http
  GET /api/v1/orders
```

| parametes    | Type      | Description                     |
| :----------- | :-------- | :------------------------------ |
| `page`       | `integer` | pagen number                    |
| `limit`      | `integer` | item per page (default:10)      |
| `statys`     | `string`  | Filter by status                |
| `start_date` | `string`  | Filter by start date (ISO 8601) |
| `end_date`   | `string`  | Filter by end date (ISO 8601)   |

#### GET Orders By ID

```http
  GET /api/v1/orders/:id
```

| parametes | Type   | Description            |
| :-------- | :----- | :--------------------- |
| `id`      | `UUID` | **Required**. order_id |

#### UPDATE Statys Orders By ID

```http
  PUT /api/v1/orders/:id/status
```

| parametes | Type   | Description            |
| :-------- | :----- | :--------------------- |
| `id`      | `UUID` | **Required**. order_id |

| Body     | Type     | Description                    |
| :------- | :------- | :----------------------------- |
| `status` | `string` | **Required**. New order statys |
| `note`   | `string` | Optional note about the update |

## Database Schema

```bash

┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    products     │     │   cart_items    │     │     carts       │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ sku (PK)        │◄────│ product_sku (FK)│     │ id (PK)         │
│ name            │     │ cart_id (FK)    │────►│ total_amount    │
│ price           │     │ quantity        │     │ discount_amount │
│ stock_quantity  │     │ unit_price      │     │ final_amount    │
│ created_at      │     │ total_price     │     │ status          │
│ updated_at      │     │ discount_amount │     │ created_at      │
└─────────────────┘     │ created_at      │     │ updated_at      │
                        │ updated_at      │     └─────────────────┘
                        └─────────────────┘
                               │
                               │
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   order_items   │     │     orders      │     │  promotions     │
├─────────────────┤     ├─────────────────┤     ├─────────────────┤
│ id (PK)         │     │ id (PK)         │     │ id (PK)         │
│ order_id (FK)   │────►│ order_number    │     │ name            │
│ product_sku (FK)│     │ user_id         │     │ description     │
│ quantity        │     │ grand_total     │     │ type            │
│ unit_price      │     │ discount_total  │     │ is_active       │
│ discount        │     │ status          │     │ priority        │
│ final_price     │     │ shipping_address│     │ start_date      │
│ promotion_applied│    │ idempotency_key │     │ end_date        │
│ promotion_details│    │ created_at      │     │ created_at      │
│ created_at      │     │ updated_at      │     │ updated_at      │
│ updated_at      │     └─────────────────┘     └─────────────────┘
└─────────────────┘                                      │
                                                         │
                                                ┌────────┴────────┐
                                                │ promotion_rules │
                                                ├─────────────────┤
                                                │ id (PK)         │
                                                │ promotion_id(FK)│
                                                │ condition_type  │
                                                │ condition_value │
                                                │ action_type     │
                                                │ action_value    │
                                                │ target_product_sku│
                                                │ created_at      │
                                                │ updated_at      │
                                                └─────────────────┘

```
