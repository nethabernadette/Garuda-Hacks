# API Overview

POST /auth/register
POST /auth/login

GET /profile
PUT /profile

GET /posts
POST /posts

## Posts, Feed, and Search

All write endpoints require an authenticated user. Producers/Farmers can create and manage supply posts. Buyers can create and manage demand posts. Public/unauthenticated reads only return visible active/open posts.

### Feed

GET /posts

Query parameters:
- `type`: `supply`, `demand`, or `all`
- `q` or `search`: keyword search across product name, category, subcategory, description, and location fields
- `category`, `subcategory`, `location`, `status`, `unit`
- `price_min`, `price_max`, `budget_min`, `budget_max`
- `quantity_min`, `quantity_max`
- `needed_from`, `needed_until`, `created_from`, `created_until`
- `sort`: `newest`, `oldest`, `updated`, `price_asc`, `price_desc`, `budget_asc`, `budget_desc`
- `page`, `limit`

Response data:

```json
{
  "items": [
    {
      "post_type": "supply",
      "id": "uuid",
      "product_name": "Rice",
      "category": "grain",
      "quantity": 100,
      "unit": "kg",
      "location": "Jakarta",
      "relevance_score": 25
    }
  ],
  "page": 1,
  "limit": 20
}
```

GET /posts/search uses the same query parameters and response shape as `/posts`.

### Supply Posts

POST /posts/supply
GET /posts/supply
GET /posts/supply/me
GET /posts/supply/:id
PUT /posts/supply/:id
PATCH /posts/supply/:id/close
DELETE /posts/supply/:id

Create request:

```json
{
  "product_name": "Rice",
  "category": "grain",
  "subcategory": "white rice",
  "description": "Premium quality",
  "quantity": 100,
  "unit": "kg",
  "minimum_order_quantity": 10,
  "price_min": 10000,
  "price_max": 12000,
  "location": "Jakarta",
  "delivery_area": "Jakarta, Bogor",
  "availability_status": "available",
  "available_from": "2026-07-20",
  "available_until": "2026-08-20",
  "status": "active"
}
```

Statuses: `draft`, `active`, `closed`, `expired`.

### Demand Posts

POST /posts/demand
GET /posts/demand
GET /posts/demand/me
GET /posts/demand/:id
PUT /posts/demand/:id
PATCH /posts/demand/:id/close
DELETE /posts/demand/:id

Create request:

```json
{
  "product_name": "Rice",
  "category": "grain",
  "subcategory": "white rice",
  "description": "Restaurant weekly purchase",
  "quantity": 50,
  "unit": "kg",
  "budget_min": 9000,
  "budget_max": 13000,
  "delivery_location": "Jakarta",
  "needed_date": "2026-07-25",
  "frequency": "weekly",
  "additional_requirements": "Food grade packaging",
  "status": "open"
}
```

Statuses: `draft`, `open`, `matched`, `closed`, `expired`, `cancelled`.

## Notifications

Notifications are persisted in the database and scoped to the authenticated user.

GET /notifications
GET /notifications/unread-count
PATCH /notifications/:id/read
PATCH /notifications/read-all
DELETE /notifications/:id

Query parameters for `GET /notifications`:
- `page`, `limit`
- `unread=true`

Notification fields:
- `id`
- `user_id`
- `type`
- `title`
- `message`
- `reference_type`
- `reference_id`
- `is_read`
- `created_at`
- `read_at`

The backend creates deduplicated notifications when a new active supply post or open demand post matches another user's public profile category or location/delivery area.

## Backend Migration and Test Commands

New GORM migrations:
- `posts.Migrate(db)` creates `supply_posts` and `demand_posts`, with owner/status/category/location/date indexes and PostgreSQL foreign keys to `users`.
- `notifications.Migrate(db)` creates `notifications`, with user/read/reference indexes and a uniqueness constraint for duplicate prevention.
- `agreement.Migrate(db)` now also keeps `buyer_confirmed_at` and `producer_confirmed_at` on `agreements`.
- `document.RegisterRoutes(mux, db, authenticate)` registers RFQ/procurement summary and contact reveal endpoints. It does not create a separate document table.

Typical backend commands:

```bash
cd Backend
go mod tidy
go test ./...
```

GET /recommendations

