CREATE INDEX idx_profiles_profile_id ON profiles(profile_id);
CREATE INDEX idx_likes_by_liker ON likes(profile_id);
CREATE INDEX idx_likes_by_liked ON likes(liked_profile_id);

CREATE INDEX IF NOT EXISTS idx_static_profile_id ON "static"(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_interests_profile_id ON profile_interests(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_preferences_profile_id ON profile_preferences(profile_id);

CREATE INDEX IF NOT EXISTS idx_profile_interests_interest_id ON profile_interests(interest_id);
CREATE INDEX IF NOT EXISTS idx_profile_preferences_preference_id ON profile_preferences(preference_id);
CREATE INDEX IF NOT EXISTS idx_likes_liked_profile_id ON likes(liked_profile_id);
CREATE INDEX IF NOT EXISTS idx_profiles_location_id ON profiles(location_id);
