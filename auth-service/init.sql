CREATE TABLE IF NOT EXISTS users (
    ID BIGSERIAL NOT NULL,
    Email VARCHAR NOT NULL UNIQUE,
    FirstName VARCHAR NOT NULL,
    LastName VARCHAR NOT NULL,
    Password VARCHAR NOT NULL,
    IIN VARCHAR NOT NULL UNIQUE,
    Phone VARCHAR NOT NULL UNIQUE,
    Registered TIMESTAMP NOT NULL,
    Role VARCHAR NOT NULL,
    CONSTRAINT users_pk PRIMARY KEY (ID)
);