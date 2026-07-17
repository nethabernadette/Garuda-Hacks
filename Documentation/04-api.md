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

Typical backend commands:

```bash
cd Backend
go mod tidy
go test ./...
```

GET /recommendations

POST /matches/interest

GET /matches

GET /chat/:id
POST /chat/:id/message

POST /agreement
GET /agreement/:id
POST /agreement/:id/confirm

GET /documents/:id
