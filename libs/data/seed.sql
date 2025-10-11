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

INSERT INTO users (name, email, avatar, dealership_id, role)
VALUES (
    'Aaron Santo',
    'santoaaron@gmail.com',
    'https://ui-avatars.com/api/?name=Aaron+Santo&background=BAFFC9',
    1,
    'admin'
);

INSERT INTO users (name, email, avatar, dealership_id, role)
VALUES (
    'Aaron Santo',
    'apenguinemail@gmail.com',
    'https://ui-avatars.com/api/?name=Aaron+Santo&background=BAFFC9',
    1,
    'user'
);

INSERT INTO catalog_items (created_at)
VALUES (
    now()
);
