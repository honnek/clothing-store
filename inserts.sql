--
-- PostgreSQL database dump
--

-- Dumped from database version 16.0 (Debian 16.0-1.pgdg120+1)
-- Dumped by pg_dump version 16.0 (Debian 16.0-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: cart; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.cart (id, session_id, created_at) VALUES (2, 'f1d58b99a63cab98811ce943c7f4a0f5', '2023-09-17 09:50:34');
INSERT INTO public.cart (id, session_id, created_at) VALUES (3, '9a23439d07c7c612dae351deedc50c39', '2023-09-17 12:50:05');
INSERT INTO public.cart (id, session_id, created_at) VALUES (5, '7737fe0f2478d85e5da9a1ebd7cde14c', '2023-09-17 13:18:58');
INSERT INTO public.cart (id, session_id, created_at) VALUES (6, '65a1a0cdf04cfe0fdd108f5c39c73714', '2023-09-17 13:33:55');
INSERT INTO public.cart (id, session_id, created_at) VALUES (7, '78a4bd763b5544f43c2bac136a115b40', '2023-09-17 13:38:21');
INSERT INTO public.cart (id, session_id, created_at) VALUES (8, '5f3f27848074b6f1ee936c410e4fb02d', '2023-09-17 16:45:47');
INSERT INTO public.cart (id, session_id, created_at) VALUES (9, '0350440d72600ae4dabfc971c9ca1cdd', '2023-09-18 20:11:07');
INSERT INTO public.cart (id, session_id, created_at) VALUES (11, '8835d8be1f0df103994d99bbd459adf5', '2023-09-19 11:21:53');
INSERT INTO public.cart (id, session_id, created_at) VALUES (12, '959a87ee56ee9ca935edd09c0b6b2c4a', '2023-09-19 14:51:30');
INSERT INTO public.cart (id, session_id, created_at) VALUES (13, 'cbeb5b6a6af1908a92b8b39f4ec4409c', '2023-09-19 16:56:48');
INSERT INTO public.cart (id, session_id, created_at) VALUES (14, '969ef2701591b9bf92905f0e00f744d5', '2023-09-19 21:49:31');
INSERT INTO public.cart (id, session_id, created_at) VALUES (15, 'c7576d3be226c703fbf4fcbd383f6f11', '2023-09-19 22:51:03');
INSERT INTO public.cart (id, session_id, created_at) VALUES (16, '45e231b1cb3f44ac3095416c4f5e2eaa', '2023-09-19 23:57:08');
INSERT INTO public.cart (id, session_id, created_at) VALUES (18, '2e1b6e20326c57ba7996ad1eaef7f672', '2023-09-20 19:05:21');
INSERT INTO public.cart (id, session_id, created_at) VALUES (19, '59be32ef37e6b58cfb3d33d0c730765b', '2023-09-23 14:30:09');
INSERT INTO public.cart (id, session_id, created_at) VALUES (21, '6dc7e6e8aa998d4a5dc8830974c0a7ed', '2023-09-24 21:50:02');
INSERT INTO public.cart (id, session_id, created_at) VALUES (22, 'ec525dc26a10f85ee441f6e7dd318fad', '2023-09-25 09:21:58');
INSERT INTO public.cart (id, session_id, created_at) VALUES (23, 'f3fe29cd3444f5f5f94f11002acb12c9', '2023-09-25 13:51:49');
INSERT INTO public.cart (id, session_id, created_at) VALUES (24, '53a68bf4bb6f2c95ccc2e63ff554d23b', '2023-09-25 13:54:14');
INSERT INTO public.cart (id, session_id, created_at) VALUES (26, '27970b1d1bd0aa078d8e91588d2969b1', '2023-09-25 13:55:48');
INSERT INTO public.cart (id, session_id, created_at) VALUES (27, 'ad042ac73e75b4078e619f80e1f1d242', '2023-09-25 14:18:50');
INSERT INTO public.cart (id, session_id, created_at) VALUES (28, 'eddbe24d3b302e0761d148cb25975a44', '2023-09-25 14:18:55');
INSERT INTO public.cart (id, session_id, created_at) VALUES (29, 'c1e18b5837308d8b1fcb21261199c0b6', '2023-09-26 14:44:52');
INSERT INTO public.cart (id, session_id, created_at) VALUES (31, '454fd8997c7cf204b958fdbb1def9b1e', '2023-09-27 00:34:47');
INSERT INTO public.cart (id, session_id, created_at) VALUES (40, '23297eb5204b541a9c7c4d23bc076486', '2023-11-28 10:33:48');
INSERT INTO public.cart (id, session_id, created_at) VALUES (41, '77fba7924009dfef70d1cea0f37731ec', '2023-11-28 10:33:50');
INSERT INTO public.cart (id, session_id, created_at) VALUES (42, 'b50bd67da6567055e33469e0c0d8beea', '2023-11-28 11:28:33');
INSERT INTO public.cart (id, session_id, created_at) VALUES (43, '7564497df268a1ffe6c0b24dd8746733', '2023-11-28 11:28:42');


--
-- Data for Name: category; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.category (id, title, slug, is_deleted) VALUES (4, 'Dresses', 'dresses', false);
INSERT INTO public.category (id, title, slug, is_deleted) VALUES (5, 'Hats', 'hats', false);
INSERT INTO public.category (id, title, slug, is_deleted) VALUES (6, 'Sneakers', 'sneakers', false);


--
-- Data for Name: product; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.product (id, title, price, quality, created_at, description, is_published, is_deleted, slug, uuid, category_id) VALUES (3, 'Хей, друг', 600.00, 23, '2023-09-17 13:17:56', 'qweqwe', true, false, NULL, '54d8c97e-cf9c-4028-98a6-0d1c86d761a7', 5);
INSERT INTO public.product (id, title, price, quality, created_at, description, is_published, is_deleted, slug, uuid, category_id) VALUES (4, 'Hey girl', 5000.00, 4, '2023-09-17 13:18:28', 'Priveeeyt', true, false, NULL, 'cddafb54-d5a8-4287-8611-f487c515430d', 4);
INSERT INTO public.product (id, title, price, quality, created_at, description, is_published, is_deleted, slug, uuid, category_id) VALUES (5, 'Product6', 1245.00, 23, '2023-09-25 13:55:36', 'Hey maaan', true, false, NULL, '7798144f-acd6-404d-a408-2d2e47784448', 4);
INSERT INTO public.product (id, title, price, quality, created_at, description, is_published, is_deleted, slug, uuid, category_id) VALUES (2, 'Whore', 1500.00, 5, '2023-09-17 13:17:35', 'e,kkkkrpokwer', true, false, NULL, '8eb4dc54-d128-4abe-9d05-d35db43be8c8', 6);


--
-- Data for Name: cart_product; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: doctrine_migration_versions; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230130213839', '2023-09-17 09:49:15', 5);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230131150414', '2023-09-17 09:49:15', 11);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230201161405', '2023-09-17 09:49:15', 4);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230201163957', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230202211121', '2023-09-17 09:49:15', 1);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230202211259', '2023-09-17 09:49:15', 1);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230202212046', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230206222524', '2023-09-17 09:49:15', 5);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230208123126', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230211095857', '2023-09-17 09:49:15', 5);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230211101216', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230222111337', '2023-09-17 09:49:15', 5);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230223111637', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230225204806', '2023-09-17 09:49:15', 7);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230226124537', '2023-09-17 09:49:15', 9);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230411203448', '2023-09-17 09:49:15', 4);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230418201200', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230420220529', '2023-09-17 09:49:15', 0);
INSERT INTO public.doctrine_migration_versions (version, executed_at, execution_time) VALUES ('DoctrineMigrations\Version20230910203715', '2023-09-17 09:49:15', 5);


