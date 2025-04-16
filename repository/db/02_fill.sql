INSERT INTO complaint_types (type_description) VALUES 
('Неприемлемый контент'),
('Домогательства'),
('Спам'),
('Ложный профиль'),
('Оскорбительный язык');

INSERT INTO notification_types (type_description) VALUES 
('Новое сообщение'),
('Новый лайк'),
('Новый матч'),
('Истечение подписки'),
('Просмотр профиля');

INSERT INTO locations (country, city, district) VALUES 
('США', 'Нью-Йорк', 'Манхэттен'),
('Канада', 'Торонто', 'Центр города'),
('Великобритания', 'Лондон', 'Центральный Лондон'),
('Германия', 'Берлин', 'Митте'),
('Франция', 'Париж', 'Иль-де-Франс');

INSERT INTO subscription_types (type_description) VALUES 
('Базовая'),
('Премиум'),
('Золотая'),
('Платиновая');

INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, location_id) VALUES
('Иван', 'Иванов', TRUE, '1990-05-15', 180, 'Программист', 1),
('Анна', 'Петрова', FALSE, '1988-07-20', 165, 'Дизайнер', 2),
('Ольга', 'Смирнова', FALSE, '1995-03-22', 170, 'Маркетолог', 3),
('Дмитрий', 'Кузнецов', TRUE, '1992-11-30', 175, 'Инженер', 4),
('Александр', 'Попов', TRUE, '1991-09-25', 185, 'Фотограф', 5),
('Максим', 'Орлов', TRUE, '1989-08-10', 182, 'Менеджер', 1),
('Инна', 'Лисова', FALSE, '1993-04-12', 168, 'Юрист', 2),
('Леонид', 'Захаров', TRUE, '1987-01-05', 178, 'Врач', 3),
('Нина', 'Морозова', FALSE, '1996-12-30', 160, 'Писатель', 4),
('Вера', 'Соколова', FALSE, '1990-10-15', 172, 'Актриса', 5);


INSERT INTO users (profile_id, status, login, email, phone, password) VALUES
(1, 1, 'ivanivanov', 'ivan@example.com', '1111111111', 'f40331940a2eb0ec842f09e39726335747f04e7cf3d3a56ea633f3b72ef2458d'),
(2, 1, 'annapetrova', 'anna@example.com', '1111111112', 'eda6e9d3f2ff853757a9f864b8c793f5700d794f410252f1222b576e56b67983'),
(3, 1, 'olgasmirnova', 'olga@example.com', '1111111113', 'ad98820e33eded62823c5abfc1db9780846d014b11d07eb28ac3186d68e15486'),
(4, 1, 'dmitrykuznetsov', 'dmitriy@example.com', '1111111114', '5bf6e3a85f0a1406b5602e21e0aad1369bc1ff57389c2f6174b2606db7b388ca'),
(5, 1, 'alexpopov', 'alex@example.com', '1111111115', 'fbdd78794b1dbc3864c7bcbefb7c90f5c4ca5a1f409f9db1d492ffbeec6fb119'),
(6, 1, 'maksorlov', 'maks@example.com', '1111111116', '78226589b60338ee48a4c411b34c9b93a3ee5a658ec8ea039e63a0e7c1274638'),
(7, 1, 'innalisova', 'inna@example.com', '1111111117', '07e0c313ec2ea761e366fbda9dbdba9c903da501962ff298bbf050500e27be53'),
(8, 1, 'leozaharov', 'leo@example.com', '1111111118', '0e03f2a800342f4bab3bb0ca88db5d37779f306898caf6ccf9d509c293f36f57'),
(9, 1, 'ninamoro', 'nina@example.com', '1111111119', '2ee3914dee3e33b53083024874f141ce7050a11afa21cb2f04378642da9ad0a7'),
(10, 1, 'verasokolova', 'vera@example.com', '1111111120', '54c32671705cc23909c4a932eca574dbd2b9d8eeaf0a20cc3e614cf506bee0ef');


INSERT INTO static (profile_id, path) VALUES
(1, '/default.png'),
(2, '/eva.png'),
(3, '/katya.png'),
(4, '/default.png'),
(5, '/default.png'),
(6, '/default.png'),
(7, '/default.png'),
(8, '/default.png'),
(9, '/default.png'),
(10, '/default.png');


INSERT INTO sessions (user_id, token, expires_at) VALUES
(1, 'token1', '2025-12-31 23:59:59'),
(2, 'token2', '2025-12-31 23:59:59'),
(3, 'token3', '2025-12-31 23:59:59'),
(4, 'token4', '2025-12-31 23:59:59'),
(5, 'token5', '2025-12-31 23:59:59'),
(6, 'token6', '2025-12-31 23:59:59'),
(7, 'token7', '2025-12-31 23:59:59'),
(8, 'token8', '2025-12-31 23:59:59'),
(9, 'token9', '2025-12-31 23:59:59'),
(10, 'token10', '2025-12-31 23:59:59');

