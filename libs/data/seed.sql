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
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    1,
    'a9fc472f-f3c7-4957-afa8-fe5f9f85a669',
    'PG-1',
    10000,
    '',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    2,
    '1bb163a1-7818-4e76-84eb-944701df5f61',
    'PG-2',
    15000,
    '',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    3,
    '3a050196-1a60-4a0c-97c9-883df0e792c4',
    'PG-3',
    20000,
    '',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    4,
    '1ec26898-feea-43f8-a1ae-d62984a6eec1',
    'PG-4',
    25000,
    '',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO catalog_items (id, uuid, catalog_code, name, description, category, default_width, default_height, min_width, min_height, default_price_group_id, svg_url, is_active)
VALUES (
    1,
    'aecf206b-d2c0-46ea-9c24-33a94c197ad1',
    'A-TUR-0001',
    'Turtle',
    'Hawaiian style turtle',
    'Animal',
    2,
    2,
    1,
    1,
    2,
    '/file/catalog-items/18aac212-af87-4fbc-93e6-17c27bd7f4f4.svg',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO catalog_items (id, uuid, catalog_code, name, description, category, default_width, default_height, min_width, min_height, default_price_group_id, svg_url, is_active)
VALUES (
    2,
    'ffab07c0-3cb1-4e8a-a55a-d612a709ad04',
    'A-LZD-0001',
    'Lizard',
    '',
    'Animal',
    2,
    2,
    1,
    1,
    1,
    '/file/catalog-items/5cae9a3a-9fef-477a-a3cf-3cfbbf05e0b1.svg',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO catalog_items (id, uuid, catalog_code, name, description, category, default_width, default_height, min_width, min_height, default_price_group_id, svg_url, is_active)
VALUES (
    3,
    'b88cfb41-c4a1-4b07-b8da-bc290191b734',
    'A-HRS-0003',
    'Horse in Horseshoe',
    '',
    'Animal',
    2,
    2,
    1,
    1,
    3,
    '/file/catalog-items/848b2456-faa5-4291-9d6e-2e25ef1dde2c.svg',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO catalog_items (id, uuid, catalog_code, name, description, category, default_width, default_height, min_width, min_height, default_price_group_id, svg_url, is_active)
VALUES (
    4,
    '448a7113-b371-4f6a-956f-6e63738d9a03',
    'A-HRS-0002',
    'White Horse',
    'White horse',
    'Animal',
    3,
    3,
    1,
    1,
    1,
    '/file/catalog-items/3669cf1c-a366-4892-90c9-579ecdb63819.svg',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO catalog_items (id, uuid, catalog_code, name, description, category, default_width, default_height, min_width, min_height, default_price_group_id, svg_url, is_active)
VALUES (
    5,
    '94bd4617-b515-4530-83bc-e259e15d1cb1',
    'A-HRS-0001',
    'Brown Horse',
    'Brown Horse',
    'Animal',
    3,
    3,
    1,
    1,
    3,
    '/file/catalog-items/7b20a091-ee77-4822-a2b1-4791879aa9cb.svg',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO catalog_item_tags (catalog_item_id, tag)
VALUES (1, 'Turtle') ON CONFLICT DO NOTHING;

INSERT INTO catalog_item_tags (catalog_item_id, tag)
VALUES (2, 'Lizard') ON CONFLICT DO NOTHING;

INSERT INTO catalog_item_tags (catalog_item_id, tag)
VALUES (3, 'Horse') ON CONFLICT DO NOTHING;

INSERT INTO catalog_item_tags (catalog_item_id, tag)
VALUES (4, 'Horse') ON CONFLICT DO NOTHING;

INSERT INTO catalog_item_tags (catalog_item_id, tag)
VALUES (5, 'Horse') ON CONFLICT DO NOTHING;