--
-- Data for Name: user; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public."user" (id, email, roles, password, is_verified, full_name, phone, address, zipcode, is_deleted, facebook_id, google_id) VALUES (1, 'heyBroDude@dude.com', '["ROLE_ADMIN"]', '$2y$13$x0ucM2C7VxQzKd7uHYMm..QadJBi.GXasxgqIs.5moSiVaco.BKqy', true, 'Дмитрий Андреевич Махинов', '+79002776758', 'ул. Лунная, д.5 кв.1', 350906, false, NULL, NULL);


--
-- Data for Name: order; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public."order" (id, owner_id, created_at, status, total_price, updated_at, is_deleted) VALUES (1, 1, '2023-09-17 13:18:54', 0, 10000, '2023-09-17 13:20:18', false);
INSERT INTO public."order" (id, owner_id, created_at, status, total_price, updated_at, is_deleted) VALUES (2, 1, '2023-09-25 13:54:31', 0, 5000, '2023-09-25 13:54:31', false);
INSERT INTO public."order" (id, owner_id, created_at, status, total_price, updated_at, is_deleted) VALUES (3, 1, '2023-09-27 00:34:53', 0, 1500, '2023-09-27 00:34:53', false);


--
-- Data for Name: order_product; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.order_product (id, app_order_id, product_id, quantity, price_per_one) VALUES (1, 1, 4, 1, 5000.00);
INSERT INTO public.order_product (id, app_order_id, product_id, quantity, price_per_one) VALUES (2, 1, 2, 2, 1500.00);
INSERT INTO public.order_product (id, app_order_id, product_id, quantity, price_per_one) VALUES (3, 1, 3, 1, 2000.00);
INSERT INTO public.order_product (id, app_order_id, product_id, quantity, price_per_one) VALUES (4, 2, 4, 1, 5000.00);
INSERT INTO public.order_product (id, app_order_id, product_id, quantity, price_per_one) VALUES (5, 2, 5, 12, 2.00);
INSERT INTO public.order_product (id, app_order_id, product_id, quantity, price_per_one) VALUES (6, 3, 2, 1, 1500.00);


