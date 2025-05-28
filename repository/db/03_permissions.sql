DO
$$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_user'
   ) THEN
      CREATE ROLE app_user WITH LOGIN PASSWORD 'your_secure_password';
   END IF;
END
$$;

GRANT SELECT ON complaint_types, notification_types, subscription_types, queries TO app_user;

GRANT SELECT ON admins TO app_user;

GRANT SELECT, INSERT, UPDATE, DELETE ON
    profiles,
    users,
    static,
    sessions,
    locations,
    interests,
    profile_interests,
    preferences,
    profile_preferences,
    likes,
    matches,
    subscriptions,
    complaints,
    blacklist,
    notifications,
    profile_ratings,
    user_answer,
    chats,
    messages
TO app_user;

GRANT USAGE ON SCHEMA public TO app_user;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

LOAD 'auto_explain';
