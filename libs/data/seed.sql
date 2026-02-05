INSERT INTO dealerships (name, street, street_ext, city, state, postal_code, country, location)
VALUES (
    'GlassAct Studios',
    '540 S Commerce Rd',
    '',
    'Orem',
    'UT',
    '84058',
    'US',
    ST_SetSRID(ST_MakePoint(-111.72878560766672, 40.28727777344243), 4326)::GEOGRAPHY);

INSERT INTO internal_users (name, email, avatar, role)
VALUES (
    'Aaron Santo',
    'santoaaron@gmail.com',
    'https://ui-avatars.com/api/?name=Aaron+Santo&background=BAFFC9',
    'admin'
);

INSERT INTO internal_users (name, email, avatar, role)
VALUES (
    'T8 Storey',
    't8storey@protonmail.com',
    'https://ui-avatars.com/api/?name=T8+Storey&background=BAFFC9',
    'admin'
);

INSERT INTO dealership_users (dealership_id, name, email, avatar, role)
VALUES (
    1,
    'Aaron Santo',
    'apenguinemail@gmail.com',
    'https://ui-avatars.com/api/?name=Aaron+Santo&background=BAFFC9',
    'admin'
);
