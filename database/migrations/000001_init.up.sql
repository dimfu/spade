BEGIN;

USE spade;

CREATE TABLE IF NOT EXISTS tournament_types(
  id INT AUTO_INCREMENT PRIMARY KEY,
  size ENUM('2', '4', '8', '16', '32', '64'),
  bracket_type ENUM('single_elim', 'double_elim'),
  has_third_winner BOOLEAN DEFAULT false
);

INSERT INTO tournament_types (size, bracket_type, has_third_winner)
VALUES
  -- Single Elimination Brackets
  ('2', 'single_elim', false),
  ('2', 'single_elim', true),
  ('4', 'single_elim', false),
  ('4', 'single_elim', true),
  ('8', 'single_elim', false),
  ('8', 'single_elim', true),
  ('16', 'single_elim', false),
  ('16', 'single_elim', true),
  ('32', 'single_elim', false),
  ('32', 'single_elim', true),
  ('64', 'single_elim', false),
  ('64', 'single_elim', true),

  -- Double Elimination Brackets
  ('2', 'double_elim', false),
  ('2', 'double_elim', true),
  ('4', 'double_elim', false),
  ('4', 'double_elim', true),
  ('8', 'double_elim', false),
  ('8', 'double_elim', true),
  ('16', 'double_elim', false),
  ('16', 'double_elim', true),
  ('32', 'double_elim', false),
  ('32', 'double_elim', true),
  ('64', 'double_elim', false),
  ('64', 'double_elim', true);

CREATE TABLE IF NOT EXISTS tournaments(
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(128),
  description TEXT NULL,
  rules TEXT NULL,
  tournament_types_id INT,
  thread_id VARCHAR(32) NULL,
  published BOOLEAN DEFAULT false,
  starting_at DATE NULL,
  created_at BIGINT NOT NULL,
  Foreign Key (tournament_types_id) REFERENCES tournament_types(id)
);

CREATE TABLE IF NOT EXISTS attendees(
  id INT AUTO_INCREMENT PRIMARY KEY,
  tournament_id CHAR(36),
  player_id CHAR(36),
  current_seat INT NULL 
);

CREATE TABLE IF NOT EXISTS players(
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(32),
  discord_id VARCHAR(64),
  discord_avatar VARCHAR(32)
);

COMMIT;