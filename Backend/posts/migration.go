package posts

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&SupplyPost{}, &DemandPost{}); err != nil {
		return err
	}

	if db.Dialector.Name() != "postgres" {
		return nil
	}

	return db.Exec(`
DO $$
BEGIN
	IF to_regclass('public.users') IS NOT NULL
		AND NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'fk_supply_posts_producer'
		)
	THEN
		ALTER TABLE supply_posts
		ADD CONSTRAINT fk_supply_posts_producer
		FOREIGN KEY (producer_id)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE RESTRICT;
	END IF;

	IF to_regclass('public.users') IS NOT NULL
		AND NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'fk_demand_posts_buyer'
		)
	THEN
		ALTER TABLE demand_posts
		ADD CONSTRAINT fk_demand_posts_buyer
		FOREIGN KEY (buyer_id)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE RESTRICT;
	END IF;
END $$;
`).Error
}
