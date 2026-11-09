-- One-time cleanup AFTER deploying password-on-verification registration.
-- The old pendings table contains plaintext passwords from unfinished registrations.
-- WARNING: running this statement permanently discards all pending sign-ups.
-- People who started registration before the deployment must request a new code.
-- This SQL file is not executed automatically by the application.
DROP TABLE IF EXISTS pendings;
