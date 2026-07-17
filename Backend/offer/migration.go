package offer

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Offer{}); err != nil {
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
			SELECT 1 FROM pg_constraint WHERE conname = 'fk_offers_producer'
		)
	THEN
		ALTER TABLE offers
		ADD CONSTRAINT fk_offers_producer
		FOREIGN KEY (producer_id)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE RESTRICT;
	END IF;

	IF to_regclass('public.demand_groups') IS NOT NULL
		AND NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'fk_offers_demand_group'
		)
	THEN
		ALTER TABLE offers
		ADD CONSTRAINT fk_offers_demand_group
		FOREIGN KEY (group_id)
		REFERENCES demand_groups(id)
		ON UPDATE CASCADE
		ON DELETE RESTRICT;
	END IF;
END $$;
`).Error
}
