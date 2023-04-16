DROP TABLE IF EXISTS "verify_emails" CASCASE;

ALTER TABLE "users" DROP COLUMN "is_email_verified";