--
-- Data for Name: product_image; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.product_image (id, product_id, filename_big, filename_middle, filename_small) VALUES (2, 2, '6506d23f74d5d_big.jpg', '6506d23f71fd1_middle.jpg', '6506d23f6f91b_small.jpg');
INSERT INTO public.product_image (id, product_id, filename_big, filename_middle, filename_small) VALUES (3, 3, '6506d25473da3_big.jpg', '6506d25471e6a_middle.jpg', '6506d2547069b_small.jpg');
INSERT INTO public.product_image (id, product_id, filename_big, filename_middle, filename_small) VALUES (4, 4, '6506d274185e3_big.jpg', '6506d274131bf_middle.jpg', '6506d2741042b_small.jpg');
INSERT INTO public.product_image (id, product_id, filename_big, filename_middle, filename_small) VALUES (5, 5, '6511672847de8_big.jpg', '6511672835961_middle.jpg', '6511672826150_small.jpg');


--
-- Data for Name: reset_password_request; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: test_api_entity; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: test_table; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.test_table (id, name) VALUES (1, 'Test 1');
INSERT INTO public.test_table (id, name) VALUES (2, 'Test 2');
INSERT INTO public.test_table (id, name) VALUES (3, 'Test 3');


--
-- Name: cart_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cart_id_seq', 46, true);


--
-- Name: cart_product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cart_product_id_seq', 4, true);


--
-- Name: category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.category_id_seq', 7, true);


--
-- Name: order_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_id_seq', 3, true);


--
-- Name: order_product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_product_id_seq', 6, true);


--
-- Name: product_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_id_seq', 5, true);


--
-- Name: product_image_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_image_id_seq', 5, true);


--
-- Name: reset_password_request_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reset_password_request_id_seq', 1, false);


--
-- Name: test_api_entity_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.test_api_entity_id_seq', 1, false);


--
-- Name: test_table_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.test_table_id_seq', 3, true);


--
-- Name: user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_id_seq', 1, true);


--
-- PostgreSQL database dump complete
--

