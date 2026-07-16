# Backend Specification

## Architecture
Modular Monolith using NestJS.

Modules:
- auth
- users
- products
- rfq
- aggregation
- offers
- orders
- pricing
- notifications

## Demand Aggregation Logic
When a new RFQ is created:
1. Find OPEN RFQs with same product.
2. Filter by city/radius.
3. Filter by delivery window.
4. Create or update Demand Group.
5. Update RFQ status to AGGREGATED.

## Main APIs
POST /auth/register
POST /auth/login

POST /products
GET /products

POST /rfq
GET /rfq

GET /demand-groups

POST /offers

POST /orders

GET /prices

## Business Rules
- Only Buyer creates RFQ.
- Only Producer submits Offer.
- One RFQ belongs to one Demand Group.
- One Demand Group can contain many RFQs.
- Accepted Offer creates Order.

## External API
Reference prices are synchronized from Badan Pangan on a scheduled job and cached locally.
