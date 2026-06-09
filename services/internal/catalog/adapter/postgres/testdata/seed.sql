INSERT INTO category (id, title, slug) VALUES
    (1, 'Hats', 'hats'),
    (2, 'Shoes', 'shoes');

INSERT INTO product (id, uuid, title, price, quality, created_at, description, is_published, is_deleted, slug, category_id) VALUES
    (1, '11111111-1111-1111-1111-111111111111', 'Red Hat',  '9.99',  5, now(), 'nice', true,  false, 'red-hat',  1),
    (2, '22222222-2222-2222-2222-222222222222', 'Blue Hat', '12.50', 4, now(), NULL,   true,  false, 'blue-hat', 1),
    (3, '33333333-3333-3333-3333-333333333333', 'Old Shoe', '5.00',  2, now(), NULL,   false, false, 'old-shoe', 2),
    (4, '44444444-4444-4444-4444-444444444444', 'Gone',     '1.00',  1, now(), NULL,   true,  true,  'gone',     2);
