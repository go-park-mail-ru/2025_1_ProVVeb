CREATE INDEX idx_profiles_profile_id ON profiles(profile_id);
CREATE INDEX idx_likes_by_liker ON likes(profile_id);
CREATE INDEX idx_likes_by_liked ON likes(liked_profile_id);