Requires authentication. Returns personalized homepage recommendations from database history and Groq when enabled.

AI aliases:

POST /ai/search-history
GET /ai/recommendations
POST /ai/matchmaking
GET /agreements/:id/ai-verification
POST /agreements/:id/ai-verification
GET /agreements/:id/negotiation-summary

Agreement verification request:

```json
{
  "buyer_submission": {
    "buyer_company": "Buyer Co",
    "producer_company": "Producer Co",
    "product": "Premium Rice",
    "quantity": 500,
    "unit": "kg",
    "agreed_unit_price": 40000,
    "agreed_total_price": 20000000,
    "currency": "IDR",
    "delivery_area": "Jakarta",
    "delivery_schedule": "2026-07-25",
    "payment_terms": "Net 14",
    "additional_terms": ["Food grade packaging"]
  },
  "producer_submission": {
    "buyer_company": "Buyer Co",
    "producer_company": "Producer Co",
    "product": "Premium Rice",
    "quantity": 500,
    "unit": "kg",
    "agreed_unit_price": 40000,
    "currency": "IDR",
    "delivery_area": "Jakarta",
    "delivery_schedule": "2026-07-25",
    "payment_terms": "Net 14",
    "additional_terms": ["Food grade packaging"]
  },
  "buyer_final_confirm": true,
  "producer_final_confirm": true
}
```

POST /matches/interest

GET /matches

GET /chat/:id
POST /chat/:id/message

POST /agreement
GET /agreement/:id
POST /agreement/:id/confirm

## Agreement Documents and Contact Reveal

These endpoints are available only after an agreement has been confirmed by both parties.

Required agreement state:
- `status == CONFIRMED`
- `buyer_confirmed == true`
- `producer_confirmed == true`

If the authenticated user is not part of the agreement match, the API returns `403 Forbidden`.

If the agreement exists but has not been fully confirmed, the API returns `409 Conflict`.

### Procurement Summary

GET /agreements/:id/document

Returns a generated RFQ/procurement summary response. This is a pre-transaction document, not an invoice, payment receipt, or shipping document.

Response data:

```json
{
  "document_number": "RFQ-2026-000001",
  "agreement_id": "uuid",
  "generated_date": "2026-07-17T10:00:00Z",
  "summary": {
    "document_number": "RFQ-2026-000001",
    "agreement_id": "uuid",
    "producer_company": "Producer Co",
    "buyer_company": "Buyer Co",
    "product_list": [
      {
        "product_name": "Rice",
        "quantity": 100,
        "unit": "kg",
        "unit_price": 5000,
        "currency": "IDR",
        "total_value": 500000,
        "specifications": "Food grade",
        "delivery_date": "2026-07-20T00:00:00Z",
        "delivery_address": "Jakarta",
        "payment_terms": "Net 14"
      }
    ],
    "total_value": 500000,
    "currency": "IDR",
    "delivery_address": "Jakarta",
    "payment_terms": "Net 14",
    "agreement_status": "CONFIRMED",
    "producer_confirmation_timestamp": "2026-07-10T10:00:00Z",
    "buyer_confirmation_timestamp": "2026-07-10T09:00:00Z"
  },
  "html": "<!doctype html>..."
}
```

GET /agreements/:id/document/html

Returns printable semantic HTML with `Content-Type: text/html; charset=utf-8`.

Document number format:
- `RFQ-{year}-{sequence}`
- Example: `RFQ-2026-000001`

The sequence is generated in the service layer from agreement creation order within the same year.

### Contact Reveal

GET /agreements/:id/contact

Returns company contact information for both parties only after full agreement confirmation.

Response data:

```json
{
  "agreement_id": "uuid",
  "match_id": "uuid",
  "buyer": {
    "user_id": "uuid",
    "company_name": "Buyer Co",
    "business_address": "Jakarta",
    "email": "buyer@example.com",
    "phone_number": "+621",
    "website": "",
    "business_representative": "Buyer Co"
  },
  "producer": {
    "user_id": "uuid",
    "company_name": "Producer Co",
    "business_address": "West Java",
    "email": "producer@example.com",
    "phone_number": "+622",
    "website": "",
    "business_representative": "Producer Co"
  }
}
```

The backend reuses existing `users` and `user_profiles` records for contact data. There is no duplicated company contact table.

GET /documents/:id
