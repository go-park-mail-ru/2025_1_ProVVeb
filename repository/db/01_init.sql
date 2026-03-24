DO $$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_database WHERE datname = 'dev'
   ) THEN
      CREATE DATABASE dev;
   END IF;
END
$$;

ALTER DATABASE dev REFRESH COLLATION VERSION;

CREATE TABLE complaint_types (
    comp_type BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type_description TEXT NOT NULL CHECK (LENGTH(type_description) <= 255)
);

CREATE TABLE notification_types (
    notif_type BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type_description TEXT NOT NULL UNIQUE CHECK (LENGTH(type_description) <= 255)
);

CREATE TABLE locations (
    location_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    country TEXT NOT NULL CHECK (LENGTH(country) <= 255),
    city TEXT NOT NULL CHECK (LENGTH(city) <= 255),
    district TEXT NOT NULL CHECK (LENGTH(district) <= 255)
);


CREATE TABLE subscription_types (
    sub_type BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type_description TEXT NOT NULL CHECK (LENGTH(type_description) <= 255)
);

CREATE TABLE profiles (
    profile_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    firstname TEXT NOT NULL CHECK (LENGTH(firstname) <= 255),
    lastname TEXT NOT NULL CHECK (LENGTH(lastname) <= 255),
    fullname_translit TEXT CHECK (LENGTH(lastname) <= 255),
    is_male BOOLEAN NOT NULL,
    goal INT NOT NULL,
    birthday DATE NOT NULL,
    height INT CHECK (height >= 50 AND height <= 280),
    description TEXT,
    location_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(location_id) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE users (
    user_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    profile_id BIGINT,
    status INT NOT NULL,
    login TEXT UNIQUE NOT NULL CHECK (LENGTH(login) <= 255),
    email TEXT UNIQUE NOT NULL CHECK (LENGTH(email) <= 255)
    CHECK (email ~* '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
    phone TEXT UNIQUE CHECK (LENGTH(phone) <= 20),
    password TEXT NOT NULL CHECK (LENGTH(password) >= 8 AND LENGTH(password) <= 255), 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TABLE static (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    profile_id BIGINT NOT NULL,
    path TEXT NOT NULL CHECK (LENGTH(path) <= 255) DEFAULT '/default.png',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token TEXT NOT NULL CHECK (LENGTH(token) <= 255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE interests (
    interest_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profile_interests (
    profile_id BIGINT NOT NULL,
    interest_id BIGINT NOT NULL,
    PRIMARY KEY (profile_id, interest_id),
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (interest_id) REFERENCES interests(interest_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE preferences (
    preference_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    preference_type INT NOT NULL,
    preference_description TEXT NOT NULL CHECK (LENGTH(preference_description) <= 255),
    preference_value TEXT NOT NULL CHECK (LENGTH(preference_value) <= 255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profile_preferences (
    profile_id BIGINT NOT NULL,
    preference_id BIGINT NOT NULL,
    PRIMARY KEY (profile_id, preference_id),
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (preference_id) REFERENCES preferences(preference_id) ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE TABLE parameters (
    parameter_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parameter_type INT NOT NULL,
    parameter_description TEXT NOT NULL CHECK (LENGTH(parameter_description) <= 255),
    parameter_value TEXT NOT NULL CHECK (LENGTH(parameter_value) <= 255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profile_parameter (
    profile_id BIGINT NOT NULL,
    parameter_id BIGINT NOT NULL,
    PRIMARY KEY (profile_id, parameter_id),
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (parameter_id) REFERENCES parameters(parameter_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE likes (
    like_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    profile_id BIGINT NOT NULL,
    liked_profile_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status INT NOT NULL,

    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (liked_profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE (profile_id, liked_profile_id)
);

CREATE TABLE matches (
    match_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    profile_id BIGINT NOT NULL,
    matched_profile_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (matched_profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT unique_match UNIQUE (profile_id, matched_profile_id)
);

CREATE TABLE subscriptions (
    sub_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    sub_type BIGINT NOT NULL,
    transaction_data TEXT,
    border INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (sub_type) REFERENCES subscription_types(sub_type) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE complaints (
    complaint_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    complaint_by BIGINT NOT NULL,
    complaint_on BIGINT NOT NULL,
    complaint_type BIGINT NOT NULL,
    complaint_text TEXT NOT NULL,
    status INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP,
    FOREIGN KEY (complaint_by) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (complaint_on) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (complaint_type) REFERENCES complaint_types(comp_type) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE blacklist (
    block_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT unique_blacklist UNIQUE (user_id)
);

CREATE TABLE notifications (
    notification_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    notification_type BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (notification_type) REFERENCES notification_types(notif_type) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE profile_ratings (
    rating_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    profile_id BIGINT NOT NULL,
    rated_profile_id BIGINT NOT NULL,
    rating_score INT NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (rated_profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE TABLE queries (
    query_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL CHECK (LENGTH(name) <= 255),
    description TEXT NOT NULL,
    min_score INT NOT NULL,
    max_score INT NOT NULL,
    is_active BOOLEAN NOT NULL
);

CREATE TABLE user_answer (
    answer_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    query_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    score INT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (query_id) REFERENCES queries(query_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE chats (
    chat_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_profile_id BIGINT NOT NULL,
    second_profile_id BIGINT NOT NULL,
    last_message TEXT NOT NULL CHECK (LENGTH(last_message) <= 400),
    last_sender BIGINT NOT NULL,
    FOREIGN KEY (first_profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (second_profile_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (last_sender) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE (first_profile_id, second_profile_id)
);


CREATE TABLE messages (
    message_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    status INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES profiles(profile_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE admins (
    admin_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    role TEXT NOT NULL CHECK (LENGTH(role) <= 40),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE
);
