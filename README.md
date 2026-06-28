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

#### Get item

```http
  GET /api/items/${id}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. Id of item to fetch |

#### add(num1, num2)

Takes two numbers and returns the sum.

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
