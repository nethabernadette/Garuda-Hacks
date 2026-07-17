package users

type UserRole string

const (
	RoleBuyer    UserRole = "BUYER"
	RoleProducer UserRole = "PRODUCER"
	RoleAdmin    UserRole = "ADMIN"
)

type User struct {
	ID           uint     `gorm:"column:id;primaryKey" json:"id"`
	Role         UserRole `gorm:"column:role;type:varchar(20);not null;index" json:"role"`
	CompanyName  string   `gorm:"column:company_name;type:varchar(255);not null" json:"company_name"`
	Email        string   `gorm:"column:email;type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string   `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Phone        string   `gorm:"column:phone;type:varchar(50);not null" json:"phone"`
	City         string   `gorm:"column:city;type:varchar(120);not null;index" json:"city"`
}

func (User) TableName() string {
	return "users"
}

func (r UserRole) IsValid() bool {
	switch r {
	case RoleBuyer, RoleProducer, RoleAdmin:
		return true
	default:
		return false
	}
}
