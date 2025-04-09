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

INSERT INTO static (path) VALUES 
('/default.png'),
('/eva.png'),
('/katya.png'),
('/default.png'),
('/default.png');

INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, photo_id, location_id) VALUES 
('Иван', 'Иванов', TRUE, '1990-05-15', 180, 'Программист', 1, 1),
('Анна', 'Петрова', FALSE, '1988-07-20', 165, 'Дизайнер', 1, 2),
('Ольга', 'Смирнова', FALSE, '1995-03-22', 170, 'Маркетолог', 1, 3),
('Дмитрий', 'Кузнецов', TRUE, '1992-11-30', 175, 'Инженер', 1, 4),
('Александр', 'Попов', TRUE, '1991-09-25', 185, 'Фотограф', 1, 5);

INSERT INTO users (profile_id, status, login, email, phone, password) VALUES
(1, 1, 'ivanivanov', 'ivanivanov@example.com', '1234567890', 'c789e1771c5bf1b6245ebb85b5d0ed883f1adedb49e86fe37c2c60c9ec0654ad'),
(2, 1, 'annapetrova', 'annapetrova@example.com', '1234567891', '6d1d8812d805a2c48a7f099279df40f156637cb3dca365781509c3489e0e16ac'),
(3, 1, 'olgasmirnova', 'olgasmirnova@example.com', '1234567892', '15f0b1cf493524a6b263064ebc906c8c1ef4353d4cd4a1c6b592698c5d0a352e'),
(4, 1, 'dmitrykuznetsov', 'dmitrykuznetsov@example.com', '1234567893', 'a7e75d733733063b93097c749dc90dc278e8035a240941073454feedfadd5512'),
(5, 1, 'alexanderpopov', 'alexanderpopov@example.com', '1234567894', '6547276a4a638f90725f7df92e8b438ed4b81954669cdee1f302134afa256cfc');

INSERT INTO sessions (user_id, token, expires_at) VALUES 
(1, 'token_ivanivanov_123', '2025-12-31 23:59:59'),
(2, 'token_annapetrova_123', '2025-12-31 23:59:59'),
(3, 'token_olgasmirnova_123', '2025-12-31 23:59:59'),
(4, 'token_dmitrykuznetsov_123', '2025-12-31 23:59:59'),
(5, 'token_alexanderpopov_123', '2025-12-31 23:59:59');

INSERT INTO interests (description) VALUES 
('Футбол'),
('Чтение'),
('Путешествия'),
('Музыка'),
('Кино');

INSERT INTO profile_interests (profile_id, interest_id) VALUES 
(1, 1), 
(2, 2),
(3, 3),
(4, 4), 
(5, 5); 


INSERT INTO preferences (preference_type, value) VALUES 
(1, 'Высокий рост'),
(2, 'Спортивная фигура'),
(3, 'Активные увлечения');

INSERT INTO profile_preferences (profile_id, preference_id) VALUES 
(1, 1),
(2, 2),
(3, 3),
(4, 1),
(5, 2);

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
(5, 1, 1);

INSERT INTO matches (profile_id, matched_profile_id) VALUES 
(1, 2),
(2, 3),
(3, 4),
(4, 5),
(5, 1);

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
