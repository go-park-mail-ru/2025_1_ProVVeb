INSERT INTO complaint_types (type_description) VALUES
    ('Неправомерное поведение'),
    ('Неуместный контент'),
    ('Нарушение условий пользования');

INSERT INTO notification_types (type_description) VALUES
    ('Сообщение от пользователя'),
    ('Предупреждение'),
    ('Оповещение о новой подписке');

INSERT INTO locations (country, city, district) VALUES
    ('Россия', 'Москва', 'Центр'),
    ('Россия', 'Санкт-Петербург', 'Невский район'),
    ('США', 'Нью-Йорк', 'Манхэттен');

INSERT INTO static (path) VALUES
    ('/avatars/default.jpg'),
    ('/avatars/default.jpg'),
    ('/avatars/default.jpg');

INSERT INTO subscription_types (type_description) VALUES
    ('Базовая подписка'),
    ('Премиум подписка'),
    ('VIP подписка');

INSERT INTO profiles (firstname, lastname, is_male, birthday, description, photo_id, location_id) VALUES
    ('Иван', 'Иванов', TRUE, '1985-05-15', 'Программист, увлекается технологиями', 1, 1),
    ('Алексей', 'Алексеев', TRUE, '1990-07-20', 'Разработчик, люблю путешествовать', 2, 1),
    ('Мария', 'Петрова', FALSE, '1992-10-10', 'Дизайнер, фанат искусства', 3, 2),
    ('Екатерина', 'Смирнова', FALSE, '1987-11-30', 'Маркетолог, активистка', 1, 2),
    ('Андрей', 'Козлов', TRUE, '1980-02-05', 'Предприниматель', 2, 3);

INSERT INTO users (profile_id, status, login, email, phone, password) VALUES
    (1, 1, 'ivan_85', 'ivan85@example.com', '+79161234567', 'password123'),
    (2, 1, 'alexey_90', 'alexey90@example.com', '+79261234567', 'password456'),
    (3, 1, 'maria_92', 'maria92@example.com', '+79361234567', 'password789'),
    (4, 1, 'ekaterina_87', 'ekaterina87@example.com', '+79461234567', 'password321'),
    (5, 1, 'andrey_80', 'andrey80@example.com', '+79561234567', 'password654');

INSERT INTO sessions (user_id, token, expires_at) VALUES
    (1, 'token_ivan_85', '2025-12-31 23:59:59'),
    (2, 'token_alexey_90', '2025-12-31 23:59:59'),
    (3, 'token_maria_92', '2025-12-31 23:59:59'),
    (4, 'token_ekaterina_87', '2025-12-31 23:59:59'),
    (5, 'token_andrey_80', '2025-12-31 23:59:59');

INSERT INTO interests (description) VALUES
    ('Программирование'),
    ('Путешествия'),
    ('Дизайн'),
    ('Маркетинг'),
    ('Предпринимательство');

INSERT INTO profile_interests (profile_id, interest_id) VALUES
    (1, 1),
    (2, 2),
    (3, 3),
    (4, 4),
    (5, 5);

INSERT INTO preferences (preference_type, value) VALUES
    (1, 'Темная тема'),
    (2, 'Получать уведомления'),
    (1, 'Светлая тема'),
    (2, 'Не получать уведомления'),
    (1, 'Темная тема');

INSERT INTO profile_preferences (profile_id, preference_id) VALUES
    (1, 1),
    (2, 2),
    (3, 3),
    (4, 4),
    (5, 5);

INSERT INTO messages (sender_profile_id, receiver_profile_id, content, status) VALUES
    (1, 2, 'Привет, Алексей!', 1),
    (2, 3, 'Здравствуй, Мария!', 1),
    (3, 4, 'Привет, Екатерина!', 1),
    (4, 5, 'Здравствуйте, Андрей!', 1),
    (5, 1, 'Привет, Иван!', 1);

INSERT INTO likes (profile_id, matched_profile_id, status) VALUES
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
    (1, 1, 'subscription_data_1'),
    (2, 2, 'subscription_data_2'),
    (3, 3, 'subscription_data_3'),
    (4, 1, 'subscription_data_4'),
    (5, 2, 'subscription_data_5');

INSERT INTO complaints (complaint_by, complaint_on, complaint_type, complaint_text, status) VALUES
    (1, 2, 1, 'Неправомерное поведение', 0),
    (2, 3, 2, 'Неуместный контент', 1),
    (3, 4, 3, 'Нарушение условий', 0),
    (4, 5, 1, 'Неправомерное поведение', 1),
    (5, 1, 2, 'Неуместный контент', 0);

INSERT INTO blacklist (user_id) VALUES
    (1),
    (2),
    (3),
    (4),
    (5);

INSERT INTO notifications (user_id, notification_type, content) VALUES
    (1, 1, 'Вы получили новое сообщение'),
    (2, 2, 'Ваш аккаунт получил предупреждение'),
    (3, 3, 'Вы подписались на премиум пакет'),
    (4, 1, 'Новый матч с пользователем'),
    (5, 2, 'Ваши данные были обновлены');

INSERT INTO profile_ratings (profile_id, rated_profile_id, rating_score, comment) VALUES
    (1, 2, 5, 'Отличный профиль!'),
    (2, 3, 4, 'Хорошие фотографии!'),
    (3, 4, 3, 'Нужно больше информации'),
    (4, 5, 5, 'Прекрасный профиль!'),
    (5, 1, 4, 'Очень интересный человек');
