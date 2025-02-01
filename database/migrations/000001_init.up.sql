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
  starting_at BIGINT NULL,
  created_at BIGINT NOT NULL,
  FOREIGN KEY (tournament_types_id) REFERENCES tournament_types(id)
);

CREATE TABLE IF NOT EXISTS players(
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(32),
  discord_id VARCHAR(64) UNIQUE
);

CREATE TABLE IF NOT EXISTS attendees(
  id INT AUTO_INCREMENT PRIMARY KEY,
  tournament_id CHAR(36),
  player_id CHAR(36),
  starting_seat INT NULL,
  current_seat INT NULL,
  FOREIGN KEY (tournament_id) REFERENCES tournaments(id),
  FOREIGN KEY (player_id) REFERENCES players(id),
  INDEX idx_tournament_id (tournament_id)
);

CREATE TABLE IF NOT EXISTS match_histories(
  id INT AUTO_INCREMENT PRIMARY KEY,
  attendee_id INT,
  result TINYINT DEFAULT 0,
  seat INT NULL,
  created_at BIGINT NULL,
  FOREIGN KEY (attendee_id) REFERENCES attendees(id),
  INDEX idx_attendee_id (attendee_id)
);

COMMIT;