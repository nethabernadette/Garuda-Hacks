package product

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Product{}); err != nil {
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
			SELECT 1 FROM pg_constraint WHERE conname = 'fk_products_producer'
		)
	THEN
		ALTER TABLE products
		ADD CONSTRAINT fk_products_producer
		FOREIGN KEY (producer_id)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE RESTRICT;
	END IF;
END $$;
`).Error
}
