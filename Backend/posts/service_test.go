package posts

import (
	"context"
	"errors"
	"testing"
	"time"

	"garuda-hacks/backend/users"
)

type fakeRepository struct {
	supplies []SupplyPost
	demands  []DemandPost
}

func (r *fakeRepository) CreateSupply(ctx context.Context, post *SupplyPost) error {
	if post.ID == "" {
		post.ID = "supply-1"
	}
	r.supplies = append(r.supplies, *post)
	return nil
}

func (r *fakeRepository) CreateDemand(ctx context.Context, post *DemandPost) error {
	if post.ID == "" {
		post.ID = "demand-1"
	}
	r.demands = append(r.demands, *post)
	return nil
}

func (r *fakeRepository) GetSupplyByID(ctx context.Context, id string) (*SupplyPost, error) {
	for i := range r.supplies {
		if r.supplies[i].ID == id {
			return &r.supplies[i], nil
		}
	}
	return nil, ErrPostNotFound
}

func (r *fakeRepository) GetDemandByID(ctx context.Context, id string) (*DemandPost, error) {
	for i := range r.demands {
		if r.demands[i].ID == id {
			return &r.demands[i], nil
		}
	}
	return nil, ErrPostNotFound
}

func (r *fakeRepository) UpdateSupply(ctx context.Context, post *SupplyPost) error {
	return nil
}

func (r *fakeRepository) UpdateDemand(ctx context.Context, post *DemandPost) error {
	return nil
}

func (r *fakeRepository) DeleteSupply(ctx context.Context, post *SupplyPost) error {
	return nil
}

func (r *fakeRepository) DeleteDemand(ctx context.Context, post *DemandPost) error {
	return nil
}

func (r *fakeRepository) ListSupply(ctx context.Context, filter QueryFilter) ([]SupplyPost, error) {
	return r.supplies, nil
}

func (r *fakeRepository) ListDemand(ctx context.Context, filter QueryFilter) ([]DemandPost, error) {
	return r.demands, nil
}

func (r *fakeRepository) ListSupplyByProducer(ctx context.Context, producerID string, filter QueryFilter) ([]SupplyPost, error) {
	return r.supplies, nil
}

func (r *fakeRepository) ListDemandByBuyer(ctx context.Context, buyerID string, filter QueryFilter) ([]DemandPost, error) {
	return r.demands, nil
}

func (r *fakeRepository) ListUsersByRole(ctx context.Context, roles []users.UserRole) ([]users.User, error) {
	return nil, nil
}

func (r *fakeRepository) GetUserByID(ctx context.Context, userID string) (*users.User, error) {
	return &users.User{ID: userID}, nil
}

func TestCreateSupplyRequiresProducer(t *testing.T) {
	svc := NewService(&fakeRepository{}, nil)
	req := validSupplyRequest()

	_, err := svc.CreateSupply(context.Background(), Principal{UserID: "buyer-1", Role: users.RoleBuyer}, req)
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}

	post, err := svc.CreateSupply(context.Background(), Principal{UserID: "producer-1", Role: users.RoleProducer}, req)
	if err != nil {
		t.Fatalf("expected producer to create supply, got %v", err)
	}
	if post.ProducerID != "producer-1" {
		t.Fatalf("expected authenticated producer id, got %q", post.ProducerID)
	}
}

func TestSupplyValidationPriceAndQuantity(t *testing.T) {
	svc := NewService(&fakeRepository{}, nil)
	req := validSupplyRequest()
	req.Quantity = 0
	_, err := svc.CreateSupply(context.Background(), Principal{UserID: "producer-1", Role: users.RoleProducer}, req)
	if err == nil {
		t.Fatal("expected quantity validation error")
	}

	req = validSupplyRequest()
	req.PriceMin = 100
	req.PriceMax = 10
	_, err = svc.CreateSupply(context.Background(), Principal{UserID: "producer-1", Role: users.RoleProducer}, req)
	if err == nil {
		t.Fatal("expected price range validation error")
	}
}

func TestCreateDemandRequiresBuyer(t *testing.T) {
	svc := NewService(&fakeRepository{}, nil)
	req := validDemandRequest()

	_, err := svc.CreateDemand(context.Background(), Principal{UserID: "producer-1", Role: users.RoleProducer}, req)
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}

	post, err := svc.CreateDemand(context.Background(), Principal{UserID: "buyer-1", Role: users.RoleBuyer}, req)
	if err != nil {
		t.Fatalf("expected buyer to create demand, got %v", err)
	}
	if post.BuyerID != "buyer-1" {
		t.Fatalf("expected authenticated buyer id, got %q", post.BuyerID)
	}
}

func TestDemandValidationBudgetAndNeededDate(t *testing.T) {
	svc := NewService(&fakeRepository{}, nil)
	req := validDemandRequest()
	req.BudgetMin = 100
	req.BudgetMax = 10
	_, err := svc.CreateDemand(context.Background(), Principal{UserID: "buyer-1", Role: users.RoleBuyer}, req)
	if err == nil {
		t.Fatal("expected budget range validation error")
	}

	past := Date{Time: time.Now().AddDate(0, 0, -1)}
	req = validDemandRequest()
	req.NeededDate = &past
	_, err = svc.CreateDemand(context.Background(), Principal{UserID: "buyer-1", Role: users.RoleBuyer}, req)
	if err == nil {
		t.Fatal("expected needed date validation error")
	}
}

func TestProducerCannotUpdateOtherProducerSupply(t *testing.T) {
	repo := &fakeRepository{supplies: []SupplyPost{{
		ID:          "supply-1",
		ProducerID:  "producer-1",
		ProductName: "Rice",
		Category:    "grain",
		Quantity:    10,
		Unit:        "kg",
		Location:    "Jakarta",
		PriceMin:    1,
		PriceMax:    2,
		Status:      SupplyPostStatusActive,
	}}}
	svc := NewService(repo, nil)
	_, err := svc.UpdateSupply(context.Background(), Principal{UserID: "producer-2", Role: users.RoleProducer}, "supply-1", UpdateSupplyPostRequest{})
	if !errors.Is(err, ErrPostForbidden) {
		t.Fatalf("expected ErrPostForbidden, got %v", err)
	}
}

func validSupplyRequest() CreateSupplyPostRequest {
	return CreateSupplyPostRequest{
		ProductName:          "Rice",
		Category:             "grain",
		Subcategory:          "white rice",
		Quantity:             100,
		Unit:                 "kg",
		MinimumOrderQuantity: 10,
		PriceMin:             10000,
		PriceMax:             12000,
		Location:             "Jakarta",
		DeliveryArea:         "Jakarta, Bogor",
		Status:               SupplyPostStatusActive,
	}
}

func validDemandRequest() CreateDemandPostRequest {
	needed := Date{Time: time.Now().AddDate(0, 0, 1)}
	return CreateDemandPostRequest{
		ProductName:      "Rice",
		Category:         "grain",
		Subcategory:      "white rice",
		Quantity:         50,
		Unit:             "kg",
		BudgetMin:        9000,
		BudgetMax:        13000,
		DeliveryLocation: "Jakarta",
		NeededDate:       &needed,
		Frequency:        "weekly",
		Status:           DemandPostStatusOpen,
	}
}
