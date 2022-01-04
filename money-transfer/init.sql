CREATE TABLE IF NOT EXISTS accounts (
    ID BIGSERIAL NOT NULL PRIMARY KEY,
    OwnerID BIGSERIAL NOT NULL,
    IIN VARCHAR NOT NULL,
    Amount BIGSERIAL NOT NULL,
    Registered TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    ID BIGSERIAL NOT NULL PRIMARY KEY,
    SenderID BIGSERIAL NOT NULL,
    ReceiverID BIGSERIAL NOT NULL,
    Amount BIGSERIAL NOT NULL,
    Date TIMESTAMP NOT NULL,
    FOREIGN KEY (SenderID) REFERENCES accounts (ID),
    FOREIGN KEY (ReceiverID) REFERENCES accounts (ID)
);

INSERT INTO accounts(ID, OwnerID, IIN, Amount, Registered)
VALUES (4405211239547816, 999, '921115350186', 1000000000, '2022-01-03 11:51:40.244153')
ON CONFLICT DO NOTHING;