-- Manual, idempotent MySQL migration. The platform also creates this table
-- on startup; the script documents it for operators who manage DDL by hand.
-- Holds at most one row (id = 1) of platform-wide settings.
-- interface_language: '' lets each user switch between Chinese and English;
-- 'en' or 'zh' locks the interface to that language and hides the switch.
-- No row means '' (users choose), so existing deployments keep their behavior.
CREATE TABLE IF NOT EXISTS platform_settings (
  id BIGINT NOT NULL PRIMARY KEY,
  interface_language VARCHAR(16) NOT NULL DEFAULT '',
  updated_by BIGINT NOT NULL DEFAULT 0,
  updated_at DATETIME(3) NULL
);
