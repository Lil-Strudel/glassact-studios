INSERT INTO dealerships (name, street, street_ext, city, state, postal_code, country, location)
VALUES (
    'Fake Dealership',
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

-- Initial grout (background) set — placeholder granite tones, refine later.
INSERT INTO grouts (name, hex, sort_order) VALUES
  ('Black Granite', '#1c1c1c', 10),
  ('Dark Grey Granite', '#3b3e40', 20),
  ('Grey Granite', '#8a8d8f', 30),
  ('Light Grey Granite', '#c7c9c8', 40),
  ('Mahogany Granite', '#4a1f1a', 50),
  ('Rose Granite', '#b58a86', 60),
  ('Tan Granite', '#c9b79c', 70),
  ('Green Granite', '#2d3b33', 80)
ON CONFLICT (hex) DO NOTHING;

-- Support / knowledge-base content shown on the Support page.
INSERT INTO support_articles (category, title, body, youtube_url, sort_order) VALUES
  (
    'installation',
    'Installing a stained glass inlay',
    E'Watch the walkthrough above, then follow these steps:\n\n1. **Clean the recess** thoroughly and let it dry.\n2. Dry-fit the inlay to confirm the depth and orientation.\n3. Apply a bead of the recommended **adhesive** around the perimeter.\n4. Seat the inlay, press evenly, and wipe away any squeeze-out.\n5. Let it cure undisturbed for **24 hours** before handling.',
    'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
    10
  ),
  (
    'installation',
    'Cold-weather installation tips',
    E'Adhesives cure slowly below 50°F. When installing in cold conditions:\n\n- Store the adhesive at room temperature overnight before use.\n- Warm the stone surface with a heat lamp if possible.\n- Allow extra cure time — up to **48 hours**.',
    NULL,
    20
  ),
  (
    'ordering',
    'How to place an order',
    E'1. Create a **project** from the Projects page.\n2. Add one or more **inlays** — pick a catalog design or request a custom piece.\n3. For catalog items you can customize colors and sizing in the customizer.\n4. Once every inlay is marked **ready**, open the cart and select the inlays to include.\n5. Click **Place Order** — pricing is locked in at this point.',
    NULL,
    10
  ),
  (
    'pricing',
    'How pricing works',
    E'Every inlay is priced by its **price group**. A catalog item has a default price group, and our designers may adjust it based on custom sizing, added colors, or special materials.\n\nThe price is locked when you place your order, so later catalog changes never affect an existing order. See the current price groups below.',
    NULL,
    10
  ),
  (
    'contact',
    'Get in touch',
    E'Still have questions? We are happy to help.\n\n- **Email:** support@glassactstudios.com\n- **Phone:** (555) 123-4567, Mon–Fri 8am–5pm ET\n\nFor order-specific questions, include your project name or reference number so we can look it up quickly.',
    NULL,
    10
  )
ON CONFLICT DO NOTHING;
