-- +goose Up
INSERT INTO orders ( ID, TrackNumber, Entry, Locale, InternalSignature, CustomerID,
                    DeliveryService, Shardkey, SmID, DateCreated, OofShard)
VALUES (
             'b563feb7b2b84b6test',
             'WBILMTESTTRACK',
             'WBIL',
             'en',
             '',
             'test',
             'meest',
             '9',
             99,
             '2021-11-26T06:22:19Z',
             '1'
         );

INSERT INTO payments ( OrderID, Transaction, RequestID, Currency, Provider, Amount, Payment_dt,
                      Bank, DeliveryCost, GoodsTotal, CustomFee)
VALUES (
             'b563feb7b2b84b6test',
             'b563feb7b2b84b6test',
             '',
             'USD',
             'wbpay',
             1817,
             1637907727,
             'alpha',
             1500,
             317,
             0
         );

INSERT INTO deliveries ( OrderID, Name, Phone, Zip, City, Address, Region, Email)
VALUES (
             'b563feb7b2b84b6test',
             'Test Testov',
             '+9720000000',
             '2639809',
             'Kiryat Mozkin',
             'Ploshad Mira 15',
             'Kraiot',
             'test@gmail.com'
         );

INSERT INTO items ( OrderID, ChrtID, Price, RID, Name, Sale, Size, TotalPrice, NmID, Brand, Status)
VALUES (
             'b563feb7b2b84b6test',
             9934930,
             453,
             'ab4219087a764ae0btest',
             'Mascaras',
             30,
             '0',
             317,
             2389212,
             'Vivienne Sabo',
             202
         );

-- +goose Down
DELETE FROM items WHERE OrderID = 'b563feb7b2b84b6test';
DELETE FROM deliveries WHERE OrderID = 'b563feb7b2b84b6test';
DELETE FROM payments WHERE OrderID = 'b563feb7b2b84b6test';
DELETE FROM orders WHERE ID = 'b563feb7b2b84b6test';
