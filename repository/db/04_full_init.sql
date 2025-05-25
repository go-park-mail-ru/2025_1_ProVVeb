-- INSERT INTO interests (description)
-- SELECT 'Интерес #' || gs
-- FROM generate_series(1, 100) AS gs;

-- INSERT INTO preferences (preference_type, preference_description, preference_value)
-- SELECT 
--   (RANDOM() * 5 + 1)::INT,
--   'Предпочтение #' || gs,
--   'Значение ' || gs
-- FROM generate_series(1, 100) AS gs;


-- INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, location_id)
-- SELECT 
--   'Имя_' || gs,
--   'Фамилия_' || gs,
--   (RANDOM() > 0.5)::BOOLEAN,
--   DATE '1980-01-01' + (random() * 15000)::INT,
--   (random() * 100 + 150)::INT, 
--   'Описание пользователя #' || gs,
--   NULL  
-- FROM generate_series(1, 100000) AS gs;


-- DO $$
-- DECLARE
--     i BIGINT;
--     interest_id INT;
--     count INT;
-- BEGIN
--     FOR i IN 1..100000 LOOP
--         count := (random() * 4 + 1)::INT;
--         FOR interest_id IN 1..count LOOP
--             BEGIN
--                 INSERT INTO profile_interests (profile_id, interest_id)
--                 VALUES (i, (random() * 99 + 1)::INT)
--                 ON CONFLICT DO NOTHING;
--             EXCEPTION WHEN OTHERS THEN
--                 CONTINUE;
--             END;
--         END LOOP;
--     END LOOP;
-- END $$;


-- DO $$
-- DECLARE
--     i BIGINT;
--     preference_id INT;
--     count INT;
-- BEGIN
--     FOR i IN 1..100000 LOOP
--         count := (random() * 2 + 1)::INT; 
--         FOR preference_id IN 1..count LOOP
--             BEGIN
--                 INSERT INTO profile_preferences (profile_id, preference_id)
--                 VALUES (i, (random() * 99 + 1)::INT)
--                 ON CONFLICT DO NOTHING;
--             EXCEPTION WHEN OTHERS THEN
--                 CONTINUE;
--             END;
--         END LOOP;
--     END LOOP;
-- END $$;
