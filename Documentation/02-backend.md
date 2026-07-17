# Backend Specification

## Stack
- Go 1.24+
- Gin
- GORM (NO Prisma)
- PostgreSQL
- JWT
- Docker
- Swagger
- Redis (optional)

## Architecture
Clean Architecture

controller/
service/
repository/
model/
dto/
middleware/
config/
utils/
database/
migration/

## Modules
### Auth
JWT login/register

### User
Profiles, NIB verification

### Posts
Supply & demand posts

### AI Matching
Compatibility scoring and recommendations

### Match
Mutual interest logic

### Chat
Negotiation messages

### Agreement
Agreement form + PDF/RFQ generation

### Notification
Real-time updates

## Database
Users
Profiles
Verification
Posts
Matches
Chats
Messages
Agreements
Documents
Notifications

GORM handles migrations and relationships.
