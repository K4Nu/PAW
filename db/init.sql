-- Usuwamy stare tabele jeśli istnieją (ważne przy odpalaniu w dev)
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS unit_types;
DROP TABLE IF EXISTS buildings;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS villages;
DROP TABLE IF EXISTS users;

-- ===========================
-- Tabela użytkowników
-- ===========================
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(50) UNIQUE NOT NULL,
                       email VARCHAR(100) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       role VARCHAR(20) DEFAULT 'player',
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================
-- Tabela wiosek
-- ===========================
CREATE TABLE villages (
                          id SERIAL PRIMARY KEY,
                          user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          name VARCHAR(100) NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================
-- Tabela zasobów
-- ===========================
CREATE TABLE resources (
                           id SERIAL PRIMARY KEY,
                           village_id INT NOT NULL REFERENCES villages(id) ON DELETE CASCADE,
                           wood INT DEFAULT 100,
                           clay INT DEFAULT 100,
                           iron INT DEFAULT 100,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================
-- Tabela budynków
-- ===========================
CREATE TABLE buildings (
                           id SERIAL PRIMARY KEY,
                           village_id INT NOT NULL REFERENCES villages(id) ON DELETE CASCADE,
                           type VARCHAR(50) NOT NULL,   -- np. 'townhall', 'lumbermill', 'claypit', 'ironmine', 'warehouse', 'barracks'
                           level INT DEFAULT 1,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================
-- Tabela jednostek w wiosce
-- ===========================
CREATE TABLE units (
                       id SERIAL PRIMARY KEY,
                       village_id INT NOT NULL REFERENCES villages(id) ON DELETE CASCADE,
                       type VARCHAR(50) NOT NULL,  -- np. 'spearman', 'swordsman', 'cavalry'
                       count INT DEFAULT 0,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================
-- Tabela typów jednostek (koszty, czas szkolenia)
-- ===========================
CREATE TABLE unit_types (
                            type VARCHAR(50) PRIMARY KEY,
                            wood INT,
                            clay INT,
                            iron INT,
                            training_time INT -- w sekundach
);

-- Domyślne typy jednostek
INSERT INTO unit_types (type, wood, clay, iron, training_time) VALUES
                                                                   ('spearman', 50, 30, 20, 30),
                                                                   ('swordsman', 100, 50, 50, 60),
                                                                   ('cavalry', 200, 100, 150, 120);
