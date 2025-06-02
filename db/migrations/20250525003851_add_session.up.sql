CREATE TABLE "sessions" (
  "id" varchar(255) PRIMARY KEY NOT NULL,
  "user_email" varchar(255) NOT NULL,
  "refresh_token" varchar(512) NOT NULL,
  "is_revoked" boolean NOT NULL DEFAULT FALSE,
  "created_at" timestamp DEFAULT now(),
  "expires_at" timestamp
);
