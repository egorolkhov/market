-- +goose Up
CREATE TABLE Users (
                       id UUID PRIMARY KEY,
                       username VARCHAR(50) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       balance INT NOT NULL DEFAULT 1000,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Merchandise (
                             id SERIAL PRIMARY KEY,
                             name VARCHAR(100) UNIQUE NOT NULL,
                             price INT NOT NULL
);

-- Вставляем товары (INSERT) сразу после создания таблицы Merchandise
INSERT INTO Merchandise (name, price) VALUES
                                          ('t-shirt', 80),
                                          ('cup', 20),
                                          ('book', 50),
                                          ('pen', 10),
                                          ('powerbank', 200),
                                          ('hoody', 300),
                                          ('umbrella', 200),
                                          ('socks', 10),
                                          ('wallet', 50),
                                          ('pink-hoody', 500);

CREATE TABLE User_Inventory (
                                id SERIAL PRIMARY KEY,
                                user_id UUID REFERENCES Users(id),
                                merchandise_id INT REFERENCES Merchandise(id),
                                quantity INT NOT NULL
);

CREATE TABLE Transactions (
                              id SERIAL PRIMARY KEY,
                              sender_id UUID REFERENCES Users(id),
                              receiver_id UUID REFERENCES Users(id),
                              amount INT NOT NULL,
                              created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_sender_id
    ON Transactions (sender_id);

CREATE INDEX idx_transactions_receiver_id
    ON Transactions (receiver_id);

-- +goose Down
DROP TABLE IF EXISTS Transactions;
DROP TABLE IF EXISTS User_Inventory;
DROP TABLE IF EXISTS Users;
DROP TABLE IF EXISTS Merchandise;
