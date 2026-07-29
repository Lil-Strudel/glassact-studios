INSERT INTO dealerships (name, street, street_ext, city, state, postal_code, country, location)
VALUES (
    'GlassAct Test Dealership',
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
    'Roy Santo',
    'roy@glassactstudios.com',
    'https://ui-avatars.com/api/?name=Roy+Santo&background=BAFFC9',
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

INSERT INTO dealership_users (dealership_id, name, email, avatar, role)
VALUES (
    1,
    'Roy Santo',
    'roysanto@yahoo.com',
    'https://ui-avatars.com/api/?name=Roy+Santo&background=BAFFC9',
    'admin'
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    1,
    'a9fc472f-f3c7-4957-afa8-fe5f9f85a669',
    'PG-1',
    10000,
    'Standard pricing for smaller, simpler designs.',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    2,
    '1bb163a1-7818-4e76-84eb-944701df5f61',
    'PG-2',
    18500,
    'Pricing for moderately sized or detailed designs.',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    3,
    '3a050196-1a60-4a0c-97c9-883df0e792c4',
    'PG-3',
    29000,
    'Pricing for larger or more elaborate designs.',
    true
) ON CONFLICT DO NOTHING;

INSERT INTO price_groups (id, uuid, name, base_price_cents, description, is_active)
VALUES (
    4,
    '1ec26898-feea-43f8-a1ae-d62984a6eec1',
    'PG-4',
    60000,
    'Pricing for our largest and most complex designs.',
    true
) ON CONFLICT DO NOTHING;

-- Glass colors GlassAct offers (extracted from the master swatch chart).
INSERT INTO glass_colors (name, hex, family, sort_order) VALUES
  ('Charcoal', '#4e4a42', 'neutral', 10),
  ('Pale Grey', '#7b8074', 'neutral', 20),
  ('Black', '#010101', 'neutral', 30),
  ('Cloud', '#c1cdca', 'neutral', 40),
  ('Almond', '#fbf8df', 'neutral', 50),
  ('Ivory', '#ebebdb', 'neutral', 60),
  ('White', '#ffffff', 'neutral', 70),
  ('Steel Blue', '#003f5a', 'blue', 80),
  ('Medium Blue', '#415ba8', 'blue', 90),
  ('Mariner', '#2a5b8f', 'blue', 100),
  ('Cobalt Blue', '#2b3278', 'blue', 110),
  ('Dark Blue', '#03569c', 'blue', 120),
  ('Alpine Blue', '#3f8b9d', 'blue', 130),
  ('Riviera Blue', '#4d90cd', 'blue', 140),
  ('Turquoise Blue', '#0d9bc7', 'blue', 150),
  ('Moss Green', '#23772d', 'green', 160),
  ('Celadon', '#9ea879', 'green', 170),
  ('Amazon Green', '#a1bd3a', 'green', 180),
  ('Pastel Green', '#9fd5b9', 'green', 190),
  ('Olive', '#0c2b2e', 'green', 200),
  ('Dark Green', '#154c3e', 'green', 210),
  ('Turquoise Green', '#15b9b0', 'green', 220),
  ('Peacock Green', '#027d71', 'green', 230),
  ('Lilac', '#7b6269', 'purple', 240),
  ('Mauve', '#ddc3cb', 'purple', 250),
  ('Violet', '#2a1423', 'purple', 260),
  ('Plum', '#46222f', 'purple', 270),
  ('Pale Purple', '#c68c85', 'purple', 280),
  ('Antique Bronze', '#543a23', 'brown', 290),
  ('Chestnut', '#7d4e2f', 'brown', 300),
  ('Terra Cotta', '#ae6219', 'brown', 310),
  ('Bronze', '#633d2a', 'brown', 320),
  ('Champagne', '#f9c996', 'amber', 330),
  ('Dark Champagne', '#fbd0a0', 'amber', 340),
  ('Marigold', '#faab54', 'amber', 350),
  ('Sunflower', '#eeb211', 'amber', 360),
  ('Orange', '#f15f25', 'red', 370),
  ('Persimmon', '#f37f43', 'red', 380),
  ('Red', '#910028', 'red', 390),
  ('Pink', '#e09090', 'red', 400)
ON CONFLICT (hex) DO NOTHING;

-- Grout (background) colors GlassAct offers.
INSERT INTO grouts (name, hex, sort_order) VALUES
  ('Raven', '#1a1a1a', 10),
  ('Summer Wheat', '#5f4d40', 20),
  ('DeLorean Gray', '#92918a', 30),
  ('Light Chocolate', '#9e7d62', 40),
  ('Light Buff', '#bdac9f', 50),
  ('White', '#e2e0df', 60)
ON CONFLICT (hex) DO NOTHING;
