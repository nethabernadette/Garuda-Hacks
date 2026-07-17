package posts

import "strings"

const (
	scoreProduct     = 50
	scoreSubcategory = 40
	scoreCategory    = 25
	scoreLocation    = 15
	scoreQuantity    = 10
	scorePrice       = 10
	scoreDate        = 5
)

func scoreSupplyForDemand(supply SupplyPost, demand DemandPost) (int, []string) {
	score := 0
	reasons := make([]string, 0, 6)

	if equalFold(supply.ProductName, demand.ProductName) {
		score += scoreProduct
		reasons = append(reasons, "product name matches")
	}
	if supply.Subcategory != "" && equalFold(supply.Subcategory, demand.Subcategory) {
		score += scoreSubcategory
		reasons = append(reasons, "subcategory matches")
	}
	if equalFold(supply.Category, demand.Category) {
		score += scoreCategory
		reasons = append(reasons, "category matches")
	}
	if containsFold(supply.DeliveryArea, demand.DeliveryLocation) || containsFold(demand.DeliveryLocation, supply.Location) {
		score += scoreLocation
		reasons = append(reasons, "location or delivery area matches")
	}
	if supply.Quantity >= demand.Quantity {
		score += scoreQuantity
		reasons = append(reasons, "quantity can satisfy demand")
	}
	if demand.BudgetMax == 0 || supply.PriceMin <= demand.BudgetMax {
		score += scorePrice
		reasons = append(reasons, "price fits budget range")
	}
	if supply.AvailableUntil == nil || demand.NeededDate == nil || !supply.AvailableUntil.Before(*demand.NeededDate) {
		score += scoreDate
		reasons = append(reasons, "availability fits needed date")
	}

	return score, reasons
}

func scoreDemandForSupply(demand DemandPost, supply SupplyPost) (int, []string) {
	return scoreSupplyForDemand(supply, demand)
}

func equalFold(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	return a != "" && b != "" && strings.EqualFold(a, b)
}

func containsFold(haystack, needle string) bool {
	haystack = strings.ToLower(strings.TrimSpace(haystack))
	needle = strings.ToLower(strings.TrimSpace(needle))
	return haystack != "" && needle != "" && strings.Contains(haystack, needle)
}
