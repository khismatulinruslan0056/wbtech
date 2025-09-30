-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS orders (
    ID TEXT UNIQUE PRIMARY KEY NOT NULL,
    TrackNumber TEXT UNIQUE NOT NULL,
    Entry TEXT,
    Locale VARCHAR(2) NOT NULL,
    InternalSignature TEXT,
    CustomerID TEXT,
    DeliveryService TEXT NOT NULL ,
    Shardkey TEXT NOT NULL,
    SmID INTEGER NOT NULL,
    DateCreated TIMESTAMPTZ NOT NULL DEFAULT now(),
    OofShard TEXT
);

CREATE TABLE IF NOT EXISTS payments
(
    ID SERIAL NOT NULL PRIMARY KEY,
    OrderID TEXT UNIQUE NOT NULL REFERENCES orders(ID) ON DELETE CASCADE,
    Transaction TEXT UNIQUE NOT NULL,
    RequestID TEXT NOT NULL,
    Currency VARCHAR(3) NOT NULL,
    Provider TEXT NOT NULL,
    Amount BIGINT NOT NULL,
    Payment_dt INTEGER NOT NULL,
    Bank TEXT NOT NULL,
    DeliveryCost INTEGER NOT NULL,
    GoodsTotal INTEGER NOT NULL,
    CustomFee INTEGER NOT NULL,

    CONSTRAINT payments_orderid_eq_transaction CHECK (transaction = orderid)

);

CREATE TABLE IF NOT EXISTS deliveries (
    ID SERIAL PRIMARY KEY,
    OrderID TEXT UNIQUE NOT NULL REFERENCES orders(ID) ON DELETE CASCADE,
    Name  TEXT NOT NULL,
    Phone TEXT NOT NULL,
    Zip TEXT NOT NULL,
    City TEXT NOT NULL,
    Address TEXT NOT NULL,
    Region TEXT NOT NULL,
    Email TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS items
(
    ID SERIAL PRIMARY KEY,
    OrderID TEXT NOT NULL REFERENCES orders(ID) ON DELETE CASCADE,
    ChrtID BIGINT NOT NULL,
    Price INTEGER NOT NULL,
    RID TEXT NOT NULL,
    Name TEXT NOT NULL,
    Sale INTEGER NOT NULL,
    Size TEXT NOT NULL,
    TotalPrice INTEGER NOT NULL,
    NmID INTEGER NOT NULL,
    Brand TEXT NOT NULL,
    Status INTEGER NOT NULL,

    UNIQUE (OrderID, ChrtID)
);

CREATE INDEX idx_orders_customer_id ON orders (CustomerID);

CREATE INDEX idx_items_order_id ON items (OrderID);

CREATE INDEX idx_items_nmid ON items (NmID);
CREATE INDEX idx_items_chrtid ON items (ChrtID);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_customer_id;
DROP INDEX IF EXISTS idx_items_order_id;
DROP INDEX IF EXISTS idx_items_nmid;
DROP INDEX IF EXISTS idx_items_chrtid;

DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS orders;
