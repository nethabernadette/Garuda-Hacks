package notifications

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Notification{}); err != nil {
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
			SELECT 1 FROM pg_constraint WHERE conname = 'fk_notifications_user'
		)
	THEN
		ALTER TABLE notifications
		ADD CONSTRAINT fk_notifications_user
		FOREIGN KEY (user_id)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE;
	END IF;
END $$;
`).Error
}
