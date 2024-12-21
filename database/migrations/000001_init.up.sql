BEGIN;

USE spade;

CREATE TABLE IF NOT EXISTS tournament_types(
  id INT AUTO_INCREMENT PRIMARY KEY,
  participants INT,
  bracket_type ENUM('single_elim', 'double_elim'),
  has_third_winner BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS tournaments(
  id CHAR(36) PRIMARY KEY,
  size ENUM('2', '4', '8', '16', '32', '64'),
  tournament_types_id INT,
  starting_at DATE,
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