INSERT INTO interests (description) VALUES
('Путешествия'), ('Чтение'), ('Музыка'), ('Спорт'), ('Фильмы'),
('Готовка'), ('Фотография'), ('Танцы'), ('Программирование'), ('Искусство');


INSERT INTO profile_interests (profile_id, interest_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5),
(2, 6), (2, 7), (2, 1), (2, 2), (2, 3),
(3, 4), (3, 5), (3, 6), (3, 7), (3, 8),
(4, 9), (4, 10), (4, 1), (4, 2), (4, 3),
(5, 4), (5, 5), (5, 6), (5, 7), (5, 8),
(6, 1), (6, 2), (6, 3), (6, 4), (6, 5),
(7, 6), (7, 7), (7, 8), (7, 9), (7, 10),
(8, 1), (8, 3), (8, 5), (8, 7), (8, 9),
(9, 2), (9, 4), (9, 6), (9, 8), (9, 10),
(10, 1), (10, 2), (10, 3), (10, 4), (10, 5);



INSERT INTO preferences (preference_type, preference_description, preference_value) VALUES
(1, 'bodyType', 'Атлетическое'),
(1, 'hairColor', 'Блондин'),
(1, 'eyeColor', 'Голубые'),
(1, 'tattoo', 'false'),
(1, 'smoking', 'false'),
(1, 'education', 'Высшее'),
(1, 'nationality', 'Русский'),
(1, 'hairColor', 'Русый'),
(1, 'eyeColor', 'Карие'),
(1, 'tattoo', 'true'),
(1, 'smoking', 'true'),
(1, 'bodyType', 'Худощавое'),
(1, 'education', 'Среднее'),
(1, 'nationality', 'Украинец');

INSERT INTO profile_preferences (profile_id, preference_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5),
(2, 6), (2, 7), (2, 8), (2, 9), (2, 10),
(3, 11), (3, 12), (3, 13), (3, 1), (3, 2),
(4, 3), (4, 4), (4, 5), (4, 6), (4, 7),
(5, 8), (5, 9), (5, 10), (5, 11), (5, 12),
(6, 1), (6, 3), (6, 5), (6, 7), (6, 9),
(7, 2), (7, 4), (7, 6), (7, 8), (7, 10),
(8, 11), (8, 13), (8, 1), (8, 3), (8, 5),
(9, 7), (9, 9), (9, 11), (9, 13), (9, 2),
(10, 4), (10, 6), (10, 8), (10, 10), (10, 12);


INSERT INTO messages (sender_profile_id, receiver_profile_id, content, status) VALUES 
(1, 2, 'Привет! Как дела?', 1),
(2, 3, 'Давно не виделись!', 1),
(3, 4, 'Привет! Чем занимаешься?', 1),
(4, 5, 'Как твои дела?', 1),
(5, 1, 'Ты очень интересный человек!', 1);

INSERT INTO likes (profile_id, liked_profile_id, status) VALUES
(1, 2, 1), 
(2, 3, 1), 
(3, 4, 1), 
(4, 5, 1),
(5, 6, 1), 
(6, 7, 1), 
(7, 8, 1), 
(8, 9, 1), 
(9, 10, 1), 
(10, 1, 1);

INSERT INTO likes (profile_id, liked_profile_id, status) VALUES
(1, 3, 2),
(2, 4, 2),
(3, 5, 2),
(4, 6, 2),
(5, 7, 2),
(6, 8, 2),
(7, 9, 2),
(8, 10, 2),
(9, 1, 2),
(10, 2, 2);


INSERT INTO matches (profile_id, matched_profile_id) VALUES
(10, 2), (3, 4);


INSERT INTO subscriptions (user_id, sub_type, transaction_data) VALUES 
(1, 1, 'Подписка базовая'),
(2, 2, 'Подписка премиум'),
(3, 3, 'Подписка золотая'),
(4, 4, 'Подписка платиновая'),
(5, 1, 'Подписка базовая');

INSERT INTO complaints (complaint_by, complaint_on, complaint_type, complaint_text, status) VALUES 
(1, 2, 1, 'Неприемлемое поведение', 1),
(2, 3, 2, 'Домогательства', 1),
(3, 4, 3, 'Спам', 1),
(4, 5, 4, 'Ложный профиль', 1),
(5, 1, 5, 'Оскорбительный язык', 1);

INSERT INTO blacklist (user_id) VALUES 
(5);

INSERT INTO notifications (user_id, notification_type, content) VALUES 
(1, 1, 'Новое сообщение от пользователя 2'),
(2, 2, 'Пользователь 1 поставил вам лайк'),
(3, 3, 'Пользователь 4 поставил вам лайк'),
(4, 4, 'Ваша подписка истекает через 3 дня'),
(5, 5, 'Ваш профиль был просмотрен пользователем 1');

INSERT INTO profile_ratings (profile_id, rated_profile_id, rating_score, comment) VALUES 
(1, 2, 5, 'Отличный человек!'),
(2, 3, 4, 'Очень интересный профиль!'),
(3, 4, 3, 'Неплохой профиль'),
(4, 5, 2, 'Не совсем мой тип'),
(5, 1, 5, 'Прекрасный человек!');
